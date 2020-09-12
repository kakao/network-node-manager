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
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"github.com/kakao/ipvs-node-controller/pkg/ip"
	"github.com/kakao/ipvs-node-controller/pkg/iptables"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Constants
const (
	EnvNetStack        = "NET_STACK"
	EnvNetStackIPv4    = "ipv4"
	EnvNetStackIPv6    = "ipv6"
	EnvNetStackDefault = EnvNetStackIPv4
	EnvNodeName        = "NODE_NAME"

	ChainNATPrerouting     = "PREROUTING"
	ChainNATOutput         = "OUTPUT"
	ChainNATKubeMasquerade = "KUBE-MARK-MASQ"
	ChainNATIPVSPrerouting = "IPVS_PREROUTING"
	ChainNATIPVSOutput     = "IPVS_OUTPUT"
)

// Variables
var (
	configIPv4Enabled bool
	configIPv6Enabled bool
	configNodeName    string

	serviceCache = map[reconcile.Request]corev1.Service{}
)

// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch

func (r *ServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("service", req.NamespacedName)

	// Get Nodes's pod CIDR
	node := &corev1.Node{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: configNodeName}, node); err != nil {
		logger.Error(err, "failed to get the pod's node info from API server")
		return ctrl.Result{}, err
	}
	podCIDRs := node.Spec.PodCIDRs
	podCIDRIPv4, podCIDRIPv6 := getPodCIDR(podCIDRs)
	logger.WithValues("pod CIDR IPV4", podCIDRIPv4).WithValues("pod CIDR IPv6", podCIDRIPv6).Info("pod CIDR")

	// Get service info
	svc := &corev1.Service{}
	err := r.Client.Get(ctx, req.NamespacedName, svc)
	if err != nil {
		if !apierror.IsNotFound(err) {
			logger.Error(err, "failed to get service info")
			return ctrl.Result{}, err
		} else {
			// Not found service means that the service is removed
			// Get svc from cache
			oldSvc, exist := serviceCache[req]
			if !exist {
				logger.Error(err, "failed to get removed service info from cache")
				return reconcile.Result{}, nil
			}

			// Check removed service is LoadBalancer type
			if oldSvc.Spec.Type != corev1.ServiceTypeLoadBalancer {
				return ctrl.Result{}, nil
			}

			// Remove rules
			for _, ingress := range oldSvc.Status.LoadBalancer.Ingress {
				oldClusterIP := oldSvc.Spec.ClusterIP
				oldExternalIP := ingress.IP
				logger.WithValues("externalIP", oldExternalIP).Info("remove rules")

				// Don't use spec.ipFamily to distingush between IPv4 and IPv6 Address
				// for kubernetes version that dosen't support IPv6 dualstack
				if configIPv4Enabled && ip.IsIPv4Addr(oldExternalIP) {
					// IPv4
					// Unset prerouting
					rulePreMasq := []string{"-s", podCIDRIPv4, "-d", oldExternalIP, "-j", ChainNATKubeMasquerade}
					out, err := iptables.DeleteRuleIPv4(iptables.TableNAT, ChainNATIPVSPrerouting, req.String(), rulePreMasq...)
					if err != nil {
						logger.Error(err, out)
						return ctrl.Result{}, err
					}
					rulePreDNAT := []string{"-s", podCIDRIPv4, "-d", oldExternalIP, "-j", "DNAT", "--to-destination", oldClusterIP}
					out, err = iptables.DeleteRuleIPv4(iptables.TableNAT, ChainNATIPVSPrerouting, req.String(), rulePreDNAT...)
					if err != nil {
						logger.Error(err, out)
						return ctrl.Result{}, err
					}

					// Unset output
					ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", oldExternalIP, "-j", ChainNATKubeMasquerade}
					out, err = iptables.DeleteRuleIPv4(iptables.TableNAT, ChainNATIPVSOutput, req.String(), ruleOutMasq...)
					if err != nil {
						logger.Error(err, out)
						return ctrl.Result{}, err
					}
					ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", oldExternalIP, "-j", "DNAT", "--to-destination", oldClusterIP}
					out, err = iptables.DeleteRuleIPv4(iptables.TableNAT, ChainNATIPVSOutput, req.String(), ruleOutDNAT...)
					if err != nil {
						logger.Error(err, out)
						return ctrl.Result{}, err
					}
				} else if configIPv6Enabled && ip.IsIPv6Addr(oldExternalIP) {
					// IPv6
					// Unset prerouting
					rulePreMasq := []string{"-s", podCIDRIPv6, "-d", oldExternalIP, "-j", ChainNATKubeMasquerade}
					out, err := iptables.DeleteRuleIPv6(iptables.TableNAT, ChainNATIPVSPrerouting, req.String(), rulePreMasq...)
					if err != nil {
						logger.Error(err, out)
						return ctrl.Result{}, err
					}
					rulePreDNAT := []string{"-s", podCIDRIPv6, "-d", oldExternalIP, "-j", "DNAT", "--to-destination", oldClusterIP}
					out, err = iptables.DeleteRuleIPv6(iptables.TableNAT, ChainNATIPVSPrerouting, req.String(), rulePreDNAT...)
					if err != nil {
						logger.Error(err, out)
						return ctrl.Result{}, err
					}

					// Unset output
					ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", oldExternalIP, "-j", ChainNATKubeMasquerade}
					out, err = iptables.DeleteRuleIPv6(iptables.TableNAT, ChainNATIPVSOutput, req.String(), ruleOutMasq...)
					if err != nil {
						logger.Error(err, out)
						return ctrl.Result{}, err
					}
					ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", oldExternalIP, "-j", "DNAT", "--to-destination", oldClusterIP}
					out, err = iptables.DeleteRuleIPv6(iptables.TableNAT, ChainNATIPVSOutput, req.String(), ruleOutDNAT...)
					if err != nil {
						logger.Error(err, out)
						return ctrl.Result{}, err
					}
				} else {
					if ip.IsVaildIP(oldExternalIP) {
						logger.WithValues("externalIP", oldExternalIP).Error(errors.New("invalid IP"), "invaild IP")
					}
					continue
				}
			}

			return ctrl.Result{}, nil
		}
	}

	// Check service is LoadBalancer type
	if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return ctrl.Result{}, nil
	}

	// Cache service to use when deleting service
	serviceCache[req] = *svc.DeepCopy()

	// Create rules
	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		clusterIP := svc.Spec.ClusterIP
		externalIP := ingress.IP
		logger.WithValues("externalIP", externalIP).Info("create rules")

		// Don't use spec.ipFamily to distingush between IPv4 and IPv6 Address
		// for kubernetes version that dosen't support IPv6 dualstack
		if configIPv4Enabled && ip.IsIPv4Addr(externalIP) {
			// IPv4
			// Set prerouting
			rulePreMasq := []string{"-s", podCIDRIPv4, "-d", externalIP, "-j", ChainNATKubeMasquerade}
			out, err := iptables.CreateRuleLastIPv4(iptables.TableNAT, ChainNATIPVSPrerouting, req.String(), rulePreMasq...)
			if err != nil {
				logger.Error(err, out)
				return ctrl.Result{}, err
			}
			rulePreDNAT := []string{"-s", podCIDRIPv4, "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
			out, err = iptables.CreateRuleLastIPv4(iptables.TableNAT, ChainNATIPVSPrerouting, req.String(), rulePreDNAT...)
			if err != nil {
				logger.Error(err, out)
				return ctrl.Result{}, err
			}

			// Set output
			ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", ChainNATKubeMasquerade}
			out, err = iptables.CreateRuleLastIPv4(iptables.TableNAT, ChainNATIPVSOutput, req.String(), ruleOutMasq...)
			if err != nil {
				logger.Error(err, out)
				return ctrl.Result{}, err
			}
			ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
			out, err = iptables.CreateRuleLastIPv4(iptables.TableNAT, ChainNATIPVSOutput, req.String(), ruleOutDNAT...)
			if err != nil {
				logger.Error(err, out)
				return ctrl.Result{}, err
			}
		} else if configIPv6Enabled && ip.IsIPv6Addr(externalIP) {
			// IPv6
			// Set prerouting
			rulePreMasq := []string{"-s", podCIDRIPv6, "-d", externalIP, "-j", ChainNATKubeMasquerade}
			out, err := iptables.CreateRuleLastIPv6(iptables.TableNAT, ChainNATIPVSPrerouting, req.String(), rulePreMasq...)
			if err != nil {
				logger.Error(err, out)
				return ctrl.Result{}, err
			}
			rulePreDNAT := []string{"-s", podCIDRIPv6, "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
			out, err = iptables.CreateRuleLastIPv6(iptables.TableNAT, ChainNATIPVSPrerouting, req.String(), rulePreDNAT...)
			if err != nil {
				logger.Error(err, out)
				return ctrl.Result{}, err
			}

			// Set output
			ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", ChainNATKubeMasquerade}
			out, err = iptables.CreateRuleLastIPv6(iptables.TableNAT, ChainNATIPVSOutput, req.String(), ruleOutMasq...)
			if err != nil {
				logger.Error(err, out)
				return ctrl.Result{}, err
			}
			ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
			out, err = iptables.CreateRuleLastIPv6(iptables.TableNAT, ChainNATIPVSOutput, req.String(), ruleOutDNAT...)
			if err != nil {
				logger.Error(err, out)
				return ctrl.Result{}, err
			}
		} else {
			if ip.IsVaildIP(externalIP) {
				logger.WithValues("externalIP", externalIP).Error(errors.New("invalid IP"), "invaild IP")
			}
			continue
		}
	}

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logger := r.Log.WithName("initialize")
	errConfig := errors.New("config error")

	// Get network stack config from env
	configNetStack := os.Getenv(EnvNetStack)
	if configNetStack == "" {
		configIPv4Enabled, configIPv6Enabled = getConfigNetStack(EnvNetStackDefault)
	} else {
		configIPv4Enabled, configIPv6Enabled = getConfigNetStack(configNetStack)
	}
	logger.WithValues("IPv4", configIPv4Enabled).WithValues("IPv6", configIPv6Enabled).Info("config network stack")

	// Get node name config from env
	configNodeName = os.Getenv(EnvNodeName)
	if configNodeName == "" {
		logger.Error(errConfig, fmt.Sprintf("failed to get the pod's node name from %s env", EnvNodeName))
		return errConfig
	}
	logger.WithValues("node name", configNodeName).Info("config node name")

	// IPv4
	if configIPv4Enabled {
		// Create chain in nat table
		logger.Info("create the IPVS IPv4 chains")
		out, err := iptables.CreateChainIPv4(iptables.TableNAT, ChainNATIPVSPrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv4(iptables.TableNAT, ChainNATIPVSOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set jump rule to each chain in nat table
		logger.Info("create jump rules for the IPVS IPv4 chains")
		ruleJumpPre := []string{"-j", ChainNATIPVSPrerouting}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableNAT, ChainNATPrerouting, "", ruleJumpPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpOut := []string{"-j", ChainNATIPVSOutput}
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
		out, err := iptables.CreateChainIPv6(iptables.TableNAT, ChainNATIPVSPrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv6(iptables.TableNAT, ChainNATIPVSOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set jump rule to each chain in nat table
		logger.Info("create jump rules for the IPVS IPv6 chains")
		ruleJumpPre := []string{"-j", ChainNATIPVSPrerouting}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableNAT, ChainNATPrerouting, "", ruleJumpPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpOut := []string{"-j", ChainNATIPVSOutput}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableNAT, ChainNATOutput, "", ruleJumpOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	// Set controller manager
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
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

func getConfigNetStack(configs string) (ipv4 bool, ipv6 bool) {
	for _, config := range strings.Split(configs, ",") {
		if config == EnvNetStackIPv4 {
			ipv4 = true
		} else if config == EnvNetStackIPv6 {
			ipv6 = true
		}
	}
	return
}
