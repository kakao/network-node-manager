package iptables

import (
	"reflect"
	"testing"
)

const (
	ruleTest = "-A testChain -s 192.168.0.1 -d 192.168.0.2 -m comment --comment \"testComment\" --to-destination 192.168.0.3"
)

var (
	ruleDeleteTest = []string{"testChain", "-s", "192.168.0.1", "-d", "192.168.0.2", "-m", "comment", "--comment", "testComment", "--to-destination", "192.168.0.3"}
)

func TestGetValue(t *testing.T) {
	value := getValue(ruleTest, "-A")
	if value != "testChain" {
		t.Errorf("get value")
	}
}

func TestGetRuleComment(t *testing.T) {
	value := GetRuleComment(ruleTest)
	if value != "testComment" {
		t.Errorf("get rule comment")
	}
}

func TestGetRuleSrc(t *testing.T) {
	value := GetRuleSrc(ruleTest)
	if value != "192.168.0.1" {
		t.Errorf("get rule src")
	}
}

func TestGetRuleDNATDest(t *testing.T) {
	value := GetRuleDNATDest(ruleTest)
	if value != "192.168.0.3" {
		t.Errorf("get rule DNAT dest")
	}
}

func TestChangeRuleToDelete(t *testing.T) {
	value := ChangeRuleToDelete(ruleTest)
	if !reflect.DeepEqual(value, ruleDeleteTest) {
		t.Errorf("change rule to delete. expected:%+v / actual:%+v", ruleDeleteTest, value)
	}
}
