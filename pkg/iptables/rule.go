package iptables

import (
	"strings"
)

func getValue(rule, opt string) string {
	tokens := strings.Split(rule, " ")
	for i, token := range tokens {
		if token == opt {
			if len(tokens)-1 >= i+1 {
				return tokens[i+1]
			} else {
				return ""
			}
		}
	}
	return ""
}

func GetRuleComment(rule string) string {
	return strings.Trim(getValue(rule, "--comment"), "\"")
}

func GetRuleSrc(rule string) string {
	return getValue(rule, "-s")
}

func GetRuleDest(rule string) string {
	return getValue(rule, "-d")
}

func GetRuleJump(rule string) string {
	return getValue(rule, "-j")
}

func GetRuleDNATDest(rule string) string {
	return getValue(rule, "--to-destination")
}

func ChangeRuleToDelete(rule string) []string {
	return strings.Split(strings.ReplaceAll(rule[3:], "\"", ""), " ")
}
