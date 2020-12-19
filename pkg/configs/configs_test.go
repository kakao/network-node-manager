package configs

import (
	"os"
	"testing"
)

func TestGetConfigPodCIDR(t *testing.T) {
	os.Setenv(EnvPodCIDRIPv4, "")
	_, err := GetConfigPodCIDRIPv4()
	if err == nil {
		t.Errorf("wrong result - %s", "empty")
	}

	os.Setenv(EnvPodCIDRIPv4, "127.0.0.1/24")
	_, err = GetConfigPodCIDRIPv4()
	if err != nil {
		t.Errorf("wrong result - %s", "127.0.0.1/24")
	}

	os.Setenv(EnvPodCIDRIPv4, "127.0.0.1/40")
	_, err = GetConfigPodCIDRIPv4()
	if err == nil {
		t.Errorf("wrong result - %s", "127.0.0.1/40")
	}
}

func TestGetConfigRuleExternalCluster(t *testing.T) {
	os.Setenv(EnvRuleExternalClusterEnable, "")
	enabled, _ := GetConfigRuleExternalClusterEnabled()
	if enabled {
		t.Errorf("wrong result - %s", "")
	}

	os.Setenv(EnvRuleExternalClusterEnable, "false")
	enabled, _ = GetConfigRuleExternalClusterEnabled()
	if enabled {
		t.Errorf("wrong result - %s", "false")
	}

	os.Setenv(EnvRuleExternalClusterEnable, "true")
	enabled, _ = GetConfigRuleExternalClusterEnabled()
	if !enabled {
		t.Errorf("wrong result - %s", "true")
	}

	os.Setenv(EnvRuleExternalClusterEnable, "none")
	_, err := GetConfigRuleExternalClusterEnabled()
	if err == nil {
		t.Errorf("wrong result - %s", "none")
	}
}
