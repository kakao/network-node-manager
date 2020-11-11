package controllers

import (
	"testing"
)

const (
	ipv4ValidCIDR   = "192.168.0.0/24"
	ipv4InvaildCIDR = "192.999.0.0/24"

	ipv6ValidCIDR   = "ffaa::/64"
	ipv6InvalidCIDR = "fdzz::/64"
)

func TestGetPodCIDR(t *testing.T) {
	ipv4CIDR, ipv6CIDR := getPodCIDR([]string{ipv4ValidCIDR, ipv4InvaildCIDR, ipv6ValidCIDR, ipv6InvalidCIDR})
	if ipv4CIDR != ipv4ValidCIDR {
		t.Errorf("wrong result - %s", ipv4ValidCIDR)
	}
	if ipv6CIDR != ipv6ValidCIDR {
		t.Errorf("wrong result - %s", ipv6ValidCIDR)
	}
}
