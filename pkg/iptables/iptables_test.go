package iptables

import (
	"testing"
)

const (
	chainTest   = "TestChain"
	commentTest = "TestComment"
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

func TestCreateGetDeleteRuleIPv4(t *testing.T) {
	// Create
	if out, err := CreateRuleFirstIPv4(TableNAT, chainTest, commentTest, ruleDNATIPv4...); err != nil {
		t.Errorf("create first rule IPv4 - out:%s", out)
	}
	if !IsExistRuleIPv4(TableNAT, chainTest, commentTest, ruleDNATIPv4...) {
		t.Errorf("check created rule IPv4")
	}

	// Get
	rules, err := GetRulesIPv4(TableNAT, chainTest)
	if err != nil {
		t.Errorf("get rules IPv4")
	}
	if len(rules) == 0 {
		t.Errorf("no rules IPv4")
	}
	expectedRule := "-A " + chainTest + " -m comment --comment " + commentTest + " -j DNAT --to-destination 192.168.0.1"
	if rules[0] != expectedRule {
		t.Errorf("rule is different. expected:%s / actual:%s", expectedRule, rules[0])
	}

	// Delete
	if out, err := DeleteRuleIPv4(TableNAT, chainTest, commentTest, ruleDNATIPv4...); err != nil {
		t.Errorf("delete rule IPv4 - out:%s", out)
	}
	if IsExistRuleIPv4(TableNAT, chainTest, commentTest, ruleDNATIPv4...) {
		t.Errorf("check deleted rule IPv4")
	}
}

func TestCreateGetDeleteRuleIPv6(t *testing.T) {
	// Create
	if out, err := CreateRuleFirstIPv6(TableNAT, chainTest, commentTest, ruleDNATIPv6...); err != nil {
		t.Errorf("create first rule IPv6 - out:%s", out)
	}
	if !IsExistRuleIPv6(TableNAT, chainTest, commentTest, ruleDNATIPv6...) {
		t.Errorf("check created rule IPv6")
	}

	// Get
	rules, err := GetRulesIPv6(TableNAT, chainTest)
	if err != nil {
		t.Errorf("get rules IPv6")
	}
	if len(rules) == 0 {
		t.Errorf("no rules IPv6")
	}
	expectedRule := "-A " + chainTest + " -m comment --comment " + commentTest + " -j DNAT --to-destination fdaa::1"
	if rules[0] != expectedRule {
		t.Errorf("rule is different. expected:%s / actual:%s", expectedRule, rules[0])
	}

	// Delete
	if out, err := DeleteRuleIPv6(TableNAT, chainTest, commentTest, ruleDNATIPv6...); err != nil {
		t.Errorf("delete rule IPv6 - out:%s", out)
	}
	if IsExistRuleIPv6(TableNAT, chainTest, commentTest, ruleDNATIPv6...) {
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
