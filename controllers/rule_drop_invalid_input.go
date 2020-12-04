package controllers

import (
	"github.com/go-logr/logr"

	"github.com/kakao/network-node-manager/pkg/iptables"
)

func createRulesDropInvalidInput(logger logr.Logger) error {
	if err := initBaseChains(logger); err != nil {
		logger.Error(err, "failed to init base chain for externalIP to clusterIP Rules")
		return err
	}

	// IPv4
	if configIPv4Enabled {
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
	if configIPv6Enabled {
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

func deleteRulesDropInvalidInput(logger logr.Logger) error {
	// IPv4
	if configIPv4Enabled {
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
	if configIPv6Enabled {
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
