package controllers

import (
	"github.com/go-logr/logr"

	"github.com/kakao/network-node-manager/pkg/iptables"
)

func createRulesNotTrackDNS(logger logr.Logger) error {
	if err := initBaseChains(logger); err != nil {
		logger.Error(err, "failed to init base chain for externalIP to clusterIP Rules")
		return err
	}

	// IPv4
	if configIPv4Enabled {
		// Create chains
		out, err := iptables.CreateChainIPv4(iptables.TableRaw, ChainRawNotTrackDNSPrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv4(iptables.TableRaw, ChainRawNotTrackDNSOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set not track rules
		ruleNotTrackSport := []string{"-p", "UDP", "-m", "udp", "--sport", "53", "-j", "CT", "--notrack"}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableRaw, ChainRawNotTrackDNSPrerouting, "", ruleNotTrackSport...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableRaw, ChainRawNotTrackDNSOutput, "", ruleNotTrackSport...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleNotTrackDport := []string{"-p", "UDP", "-m", "udp", "--dport", "53", "-j", "CT", "--notrack"}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableRaw, ChainRawNotTrackDNSPrerouting, "", ruleNotTrackDport...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableRaw, ChainRawNotTrackDNSOutput, "", ruleNotTrackDport...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set jump rule
		ruleJumpPre := []string{"-j", ChainRawNotTrackDNSPrerouting}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableRaw, ChainBasePrerouting, "", ruleJumpPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpOut := []string{"-j", ChainRawNotTrackDNSOutput}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableRaw, ChainBaseOutput, "", ruleJumpOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	// IPv6
	if configIPv6Enabled {
		// Create chains
		out, err := iptables.CreateChainIPv6(iptables.TableRaw, ChainRawNotTrackDNSPrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv6(iptables.TableRaw, ChainRawNotTrackDNSOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set not track rules
		ruleNotTrackSport := []string{"-p", "UDP", "-m", "udp", "--sport", "53", "-j", "CT", "--notrack"}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableRaw, ChainRawNotTrackDNSPrerouting, "", ruleNotTrackSport...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableRaw, ChainRawNotTrackDNSOutput, "", ruleNotTrackSport...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleNotTrackDport := []string{"-p", "UDP", "-m", "udp", "--dport", "53", "-j", "CT", "--notrack"}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableRaw, ChainRawNotTrackDNSPrerouting, "", ruleNotTrackDport...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableRaw, ChainRawNotTrackDNSOutput, "", ruleNotTrackDport...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set jump rule
		ruleJumpPre := []string{"-j", ChainRawNotTrackDNSPrerouting}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableRaw, ChainBasePrerouting, "", ruleJumpPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpOut := []string{"-j", ChainRawNotTrackDNSOutput}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableRaw, ChainBaseOutput, "", ruleJumpOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	return nil
}

func deleteRulesNotTrackDNS(logger logr.Logger) error {
	// IPv4
	if configIPv4Enabled {
		// Delete jump rule
		ruleJumpPre := []string{"-j", ChainRawNotTrackDNSPrerouting}
		out, err := iptables.DeleteRuleIPv4(iptables.TableRaw, ChainBasePrerouting, "", ruleJumpPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpOut := []string{"-j", ChainRawNotTrackDNSOutput}
		out, err = iptables.DeleteRuleIPv4(iptables.TableRaw, ChainBaseOutput, "", ruleJumpOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Delete chain
		out, err = iptables.DeleteChainIPv4(iptables.TableRaw, ChainRawNotTrackDNSPrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.DeleteChainIPv4(iptables.TableRaw, ChainRawNotTrackDNSOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	// IPv6
	if configIPv6Enabled {
		// Delete jump rule
		ruleJumpPre := []string{"-j", ChainRawNotTrackDNSPrerouting}
		out, err := iptables.DeleteRuleIPv6(iptables.TableRaw, ChainBasePrerouting, "", ruleJumpPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpOut := []string{"-j", ChainRawNotTrackDNSOutput}
		out, err = iptables.DeleteRuleIPv6(iptables.TableRaw, ChainBaseOutput, "", ruleJumpOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Delete chain
		out, err = iptables.DeleteChainIPv6(iptables.TableRaw, ChainRawNotTrackDNSPrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.DeleteChainIPv6(iptables.TableRaw, ChainRawNotTrackDNSOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	return nil
}
