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
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/kakao/network-node-manager/pkg/configs"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Variables
var (
	configNodeName    string
	configIPv4Enabled bool
	configIPv6Enabled bool

	configRuleExternalCluster  bool
	configRuleDropInvalidInput bool

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
		configIPv4Enabled, configIPv6Enabled, err = configs.GetConfigNetStack()
		if err != nil {
			logger.Error(err, "config error")
			os.Exit(1)
		}
		configRuleExternalCluster, err = configs.GetConfigRuleExternalCluster()
		if err != nil {
			logger.Error(err, "config error")
			os.Exit(1)
		}
		configRuleDropInvalidInput, err = configs.GetConfigRuleDropInvalidInput()
		if err != nil {
			logger.Error(err, "config error")
			os.Exit(1)
		}

		logger.WithValues("node name", configNodeName).Info("config node name")
		logger.WithValues("IPv4", configIPv4Enabled).WithValues("IPv6", configIPv6Enabled).Info("config network stack")
		logger.WithValues("flag", configRuleExternalCluster).Info("config rule externalIP to clusterIP")
		logger.WithValues("flag", configRuleDropInvalidInput).Info("config rule drop invalid in input chain")

		// Run rules first
		if configRuleDropInvalidInput {
			if err := createRulesDropInvalidInput(logger); err != nil {
				logger.Error(err, "failed to create rule drop invalid input")
				os.Exit(1)
			}
		} else {
			if err := deleteRulesDropInvalidInput(logger); err != nil {
				logger.Error(err, "failed to delete rule drop invalid input")
				os.Exit(1)
			}
		}

		// Run rules periodically
		ticker := time.NewTicker(60 * time.Second)
		go func() {
			for {
				<-ticker.C
				if configRuleDropInvalidInput {
					if err := createRulesDropInvalidInput(logger); err != nil {
						logger.Error(err, "failed to cleanup rule drop invalid input")
					}
				}
			}
		}()

		if configRuleExternalCluster {
			// Get nodes's pod CIDR
			node := &corev1.Node{}
			if err := r.Client.Get(ctx, types.NamespacedName{Name: configNodeName}, node); err != nil {
				logger.Error(err, "failed to get the pod's node info from API server")
				os.Exit(1)
			}
			podCIDRIPv4, podCIDRIPv6 = getPodCIDR(node.Spec.PodCIDRs)
			logger.WithValues("pod CIDR IPV4", podCIDRIPv4).WithValues("pod CIDR IPv6", podCIDRIPv6).Info("pod CIDR")

			// Initialize externalIP to clusterIP rules
			if err := initRulesExternalCluster(logger); err != nil {
				logger.Error(err, "failed to initalize rule externalIP to clusterIP")
				os.Exit(1)
			}

			// Get all services
			svcs := &corev1.ServiceList{}
			if err := r.Client.List(ctx, svcs, client.InNamespace("")); err != nil {
				logger.Error(err, "failed to get all services from API server")
				os.Exit(1)
			}

			// Cleanup externalIP to clusterIP rules for deleted services
			if err := cleanupRulesExternalCluster(logger, svcs, podCIDRIPv4, podCIDRIPv6); err != nil {
				logger.Error(err, "failed to cleanup rule externalIP to clusterIP")
				os.Exit(1)
			}
		} else {
			// Destroy externalIP to clusterIP rules
			if err := destoryRulesExternalCluster(logger); err != nil {
				logger.Error(err, "failed to destroy rule externalIP to clusterIP")
				os.Exit(1)
			}
		}
	}

	// Loop
	if configRuleExternalCluster {
		// In case the iptables chain is deleted, initalize again
		if err := initRulesExternalCluster(logger); err != nil {
			logger.Error(err, "failed to initalize rule externalIP to clusterIP")
			os.Exit(1)
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

				// Delete rules
				for _, oldIngress := range oldSvc.Status.LoadBalancer.Ingress {
					oldClusterIP := oldSvc.Spec.ClusterIP
					oldExternalIP := oldIngress.IP

					// Delete rules
					logger.WithValues("externalIP", oldExternalIP).WithValues("clusterIP", oldClusterIP).Info("delete rule externalIp to clusterIP")
					if err := deleteRulesExternalCluster(logger, &req, oldClusterIP, oldExternalIP, podCIDRIPv4, podCIDRIPv6); err != nil {
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

		// Create rules
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			clusterIP := svc.Spec.ClusterIP
			externalIP := ingress.IP

			// Cache service to use deleting service
			serviceCache[req] = *svc.DeepCopy()

			// Create rules
			logger.WithValues("externalIP", externalIP).WithValues("clusterIP", clusterIP).Info("create iptables rules")
			if err := createRulesExternalCluster(logger, &req, clusterIP, externalIP, podCIDRIPv4, podCIDRIPv6); err != nil {
				return ctrl.Result{}, err
			}
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
