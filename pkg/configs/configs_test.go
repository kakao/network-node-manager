package configs

import (
	"os"
	"testing"
)

func TestGetConfigNodeName(t *testing.T) {
	os.Setenv(EnvNodeName, "node")
	nodeName, _ := GetConfigNodeName()
	if nodeName != "node" {
		t.Errorf("wrong result - %s", "node")
	}
}

func TestGetConfigNetStack(t *testing.T) {
	os.Setenv(EnvNetStack, "ipv4")
	ipv4, ipv6, _ := GetConfigNetStack()
	if ipv4 != true || ipv6 != false {
		t.Errorf("wrong result - %s", "ipv4")
	}

	os.Setenv(EnvNetStack, "IPV4")
	ipv4, ipv6, _ = GetConfigNetStack()
	if ipv4 != true || ipv6 != false {
		t.Errorf("wrong result - %s", "IPV4")
	}

	os.Setenv(EnvNetStack, "ipv6")
	ipv4, ipv6, _ = GetConfigNetStack()
	if ipv4 != false || ipv6 != true {
		t.Errorf("wrong result - %s", "ipv6")
	}

	os.Setenv(EnvNetStack, "IPV6")
	ipv4, ipv6, _ = GetConfigNetStack()
	if ipv4 != false || ipv6 != true {
		t.Errorf("wrong result - %s", "IPV6")
	}

	os.Setenv(EnvNetStack, "ipv5")
	_, _, err := GetConfigNetStack()
	if err == nil {
		t.Errorf("wrong result - %s", "ipv5")
	}

	os.Setenv(EnvNetStack, "ipv4,ipv6")
	ipv4, ipv6, _ = GetConfigNetStack()
	if ipv4 != true || ipv6 != true {
		t.Errorf("wrong result - %s", "ipv4,ipv6")
	}

	os.Setenv(EnvNetStack, "ipv4, ipv6")
	ipv4, ipv6, _ = GetConfigNetStack()
	if ipv4 != true || ipv6 != true {
		t.Errorf("wrong result - %s", "ipv4, ipv6")
	}

	os.Setenv(EnvNetStack, "ipv6, ipv4")
	ipv4, ipv6, _ = GetConfigNetStack()
	if ipv4 != true || ipv6 != true {
		t.Errorf("wrong result - %s", "ipv6,ipv4")
	}
}
