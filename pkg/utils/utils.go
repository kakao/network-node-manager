package utils

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/kakao/network-node-manager/pkg/ip"
)

// GetClusterIPByFamily returns a service clusterip by family
// Reference - https://github.com/kubernetes/kubernetes/blob/v1.20.5/pkg/proxy/util/utils.go#L386-L411
func GetClusterIPByFamily(ipFamily corev1.IPFamily, service *corev1.Service) string {
	// Process when the service doesn't have IPFamiles info
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

	// Process when the service has IPFamlies info
	for idx, family := range service.Spec.IPFamilies {
		if family == ipFamily {
			if idx < len(service.Spec.ClusterIPs) {
				return service.Spec.ClusterIPs[idx]
			}
		}
	}
	return ""
}
