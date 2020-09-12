package ip

import (
	"testing"
)

const (
	ipv4LocalHost = "127.0.0.1"
	ipv4InvaildIP = "999.0.0.1"

	ipv6LocalHost = "::1"
	ipv6InvaildIP = "fdaa:1"
)

func TestIsValidIP(t *testing.T) {
	v := IsVaildIP(ipv4LocalHost)
	if v != true {
		t.Errorf("wrong result - %s", ipv4LocalHost)
	}

	v = IsVaildIP(ipv4InvaildIP)
	if v != false {
		t.Errorf("wrong result - %s", ipv4InvaildIP)
	}

	v = IsVaildIP(ipv6LocalHost)
	if v != true {
		t.Errorf("wrong result - %s", ipv6LocalHost)
	}

	v = IsVaildIP(ipv4InvaildIP)
	if v != false {
		t.Errorf("wrong result - %s", ipv6InvaildIP)
	}
}

func TestIsIPv4Addr(t *testing.T) {
	v := IsIPv4Addr(ipv4LocalHost)
	if v != true {
		t.Errorf("wrong result - %s", ipv4LocalHost)
	}

	v = IsIPv4Addr(ipv4InvaildIP)
	if v != false {
		t.Errorf("wrong result - %s", ipv4InvaildIP)
	}

	v = IsIPv4Addr(ipv6LocalHost)
	if v != false {
		t.Errorf("wrong - %s", ipv6LocalHost)
	}
}

func TestIsIPv6Addr(t *testing.T) {
	v := IsIPv6Addr(ipv6LocalHost)
	if v != true {
		t.Errorf("wrong result - %s", ipv6LocalHost)
	}

	v = IsIPv6Addr(ipv6InvaildIP)
	if v != false {
		t.Errorf("wrong result - %s", ipv6InvaildIP)
	}

	v = IsIPv6Addr(ipv4LocalHost)
	if v != false {
		t.Errorf("wrong - %s", ipv4LocalHost)
	}
}
