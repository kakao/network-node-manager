package configs

import (
	"fmt"
	"os"
	"strings"
)

const (
	EnvNodeName = "NODE_NAME"

	EnvNetStack     = "NET_STACK"
	EnvNetStackIPv4 = "ipv4"
	EnvNetStackIPv6 = "ipv6"
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
			return false, false, fmt.Errorf("wrong network stack config : %s", config)
		}
	}
	return ipv4, ipv6, nil
}
