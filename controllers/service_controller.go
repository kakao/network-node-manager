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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/kakao/network-node-manager/pkg/configs"
	"github.com/kakao/network-node-manager/pkg/ip"
	"github.com/kakao/network-node-manager/pkg/rules"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Variables
var (
	configPodCIDRIPv4 string
	configPodCIDRIPv6 string

	configRuleDropInvalidInputEnabled bool
	configRuleExternalClusterEnabled  bool

	initFlag    = false
	podCIDRIPv4 string
	podCIDRIPv6 string

	serviceCache = map[ctrl.Request]corev1.Service{}
)

// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch

func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithName("reconcile").WithValues("service", req.NamespacedName)
	var err error

	// ** Init service controller **
	// In SetupWithManager, function k8s client cannot be used.
	// So initialize controller in Reconile() function.
	if !initFlag {
		defer func() {
			initFlag = true
		}()

		// Init logger for only initialize controller
		logger := r.Log.WithName("initalize")
		logger.Info("initalize service contoller")

		// Get pod CIDR configs
		configPodCIDRIPv4, _ = configs.GetConfigPodCIDRIPv4()
		configPodCIDRIPv6, _ = configs.GetConfigPodCIDRIPv6()
		logger.WithValues("IPv4 pod cIDR", configPodCIDRIPv4).Info("config IPv4 pod CIDR")
		logger.WithValues("IPv6 pod cIDR", configPodCIDRIPv6).Info("config IPv6 pod CIDR")

		// Get rule configs
		configRuleDropInvalidInputEnabled, err = configs.GetConfigRuleDropInvalidInputEnabled()
		if err != nil {
			logger.Error(err, "config error")
			os.Exit(1)
		}
		configRuleExternalClusterEnabled, err = configs.GetConfigRuleExternalClusterEnabled()
		if err != nil {
			logger.Error(err, "config error")
			os.Exit(1)
		}
		logger.WithValues("enabled", configRuleDropInvalidInputEnabled).Info("config rule drop invalid packet in INPUT chain")
		logger.WithValues("enabled", configRuleExternalClusterEnabled).Info("config rule externalIP to clusterIP")

		// Init packages
		rules.Init(configPodCIDRIPv4, configPodCIDRIPv6)

		// Init or Cleanup rules
		if configRuleDropInvalidInputEnabled {
			if err := rules.InitRulesDropInvalidInput(logger); err != nil {
				logger.Error(err, "failed to init rule drop invalid packet in INPUT chain")
				os.Exit(1)
			}
		} else {
			if err := rules.CleanupRulesDropInvalidInput(logger); err != nil {
				logger.Error(err, "failed to cleanup rule drop invalid packet in INPUT chain")
				os.Exit(1)
			}
		}

		if configRuleExternalClusterEnabled {
			// Init externalIP to clusterIP rules
			if err := rules.InitRulesExternalCluster(logger); err != nil {
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
			if err := rules.CleanupRulesExternalCluster(logger, svcs); err != nil {
				logger.Error(err, "failed to cleanup rule externalIP to clusterIP")
				os.Exit(1)
			}
		} else {
			// Destroy externalIP to clusterIP rules
			if err := rules.DestoryRulesExternalCluster(logger); err != nil {
				logger.Error(err, "failed to destroy rule externalIP to clusterIP")
				os.Exit(1)
			}
		}

		// Run rules periodically
		ticker := time.NewTicker(60 * time.Second)
		go func() {
			for {
				<-ticker.C

				if configRuleDropInvalidInputEnabled {
					if err := rules.InitRulesDropInvalidInput(logger); err != nil {
						logger.Error(err, "failed to init rule drop invalid packet in INPUT chain")
					}
				}
			}
		}()
	}

	// ** Reconcile Loop **
	if configRuleExternalClusterEnabled {
		// In case the iptables chain is deleted, initalize again
		if err := rules.InitRulesExternalCluster(logger); err != nil {
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

				// Get service cluster IPs for each family
				oldClusterIPv4 := GetClusterIPByFamily(corev1.IPv4Protocol, &oldSvc)
				oldClusterIPv6 := GetClusterIPByFamily(corev1.IPv6Protocol, &oldSvc)

				// Get all the service's external IPs
				oldExternalIPs := []string{}
				for _, ingress := range svc.Status.LoadBalancer.Ingress {
					oldExternalIPs = append(oldExternalIPs, ingress.IP)
				}

				// Delete rules
				for _, oldExternalIP := range oldExternalIPs {
					oldClusterIP := oldClusterIPv4
					if ip.IsIPv6Addr(oldExternalIP) {
						oldClusterIP = oldClusterIPv6
					}

					// Delete rules
					logger.WithValues("externalIP", oldExternalIP).WithValues("clusterIP", oldClusterIP).Info("delete rule externalIp to clusterIP")
					if err := rules.DeleteRulesExternalCluster(logger, &req, oldClusterIP, oldExternalIP); err != nil {
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

		// Get service cluster IPs for each family
		clusterIPv4 := GetClusterIPByFamily(corev1.IPv4Protocol, svc)
		clusterIPv6 := GetClusterIPByFamily(corev1.IPv6Protocol, svc)

		// Get all the service's external IPs
		externalIPs := []string{}
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			externalIPs = append(externalIPs, ingress.IP)
		}

		// Create rules
		for _, externalIP := range externalIPs {
			clusterIP := clusterIPv4
			if ip.IsIPv6Addr(externalIP) {
				clusterIP = clusterIPv6
			}

			// Cache service to use deleting service
			serviceCache[req] = *svc.DeepCopy()

			// Create rules
			logger.WithValues("externalIP", externalIP).WithValues("clusterIP", clusterIP).Info("create iptables rules")
			if err := rules.CreateRulesExternalCluster(logger, &req, clusterIP, externalIP); err != nil {
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

// GetClusterIPByFamily returns a service clusterip by family
// Borrowed from https://github.com/kubernetes/kubernetes/blob/v1.20.5/pkg/proxy/util/utils.go#L386-L411
func GetClusterIPByFamily(ipFamily corev1.IPFamily, service *corev1.Service) string {
	// allowing skew
	if len(service.Spec.IPFamilies) == 0 {
		if len(service.Spec.ClusterIP) == 0 || service.Spec.ClusterIP == corev1.ClusterIPNone {
			return ""
		}

		IsIPv6Family := (ipFamily == corev1.IPv6Protocol)
		if IsIPv6Family == ip.IsIPv6Addr(service.Spec.ClusterIP) {
			return service.Spec.ClusterIP
		}

		return ""
	}

	for idx, family := range service.Spec.IPFamilies {
		if family == ipFamily {
			if idx < len(service.Spec.ClusterIPs) {
				return service.Spec.ClusterIPs[idx]
			}
		}
	}

	return ""
}
