package rules

import (
	"github.com/go-logr/logr"

	"github.com/kakao/network-node-manager/pkg/ip"
	"github.com/kakao/network-node-manager/pkg/iptables"
)

func InitRulesDropInvalidInput(logger logr.Logger) error {
	// Init base chains
	if err := initBaseChains(logger); err != nil {
		logger.Error(err, "failed to init base chain for externalIP to clusterIP Rules")
		return err
	}

	// IPv4
	if ip.IsIPv4CIDR(podCIDRIPv4) {
		// Create chain
		out, err := iptables.CreateChainIPv4(iptables.TableFilter, ChainFilterDropInvalidInput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set drop rule
		ruleDrop := []string{"-m", "conntrack", "--ctstate", "INVALID", "-j", "DROP"}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableFilter, ChainFilterDropInvalidInput, "", ruleDrop...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set jump rule
		ruleJump := []string{"-j", ChainFilterDropInvalidInput}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableFilter, ChainBaseInput, "", ruleJump...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	// IPv6
	if ip.IsIPv6CIDR(podCIDRIPv6) {
		// Create chain
		out, err := iptables.CreateChainIPv6(iptables.TableFilter, ChainFilterDropInvalidInput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set drop rule
		ruleDrop := []string{"-m", "conntrack", "--ctstate", "INVALID", "-j", "DROP"}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableFilter, ChainFilterDropInvalidInput, "", ruleDrop...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set jump rule
		ruleJump := []string{"-j", ChainFilterDropInvalidInput}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableFilter, ChainBaseInput, "", ruleJump...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	return nil
}

func CleanupRulesDropInvalidInput(logger logr.Logger) error {
	// IPv4
	if ip.IsIPv4CIDR(podCIDRIPv4) {
		// Delete jump rule
		ruleJump := []string{"-j", ChainFilterDropInvalidInput}
		out, err := iptables.DeleteRuleIPv4(iptables.TableFilter, ChainBaseInput, "", ruleJump...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Delete chain
		out, err = iptables.DeleteChainIPv4(iptables.TableFilter, ChainFilterDropInvalidInput)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	// IPv6
	if ip.IsIPv6CIDR(podCIDRIPv6) {
		// Delete jump rule
		ruleJump := []string{"-j", ChainFilterDropInvalidInput}
		out, err := iptables.DeleteRuleIPv6(iptables.TableFilter, ChainBaseInput, "", ruleJump...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Delete chain
		out, err = iptables.DeleteChainIPv6(iptables.TableFilter, ChainFilterDropInvalidInput)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	return nil
}
