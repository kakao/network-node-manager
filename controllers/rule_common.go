package controllers

import (
	"github.com/go-logr/logr"

	"github.com/kakao/network-node-manager/pkg/configs"
	"github.com/kakao/network-node-manager/pkg/iptables"
)

// Constants
const (
	ChainFilterInput   = "INPUT"
	ChainNATPrerouting = "PREROUTING"
	ChainNATOutput     = "OUTPUT"

	ChainNATKubeMarkMasq = "KUBE-MARK-MASQ"

	ChainFilterBaseInput   = "NMANAGER_INPUT"
	ChainNATBasePrerouting = "NMANAGER_PREROUTING"
	ChainNATBaseOutput     = "NMANAGER_OUTPUT"

	ChainFilterDropInvalidInput       = "NMANAGER_DROP_INVALID_INPUT"
	ChainNATExternalClusterPrerouting = "NMANAGER_EX_CLUS_PREROUTING"
	ChainNATExternalClusterOutput     = "NMANAGER_EX_CLUS_OUTPUT"
)

func initBaseChains(logger logr.Logger) error {
	// Get network stack config
	configIPv4Enabled, configIPv6Enabled, err := configs.GetConfigNetStack()
	if err != nil {
		return err
	}

	// IPv4
	if configIPv4Enabled {
		// Create base chain in tables
		out, err := iptables.CreateChainIPv4(iptables.TableFilter, ChainFilterBaseInput)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv4(iptables.TableNAT, ChainNATBasePrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv4(iptables.TableNAT, ChainNATBaseOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Create jump rule to each chain in tables
		ruleJumpFilterInput := []string{"-j", ChainFilterBaseInput}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableFilter, ChainFilterInput, "", ruleJumpFilterInput...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpNATPre := []string{"-j", ChainNATBasePrerouting}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableNAT, ChainNATPrerouting, "", ruleJumpNATPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpNATOut := []string{"-j", ChainNATBaseOutput}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableNAT, ChainNATOutput, "", ruleJumpNATOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

	}

	// IPv6
	if configIPv6Enabled {
		// Create base chain in nat table
		out, err := iptables.CreateChainIPv6(iptables.TableFilter, ChainFilterBaseInput)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv6(iptables.TableNAT, ChainNATBasePrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv6(iptables.TableNAT, ChainNATBaseOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Create jump rule to each chain in nat table
		ruleJumpFilterInput := []string{"-j", ChainFilterBaseInput}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableFilter, ChainFilterInput, "", ruleJumpFilterInput...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpNATPre := []string{"-j", ChainNATBasePrerouting}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableNAT, ChainNATPrerouting, "", ruleJumpNATPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpNATOut := []string{"-j", ChainNATBaseOutput}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableNAT, ChainNATOutput, "", ruleJumpNATOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	return nil
}
