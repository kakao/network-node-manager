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

	corev1 "k8s.io/api/core/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
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
	EnvPodNodeName = "POD_NODE_NAME"

	ChainNATPrerouting     = "PREROUTING"
	ChainNATOutput         = "OUTPUT"
	ChainNATKubeMasquerade = "KUBE-MARK-MASQ"
	ChainNATIPVSPrerouting = "IPVS_PREROUTING"
	ChainNATIPVSOutput     = "IPVS_OUTPUT"
)

// Variables
var (
	serviceCache = map[reconcile.Request]corev1.Service{}
)

// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch

func (r *ServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("service", req.NamespacedName)

	// Get podCIDR for the node
	nodeName := os.Getenv(EnvPodNodeName)
	if nodeName == "" {
		err := errors.New("invalid config")
		logger.Error(err, fmt.Sprintf("failed to get the pod's node name from %s env", EnvPodNodeName))
		return ctrl.Result{}, err
	}
	node := &corev1.Node{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: nodeName}, node); err != nil {
		logger.Error(err, "failed to get the pod's node info from API server")
		return ctrl.Result{}, err
	}
	podCIDR := node.Spec.PodCIDR

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

				// Unset prerouting
				rulePreMasq := []string{"-s", podCIDR, "-d", oldExternalIP, "-j", ChainNATKubeMasquerade}
				out, err := iptables.DeleteRule(iptables.TableNAT, ChainNATIPVSPrerouting, req.String(), rulePreMasq...)
				if err != nil {
					logger.Error(err, out)
					return ctrl.Result{}, err
				}
				rulePreDNAT := []string{"-s", podCIDR, "-d", oldExternalIP, "-j", "DNAT", "--to-destination", oldClusterIP}
				out, err = iptables.DeleteRule(iptables.TableNAT, ChainNATIPVSPrerouting, req.String(), rulePreDNAT...)
				if err != nil {
					logger.Error(err, out)
					return ctrl.Result{}, err
				}

				// Unset output
				ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", oldExternalIP, "-j", ChainNATKubeMasquerade}
				out, err = iptables.DeleteRule(iptables.TableNAT, ChainNATIPVSOutput, req.String(), ruleOutMasq...)
				if err != nil {
					logger.Error(err, out)
					return ctrl.Result{}, err
				}
				ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", oldExternalIP, "-j", "DNAT", "--to-destination", oldClusterIP}
				out, err = iptables.DeleteRule(iptables.TableNAT, ChainNATIPVSOutput, req.String(), ruleOutDNAT...)
				if err != nil {
					logger.Error(err, out)
					return ctrl.Result{}, err
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

		// Set prerouting
		rulePreMasq := []string{"-s", podCIDR, "-d", externalIP, "-j", ChainNATKubeMasquerade}
		out, err := iptables.CreateRuleLast(iptables.TableNAT, ChainNATIPVSPrerouting, req.String(), rulePreMasq...)
		if err != nil {
			logger.Error(err, out)
			return ctrl.Result{}, err
		}
		rulePreDNAT := []string{"-s", podCIDR, "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.CreateRuleLast(iptables.TableNAT, ChainNATIPVSPrerouting, req.String(), rulePreDNAT...)
		if err != nil {
			logger.Error(err, out)
			return ctrl.Result{}, err
		}

		// Set output
		ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", ChainNATKubeMasquerade}
		out, err = iptables.CreateRuleLast(iptables.TableNAT, ChainNATIPVSOutput, req.String(), ruleOutMasq...)
		if err != nil {
			logger.Error(err, out)
			return ctrl.Result{}, err
		}
		ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.CreateRuleLast(iptables.TableNAT, ChainNATIPVSOutput, req.String(), ruleOutDNAT...)
		if err != nil {
			logger.Error(err, out)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logger := r.Log.WithName("initialize")

	// Create chain in nat table
	logger.Info("create the IPVS chains")
	out, err := iptables.CreateChain(iptables.TableNAT, ChainNATIPVSPrerouting)
	if err != nil {
		logger.Error(err, out)
		return err
	}
	out, err = iptables.CreateChain(iptables.TableNAT, ChainNATIPVSOutput)
	if err != nil {
		logger.Error(err, out)
		return err
	}

	// Set jump rule to each chain in nat table
	logger.Info("create jump rules for the IPVS chains")
	ruleJumpPre := []string{"-j", ChainNATIPVSPrerouting}
	out, err = iptables.CreateRuleFirst(iptables.TableNAT, ChainNATPrerouting, "", ruleJumpPre...)
	if err != nil {
		logger.Error(err, out)
		return err
	}
	ruleJumpOut := []string{"-j", ChainNATIPVSOutput}
	out, err = iptables.CreateRuleFirst(iptables.TableNAT, ChainNATOutput, "", ruleJumpOut...)
	if err != nil {
		logger.Error(err, out)
		return err
	}

	// Set controller manager
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}
