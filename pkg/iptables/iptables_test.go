package iptables

import (
	"testing"
)

const (
	chainTest = "TestChain"
)

var (
	ruleDNATIPv4 = []string{"-j", "DNAT", "--to-destination", "192.168.0.1"}
	ruleDNATIPv6 = []string{"-j", "DNAT", "--to-destination", "fdaa::1"}
)

func TestCreateChainIPv4(t *testing.T) {
	if out, err := CreateChainIPv4(TableNAT, chainTest); err != nil {
		t.Errorf("create chain IPv4 - out:%s", out)
	}
	if !IsExistChainIPv4(TableNAT, chainTest) {
		t.Error("check created chain IPv4")
	}
}

func TestCreateChainIPv6(t *testing.T) {
	if out, err := CreateChainIPv6(TableNAT, chainTest); err != nil {
		t.Errorf("create chain IPv6 - out:%s", out)
	}
	if !IsExistChainIPv6(TableNAT, chainTest) {
		t.Error("check created chain IPv6")
	}
}

func TestCreateDeleteRuleIPv4(t *testing.T) {
	// Create
	if out, err := CreateRuleFirstIPv4(TableNAT, chainTest, "create first rule", ruleDNATIPv4...); err != nil {
		t.Errorf("create first rule IPv4 - out:%s", out)
	}
	if !IsExistRuleIPv4(TableNAT, chainTest, "create first rule", ruleDNATIPv4...) {
		t.Errorf("check created rule IPv4")
	}

	// Delete
	if out, err := DeleteRuleIPv4(TableNAT, chainTest, "create first rule", ruleDNATIPv4...); err != nil {
		t.Errorf("delete rule IPv4 - out:%s", out)
	}
	if IsExistRuleIPv4(TableNAT, chainTest, "create first rule", ruleDNATIPv4...) {
		t.Errorf("check deleted rule IPv4")
	}
}

func TestCreateDeleteRuleIPv6(t *testing.T) {
	// Create
	if out, err := CreateRuleFirstIPv6(TableNAT, chainTest, "create first rule", ruleDNATIPv6...); err != nil {
		t.Errorf("create first rule IPv6 - out:%s", out)
	}
	if !IsExistRuleIPv6(TableNAT, chainTest, "create first rule", ruleDNATIPv6...) {
		t.Errorf("check created rule IPv6")
	}

	// Delete
	if out, err := DeleteRuleIPv6(TableNAT, chainTest, "create first rule", ruleDNATIPv6...); err != nil {
		t.Errorf("delete rule IPv6 - out:%s", out)
	}
	if IsExistRuleIPv6(TableNAT, chainTest, "create first rule", ruleDNATIPv6...) {
		t.Errorf("check deleted rule IPv6")
	}
}

func TestDeleteChainIPv4(t *testing.T) {
	if out, err := DeleteChainIPv4(TableNAT, chainTest); err != nil {
		t.Errorf("delete chain IPv4 - out:%s", out)
	}
	if IsExistChainIPv4(TableNAT, chainTest) {
		t.Error("check deleted chain IPv4")
	}
}

func TestDeleteChainIPv6(t *testing.T) {
	if out, err := DeleteChainIPv6(TableNAT, chainTest); err != nil {
		t.Errorf("delete chain IPv6 - out:%s", out)
	}
	if IsExistChainIPv6(TableNAT, chainTest) {
		t.Error("check deleted chain IPv6")
	}
}
