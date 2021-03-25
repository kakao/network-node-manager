package utils

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

const (
	ipv4Local = "127.0.0.1"
	ipv6Local = "::1"
)

var (
	ipv4Svc = corev1.Service{
		Spec: corev1.ServiceSpec{
			ClusterIP: ipv4Local,
		},
	}

	ipv6Svc = corev1.Service{
		Spec: corev1.ServiceSpec{
			ClusterIP: ipv6Local,
		},
	}

	noneSvc = corev1.Service{
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone,
		},
	}

	ipv4SvcFamily = corev1.Service{
		Spec: corev1.ServiceSpec{
			IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
			ClusterIPs: []string{ipv4Local},
		},
	}

	ipv6SvcFamily = corev1.Service{
		Spec: corev1.ServiceSpec{
			IPFamilies: []corev1.IPFamily{corev1.IPv6Protocol},
			ClusterIPs: []string{ipv6Local},
		},
	}

	ipv46SvcFamily = corev1.Service{
		Spec: corev1.ServiceSpec{
			IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol},
			ClusterIPs: []string{ipv4Local, ipv6Local},
		},
	}
)

func TestGetClusterIPByFamily(t *testing.T) {
	clusterIP := GetClusterIPByFamily(corev1.IPv4Protocol, &ipv4Svc)
	if clusterIP != ipv4Local {
		t.Errorf("wrong result - ipv4Svc")
	}

	clusterIP = GetClusterIPByFamily(corev1.IPv6Protocol, &ipv6Svc)
	if clusterIP != ipv6Local {
		t.Errorf("wrong result - ipv6Svc")
	}

	clusterIP = GetClusterIPByFamily(corev1.IPv4Protocol, &noneSvc)
	if clusterIP != "" {
		t.Errorf("wrong result - noneSvc - ipv4")
	}

	clusterIP = GetClusterIPByFamily(corev1.IPv6Protocol, &noneSvc)
	if clusterIP != "" {
		t.Errorf("wrong result - noneSvc - ipv6")
	}

	clusterIP = GetClusterIPByFamily(corev1.IPv4Protocol, &ipv4SvcFamily)
	if clusterIP != ipv4Local {
		t.Errorf("wrong result - ipv4SvcFamily")
	}

	clusterIP = GetClusterIPByFamily(corev1.IPv6Protocol, &ipv6SvcFamily)
	if clusterIP != ipv6Local {
		t.Errorf("wrong result - ipv6SvcFamily")
	}

	clusterIP = GetClusterIPByFamily(corev1.IPv4Protocol, &ipv46SvcFamily)
	if clusterIP != ipv4Local {
		t.Errorf("wrong result - ipv46SvcFamily - ipv4")
	}

	clusterIP = GetClusterIPByFamily(corev1.IPv6Protocol, &ipv46SvcFamily)
	if clusterIP != ipv6Local {
		t.Errorf("wrong result - ipv46SvcFamily - ipv6")
	}
}
