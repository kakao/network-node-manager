package configs

import (
	"fmt"
	"os"
	"strings"
)

const (
	EnvNodeName    = "NODE_NAME"
	EnvConfigTrue  = "true"
	EnvConfigFalse = "false"

	EnvNetStack     = "NET_STACK"
	EnvNetStackIPv4 = "ipv4"
	EnvNetStackIPv6 = "ipv6"

	EnvRuleDropNotTrackDNS  = "RULE_NOT_TRACK_DNS"
	EnvRuleDropInvalidInput = "RULE_DROP_INVALID_INPUT"
	EnvRuleExternalCluster  = "RULE_EXTERNAL_CLUSTER"
)

func GetConfigNodeName() (string, error) {
	// return config
	config := os.Getenv(EnvNodeName)
	if config == "" {
		return "", fmt.Errorf("failed to get node name of controller's pod from %s env", EnvNodeName)
	}
	return config, nil
}

func GetConfigNetStack() (bool, bool, error) {
	// organize configs
	configs := os.Getenv(EnvNetStack)
	configs = strings.Replace(configs, " ", "", -1)
	configs = strings.ToLower(configs)
	if configs == "" {
		return true, false, nil
	}

	// return configs
	ipv4 := false
	ipv6 := false
	for _, config := range strings.Split(configs, ",") {
		if config == EnvNetStackIPv4 {
			ipv4 = true
		} else if config == EnvNetStackIPv6 {
			ipv6 = true
		} else {
			return false, false, fmt.Errorf("wrong config for network stack : %s", config)
		}
	}
	return ipv4, ipv6, nil
}

func GetConfigRuleNotTrackDNS() (bool, error) {
	// organize configs
	config := os.Getenv(EnvRuleDropNotTrackDNS)
	config = strings.ToLower(config)
	if config == "" {
		return true, nil
	}

	// return configs
	if config == EnvConfigFalse {
		return false, nil
	} else if config == EnvConfigTrue {
		return true, nil
	}
	return false, fmt.Errorf("wrong config for not tracking DNS packet : %s", config)
}

func GetConfigRuleDropInvalidInput() (bool, error) {
	// organize configs
	config := os.Getenv(EnvRuleDropInvalidInput)
	config = strings.ToLower(config)
	if config == "" {
		return true, nil
	}

	// return configs
	if config == EnvConfigFalse {
		return false, nil
	} else if config == EnvConfigTrue {
		return true, nil
	}
	return false, fmt.Errorf("wrong config for drop invalid packet in INPUT chain : %s", config)
}

func GetConfigRuleExternalCluster() (bool, error) {
	// organize configs
	config := os.Getenv(EnvRuleExternalCluster)
	config = strings.ToLower(config)
	if config == "" {
		return false, nil
	}

	// return configs
	if config == EnvConfigFalse {
		return false, nil
	} else if config == EnvConfigTrue {
		return true, nil
	}
	return false, fmt.Errorf("wrong config for externalIP to clusterIP DNAT : %s", config)
}
