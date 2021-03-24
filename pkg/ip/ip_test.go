package ip

import (
	"testing"
)

const (
	ipv4Local       = "127.0.0.1"
	ipv4InvalidAddr = "999.0.0.1"
	ipv6Local       = "::1"
	ipv6InvalidAddr = "fffff::1"

	ipv4LocalCIDR       = "127.0.0.1/24"
	ipv4InvalidAddrCIDR = "999.0.0.1/24"
	ipv4InvalidMaskCIDR = "127.0.0.1/999"
	ipv6LocalCIDR       = "::1/64"
	ipv6InvalidAddrCIDR = "::fffff/64"
	ipv6InvalidMaskCIDR = "::1/999"
)

func TestIsValidIP(t *testing.T) {
	v := IsVaildIP(ipv4Local)
	if v != true {
		t.Errorf("wrong result - %s", ipv4Local)
	}

	v = IsVaildIP(ipv4InvalidAddr)
	if v != false {
		t.Errorf("wrong result - %s", ipv4InvalidAddr)
	}

	v = IsVaildIP(ipv6Local)
	if v != true {
		t.Errorf("wrong result - %s", ipv6Local)
	}

	v = IsVaildIP(ipv4InvalidAddr)
	if v != false {
		t.Errorf("wrong result - %s", ipv6InvalidAddr)
	}
}

func TestIsIPv4Addr(t *testing.T) {
	v := IsIPv4Addr(ipv4Local)
	if v != true {
		t.Errorf("wrong result - %s", ipv4Local)
	}

	v = IsIPv4Addr(ipv4InvalidAddr)
	if v != false {
		t.Errorf("wrong result - %s", ipv4InvalidAddr)
	}

	v = IsIPv4Addr(ipv6Local)
	if v != false {
		t.Errorf("wrong result - %s", ipv6Local)
	}
}

func TestIsIPv6Addr(t *testing.T) {
	v := IsIPv6Addr(ipv6Local)
	if v != true {
		t.Errorf("wrong result - %s", ipv6Local)
	}

	v = IsIPv6Addr(ipv6InvalidAddr)
	if v != false {
		t.Errorf("wrong result - %s", ipv6InvalidAddr)
	}

	v = IsIPv6Addr(ipv4Local)
	if v != false {
		t.Errorf("wrong result - %s", ipv4Local)
	}
}

func TestIsIPv4CIDR(t *testing.T) {
	v := IsIPv4CIDR(ipv4LocalCIDR)
	if v != true {
		t.Errorf("wrong result - %s", ipv4LocalCIDR)
	}

	v = IsIPv4CIDR(ipv4InvalidAddrCIDR)
	if v != false {
		t.Errorf("wrong result - %s", ipv4InvalidAddrCIDR)
	}

	v = IsIPv4CIDR(ipv4InvalidMaskCIDR)
	if v != false {
		t.Errorf("wrong result - %s", ipv4InvalidMaskCIDR)
	}

	v = IsIPv4CIDR(ipv6LocalCIDR)
	if v != false {
		t.Errorf("wrong result - %s", ipv6LocalCIDR)
	}
}

func TestIsIPv6CIDR(t *testing.T) {
	v := IsIPv6CIDR(ipv6LocalCIDR)
	if v != true {
		t.Errorf("wrong result - %s", ipv6LocalCIDR)
	}

	v = IsIPv6CIDR(ipv6InvalidAddrCIDR)
	if v != false {
		t.Errorf("wrong result - %s", ipv6InvalidAddrCIDR)
	}

	v = IsIPv6CIDR(ipv6InvalidMaskCIDR)
	if v != false {
		t.Errorf("wrong result - %s", ipv6InvalidMaskCIDR)
	}

	v = IsIPv6CIDR(ipv4LocalCIDR)
	if v != false {
		t.Errorf("wrong result - %s", ipv4LocalCIDR)
	}
}
