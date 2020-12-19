package configs

import (
	"fmt"
	"os"
	"strings"

	"github.com/kakao/network-node-manager/pkg/ip"
)

const (
	EnvConfigTrue  = "true"
	EnvConfigFalse = "false"

	EnvPodCIDRIPv4 = "POD_CIDR_IPV4"
	EnvPodCIDRIPv6 = "POD_CIDR_IPV6"

	EnvRuleDropInvalidInputEnable = "RULE_DROP_INVALID_INPUT_ENABLE"
	EnvRuleExternalClusterEnable  = "RULE_EXTERNAL_CLUSTER_ENABLE"
)

func GetConfigPodCIDRIPv4() (string, error) {
	cidr := os.Getenv(EnvPodCIDRIPv4)
	cidr = strings.Replace(cidr, " ", "", -1)

	if cidr == "" {
		return "", fmt.Errorf("IPv4 pod CIDR isn't set")
	} else if !ip.IsIPv4CIDR(cidr) {
		return "", fmt.Errorf("wrong IPv4 pod CIDR")
	}
	return cidr, nil
}

func GetConfigPodCIDRIPv6() (string, error) {
	cidr := os.Getenv(EnvPodCIDRIPv6)
	cidr = strings.Replace(cidr, " ", "", -1)

	if cidr == "" {
		return "", fmt.Errorf("IPv6 pod CIDR isn't set")
	} else if !ip.IsIPv6CIDR(cidr) {
		return "", fmt.Errorf("wrong IPv6 pod CIDR")
	}
	return cidr, nil
}

func GetConfigRuleDropInvalidInputEnabled() (bool, error) {
	config := os.Getenv(EnvRuleDropInvalidInputEnable)
	config = strings.ToLower(config)

	if config == "" {
		return true, nil
	} else if config == EnvConfigFalse {
		return false, nil
	} else if config == EnvConfigTrue {
		return true, nil
	}
	return false, fmt.Errorf("wrong config for drop invalid packet in INPUT chain : %s", config)
}

func GetConfigRuleExternalClusterEnabled() (bool, error) {
	config := os.Getenv(EnvRuleExternalClusterEnable)
	config = strings.ToLower(config)

	if config == "" {
		return false, nil
	} else if config == EnvConfigFalse {
		return false, nil
	} else if config == EnvConfigTrue {
		return true, nil
	}
	return false, fmt.Errorf("wrong config for externalIP to clusterIP DNAT : %s", config)
}
