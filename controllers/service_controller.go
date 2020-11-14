/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/kakao/network-node-manager/pkg/configs"
	"github.com/kakao/network-node-manager/pkg/ip"
	"github.com/kakao/network-node-manager/pkg/iptables"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Constants
const (
	ChainNATPrerouting     = "PREROUTING"
	ChainNATOutput         = "OUTPUT"
	ChainNATKubeMasquerade = "KUBE-MARK-MASQ"

	ChainNATIPVSLBPrerouting = "NMANAGER_IPVS_LB_PREROUTING"
	ChainNATIPVSLBOutput     = "NMANAGER_IPVS_LB_OUTPUT"
)

// Variables
var (
	configNodeName    string
	configIPv4Enabled bool
	configIPv6Enabled bool

	initFlag    = false
	podCIDRIPv4 string
	podCIDRIPv6 string

	serviceCache = map[ctrl.Request]corev1.Service{}
)

// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch

func (r *ServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithName("reconcile").WithValues("service", req.NamespacedName)
	var err error

	// Init service controller
	// In SetupWithManager, function k8s client cannot be used.
	// So initialize controller in Reconile() function.
	if !initFlag {
		defer func() {
			initFlag = true
		}()

		// Init logger for only initialize controller
		logger := r.Log.WithName("initalize")
		logger.Info("initalize service contoller")

		// Get configs from env
		configNodeName, err = configs.GetConfigNodeName()
		if err != nil {
			logger.Error(err, "config error")
			os.Exit(1)
		}
		logger.WithValues("node name", configNodeName).Info("config node name")
		configIPv4Enabled, configIPv6Enabled, err = configs.GetConfigNetStack()
		if err != nil {
			logger.Error(err, "config error")
			os.Exit(1)
		}
		logger.WithValues("IPv4", configIPv4Enabled).WithValues("IPv6", configIPv6Enabled).Info("config network stack")

		// Get nodes's pod CIDR
		node := &corev1.Node{}
		if err := r.Client.Get(ctx, types.NamespacedName{Name: configNodeName}, node); err != nil {
			logger.Error(err, "failed to get the pod's node info from API server")
			os.Exit(1)
		}
		podCIDRIPv4, podCIDRIPv6 = getPodCIDR(node.Spec.PodCIDRs)
		logger.WithValues("pod CIDR IPV4", podCIDRIPv4).WithValues("pod CIDR IPv6", podCIDRIPv6).Info("pod CIDR")

		// Init iptables
		if err := initIptables(logger); err != nil {
			logger.Error(err, "failed to init iptables")
			os.Exit(1)
		}

		// Get all services
		svcs := &corev1.ServiceList{}
		if err := r.Client.List(ctx, svcs, client.InNamespace("")); err != nil {
			logger.Error(err, "failed to get all services from API server")
			os.Exit(1)
		}

		// Cleanup iptables for deleted services
		if err := cleanupIptables(logger, svcs, podCIDRIPv4, podCIDRIPv6); err != nil {
			logger.Error(err, "failed to cleanup iptables")
			os.Exit(1)
		}
	}

	// Get service info
	svc := &corev1.Service{}
	if err := r.Client.Get(ctx, req.NamespacedName, svc); err != nil {
		if apierror.IsNotFound(err) {
			// Not found service means that the service is removed.
			// Delete iptables rules by using cache

			// Get service from cache
			oldSvc, exist := serviceCache[req]
			if !exist {
				// If there is no service info in cache, skip it
				return ctrl.Result{}, nil
			}

			// Delete iptables rules
			for _, oldIngress := range oldSvc.Status.LoadBalancer.Ingress {
				oldClusterIP := oldSvc.Spec.ClusterIP
				oldExternalIP := oldIngress.IP

				// Delete iptables rules
				logger.WithValues("externalIP", oldExternalIP).Info("delete iptables rules")
				if err := deleteIptablesRules(logger, &req, oldClusterIP, oldExternalIP, podCIDRIPv4, podCIDRIPv6); err != nil {
					return ctrl.Result{}, err
				}
			}
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "failed to get service info")
			return ctrl.Result{}, err
		}
	}

	// Check service is LoadBalancer type
	if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return ctrl.Result{}, nil
	}

	// Create iptables rules
	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		clusterIP := svc.Spec.ClusterIP
		externalIP := ingress.IP

		// Cache service to use when deleting service
		serviceCache[req] = *svc.DeepCopy()

		// Create iptables rules
		logger.WithValues("externalIP", externalIP).Info("create iptables rules")
		if err := createIptablesRules(logger, &req, clusterIP, externalIP, podCIDRIPv4, podCIDRIPv6); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Set controller manager
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}

func initIptables(logger logr.Logger) error {
	logger.Info("create iptables chains")

	// IPv4
	if configIPv4Enabled {
		// Create chain in nat table
		logger.Info("create the IPVS IPv4 chains")
		out, err := iptables.CreateChainIPv4(iptables.TableNAT, ChainNATIPVSLBPrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv4(iptables.TableNAT, ChainNATIPVSLBOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set jump rule to each chain in nat table
		logger.Info("create jump rules for the IPVS IPv4 chains")
		ruleJumpPre := []string{"-j", ChainNATIPVSLBPrerouting}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableNAT, ChainNATPrerouting, "", ruleJumpPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpOut := []string{"-j", ChainNATIPVSLBOutput}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableNAT, ChainNATOutput, "", ruleJumpOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}
	// IPv6
	if configIPv6Enabled {
		// Create chain in nat table
		logger.Info("create the IPVS IPv6 chains")
		out, err := iptables.CreateChainIPv6(iptables.TableNAT, ChainNATIPVSLBPrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv6(iptables.TableNAT, ChainNATIPVSLBOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set jump rule to each chain in nat table
		logger.Info("create jump rules for the IPVS IPv6 chains")
		ruleJumpPre := []string{"-j", ChainNATIPVSLBPrerouting}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableNAT, ChainNATPrerouting, "", ruleJumpPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpOut := []string{"-j", ChainNATIPVSLBOutput}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableNAT, ChainNATOutput, "", ruleJumpOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	return nil
}

func cleanupIptables(logger logr.Logger, svcs *corev1.ServiceList, podCIDRIPv4, podCIDRIPv6 string) error {
	logger.Info("cleanup iptables for deleted services")

	// IPv4
	if configIPv4Enabled {
		// Make up service map
		svcMap := make(map[string]*corev1.Service)
		for _, svc := range svcs.Items {
			if ip.IsIPv4Addr(svc.Spec.ClusterIP) && svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
				svcMap[svc.Namespace+"/"+svc.Name] = svc.DeepCopy()
			}
		}

		// Cleanup prerouting chain
		preRules, err := iptables.GetRulesIPv4(iptables.TableNAT, ChainNATIPVSLBPrerouting)
		if err != nil {
			return err
		}
		for _, rule := range preRules {
			// Get service info from rule and k8s, and delete iptables rules
			nsName, src, dest, jump, dnatDest := getSvcInfoFromRule(rule)
			svc, ok := svcMap[nsName]
			if !ok {
				logger.WithValues("rule", rule).Info("there is no service info in k8s. cleanup prerouting chain IPv4 rule")
				out, err := iptables.DeleteRuleRawIPv4(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				if err != nil {
					logger.Error(err, out)
				}
				continue
			}

			// Compare service info and delete iptables rules
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				externalIP := ingress.IP + "/32"
				if (jump == ChainNATKubeMasquerade && (src == podCIDRIPv4 && dest == externalIP)) ||
					(jump == "DNAT" && (src == podCIDRIPv4 && dest == externalIP && dnatDest == svc.Spec.ClusterIP)) {
					continue
				}
				logger.WithValues("rule", rule).Info("service info is diff. cleanup prerouting chain IPv4 rule")
				out, err := iptables.DeleteRuleRawIPv4(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				if err != nil {
					logger.Error(err, out)
				}
			}
		}

		// Cleanup output chain
		outRules, err := iptables.GetRulesIPv4(iptables.TableNAT, ChainNATIPVSLBOutput)
		if err != nil {
			return err
		}
		for _, rule := range outRules {
			// Get service info from rule and k8s, and delete iptables rules
			nsName, _, dest, jump, dnatDest := getSvcInfoFromRule(rule)
			svc, ok := svcMap[nsName]
			if !ok {
				logger.WithValues("rule", rule).Info("there is no service info in k8s. cleanup output chain IPv4 rule")
				iptables.DeleteRuleRawIPv4(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				continue
			}

			// Compare service info and delete diff iptables rules
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				externalIP := ingress.IP + "/32"
				if (jump == ChainNATKubeMasquerade && dest == externalIP) ||
					(jump == "DNAT" && (dest == externalIP && dnatDest == svc.Spec.ClusterIP)) {
					continue
				}
				logger.WithValues("rule", rule).Info("service info is diff. cleanup output chain IPv4 rule")
				out, err := iptables.DeleteRuleRawIPv4(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				if err != nil {
					logger.Error(err, out)
				}
			}
		}
	}
	// IPv6
	if configIPv6Enabled {
		// Make up service map
		svcMap := make(map[string]*corev1.Service)
		for _, svc := range svcs.Items {
			if ip.IsIPv6Addr(svc.Spec.ClusterIP) && svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
				svcMap[svc.Namespace+"/"+svc.Name] = svc.DeepCopy()
			}
		}

		// Cleanup prerouting chain
		preRules, err := iptables.GetRulesIPv6(iptables.TableNAT, ChainNATIPVSLBPrerouting)
		if err != nil {
			return err
		}
		for _, rule := range preRules {
			// Get service info from rule and k8s, and delete iptables rules
			nsName, src, dest, jump, dnatDest := getSvcInfoFromRule(rule)
			svc, ok := svcMap[nsName]
			if !ok {
				logger.WithValues("rule", rule).Info("there is no service info in k8s. cleanup prerouting chain IPv6 rule")
				out, err := iptables.DeleteRuleRawIPv6(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				if err != nil {
					logger.Error(err, out)
				}
				continue
			}

			// Compare service info and delete iptables rules
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				externalIP := ingress.IP + "/128"
				if (jump == ChainNATKubeMasquerade && (src == podCIDRIPv6 && dest == externalIP)) ||
					(jump == "DNAT" && (src == podCIDRIPv6 && dest == externalIP && dnatDest == svc.Spec.ClusterIP)) {
					continue
				}
				logger.WithValues("rule", rule).Info("service info is diff. cleanup prerouting chain IPv6 rule")
				out, err := iptables.DeleteRuleRawIPv6(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				if err != nil {
					logger.Error(err, out)
				}
			}
		}

		// Cleanup output
		outRules, err := iptables.GetRulesIPv6(iptables.TableNAT, ChainNATIPVSLBOutput)
		if err != nil {
			return err
		}
		for _, rule := range outRules {
			// Get service info from rule and k8s, and delete iptables rules
			nsName, _, dest, jump, dnatDest := getSvcInfoFromRule(rule)
			svc, ok := svcMap[nsName]
			if !ok {
				logger.WithValues("rule", rule).Info("there is no service info in k8s. cleanup output chain IPv6 rule")
				iptables.DeleteRuleRawIPv6(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				continue
			}

			// Compare service info and delete diff iptables rules
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				externalIP := ingress.IP + "/128"
				if (jump == ChainNATKubeMasquerade && dest == externalIP) ||
					(jump == "DNAT" && (dest == externalIP && dnatDest == svc.Spec.ClusterIP)) {
					continue
				}
				logger.WithValues("rule", rule).Info("service info is diff. cleanup output chain IPv6 rule")
				iptables.DeleteRuleRawIPv6(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
			}
		}
	}

	return nil
}

func createIptablesRules(logger logr.Logger, req *ctrl.Request, clusterIP, externalIP, podCIDRIPv4, podCIDRIPv6 string) error {
	// Don't use spec.ipFamily to distingush between IPv4 and IPv6 Address
	// for kubernetes version that dosen't support IPv6 dualstack
	if configIPv4Enabled && ip.IsIPv4Addr(clusterIP) {
		// IPv4
		// Set prerouting
		rulePreMasq := []string{"-s", podCIDRIPv4, "-d", externalIP, "-j", ChainNATKubeMasquerade}
		out, err := iptables.CreateRuleLastIPv4(iptables.TableNAT, ChainNATIPVSLBPrerouting, req.String(), rulePreMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		rulePreDNAT := []string{"-s", podCIDRIPv4, "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.CreateRuleLastIPv4(iptables.TableNAT, ChainNATIPVSLBPrerouting, req.String(), rulePreDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set output
		ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", ChainNATKubeMasquerade}
		out, err = iptables.CreateRuleLastIPv4(iptables.TableNAT, ChainNATIPVSLBOutput, req.String(), ruleOutMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.CreateRuleLastIPv4(iptables.TableNAT, ChainNATIPVSLBOutput, req.String(), ruleOutDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	} else if configIPv6Enabled && ip.IsIPv6Addr(clusterIP) {
		// IPv6
		// Set prerouting
		rulePreMasq := []string{"-s", podCIDRIPv6, "-d", externalIP, "-j", ChainNATKubeMasquerade}
		out, err := iptables.CreateRuleLastIPv6(iptables.TableNAT, ChainNATIPVSLBPrerouting, req.String(), rulePreMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		rulePreDNAT := []string{"-s", podCIDRIPv6, "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.CreateRuleLastIPv6(iptables.TableNAT, ChainNATIPVSLBPrerouting, req.String(), rulePreDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set output
		ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", ChainNATKubeMasquerade}
		out, err = iptables.CreateRuleLastIPv6(iptables.TableNAT, ChainNATIPVSLBOutput, req.String(), ruleOutMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.CreateRuleLastIPv6(iptables.TableNAT, ChainNATIPVSLBOutput, req.String(), ruleOutDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	} else {
		if ip.IsVaildIP(clusterIP) {
			logger.WithValues("clusterIP", clusterIP).Error(errors.New("invalid IP"), "invaild IP")
		}
	}

	return nil
}

func deleteIptablesRules(logger logr.Logger, req *ctrl.Request, clusterIP, externalIP, podCIDRIPv4, podCIDRIPv6 string) error {
	// Don't use spec.ipFamily to distingush between IPv4 and IPv6 Address
	// for kubernetes version that dosen't support IPv6 dualstack
	if configIPv4Enabled && ip.IsIPv4Addr(clusterIP) {
		// IPv4
		// Unset prerouting
		rulePreMasq := []string{"-s", podCIDRIPv4, "-d", externalIP, "-j", ChainNATKubeMasquerade}
		out, err := iptables.DeleteRuleIPv4(iptables.TableNAT, ChainNATIPVSLBPrerouting, req.String(), rulePreMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		rulePreDNAT := []string{"-s", podCIDRIPv4, "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.DeleteRuleIPv4(iptables.TableNAT, ChainNATIPVSLBPrerouting, req.String(), rulePreDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Unset output
		ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", ChainNATKubeMasquerade}
		out, err = iptables.DeleteRuleIPv4(iptables.TableNAT, ChainNATIPVSLBOutput, req.String(), ruleOutMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.DeleteRuleIPv4(iptables.TableNAT, ChainNATIPVSLBOutput, req.String(), ruleOutDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	} else if configIPv6Enabled && ip.IsIPv6Addr(clusterIP) {
		// IPv6
		// Unset prerouting
		rulePreMasq := []string{"-s", podCIDRIPv6, "-d", externalIP, "-j", ChainNATKubeMasquerade}
		out, err := iptables.DeleteRuleIPv6(iptables.TableNAT, ChainNATIPVSLBPrerouting, req.String(), rulePreMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		rulePreDNAT := []string{"-s", podCIDRIPv6, "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.DeleteRuleIPv6(iptables.TableNAT, ChainNATIPVSLBPrerouting, req.String(), rulePreDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Unset output
		ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", ChainNATKubeMasquerade}
		out, err = iptables.DeleteRuleIPv6(iptables.TableNAT, ChainNATIPVSLBOutput, req.String(), ruleOutMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.DeleteRuleIPv6(iptables.TableNAT, ChainNATIPVSLBOutput, req.String(), ruleOutDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	} else {
		if ip.IsVaildIP(clusterIP) {
			logger.WithValues("clusterIP", clusterIP).Error(errors.New("invalid IP"), "invaild IP")
		}
	}
	return nil
}

func getPodCIDR(cidrs []string) (ipv4CIDR string, ipv6CIDR string) {
	for _, cidr := range cidrs {
		addr := strings.Split(cidr, "/")[0]
		if ip.IsIPv4Addr(addr) {
			ipv4CIDR = cidr
		} else if ip.IsIPv6Addr(addr) {
			ipv6CIDR = cidr
		}
	}
	return
}

func getSvcInfoFromRule(rule string) (nsName, src, dest, jump, dnatDest string) {
	nsName = iptables.GetRuleComment(rule)
	src = iptables.GetRuleSrc(rule)
	dest = iptables.GetRuleDest(rule)
	jump = iptables.GetRuleJump(rule)
	dnatDest = iptables.GetRuleDNATDest(rule)
	return
}
