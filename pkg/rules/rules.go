package rules

import (
	"github.com/go-logr/logr"

	"github.com/kakao/network-node-manager/pkg/ip"
	"github.com/kakao/network-node-manager/pkg/iptables"
)

// Constants
const (
	ChainInput      = "INPUT"
	ChainPrerouting = "PREROUTING"
	ChainOutput     = "OUTPUT"

	ChainBasePrerouting = "NMANAGER_PREROUTING"
	ChainBaseInput      = "NMANAGER_INPUT"
	ChainBaseOutput     = "NMANAGER_OUTPUT"

	ChainFilterDropInvalidInput       = "NMANAGER_DROP_INVALID_INPUT"
	ChainNATExternalClusterPrerouting = "NMANAGER_EX_CLUS_PREROUTING"
	ChainNATExternalClusterOutput     = "NMANAGER_EX_CLUS_OUTPUT"

	ChainNATKubeMarkMasq = "KUBE-MARK-MASQ"
)

// Vars
var (
	podCIDRIPv4 string
	podCIDRIPv6 string
)

func Init(cidrIPv4, cidrIPv6 string) {
	podCIDRIPv4 = cidrIPv4
	podCIDRIPv6 = cidrIPv6
}

func initBaseChains(logger logr.Logger) error {
	// IPv4
	if ip.IsIPv4CIDR(podCIDRIPv4) {
		// Create base chain in tables
		out, err := iptables.CreateChainIPv4(iptables.TableFilter, ChainBaseInput)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv4(iptables.TableNAT, ChainBasePrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv4(iptables.TableNAT, ChainBaseOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Create jump rule to each chain in tables
		ruleJumpFilterInput := []string{"-j", ChainBaseInput}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableFilter, ChainInput, "", ruleJumpFilterInput...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpNATPre := []string{"-j", ChainBasePrerouting}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableNAT, ChainPrerouting, "", ruleJumpNATPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpNATOut := []string{"-j", ChainBaseOutput}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableNAT, ChainOutput, "", ruleJumpNATOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	// IPv6
	if ip.IsIPv6CIDR(podCIDRIPv6) {
		// Create base chain in nat table
		out, err := iptables.CreateChainIPv6(iptables.TableFilter, ChainBaseInput)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv6(iptables.TableNAT, ChainBasePrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv6(iptables.TableNAT, ChainBaseOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Create jump rule to each chain in tables
		ruleJumpFilterInput := []string{"-j", ChainBaseInput}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableFilter, ChainInput, "", ruleJumpFilterInput...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpNATPre := []string{"-j", ChainBasePrerouting}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableNAT, ChainPrerouting, "", ruleJumpNATPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpNATOut := []string{"-j", ChainBaseOutput}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableNAT, ChainOutput, "", ruleJumpNATOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	return nil
}
