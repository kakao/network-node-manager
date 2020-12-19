package rules

import (
	"github.com/go-logr/logr"

	"github.com/kakao/network-node-manager/pkg/ip"
	"github.com/kakao/network-node-manager/pkg/iptables"
)

func InitRulesNotTrackDNS(logger logr.Logger) error {
	// Init base chains
	if err := initBaseChains(logger); err != nil {
		logger.Error(err, "failed to init base chain for externalIP to clusterIP Rules")
		return err
	}

	// IPv4
	if ip.IsIPv4CIDR(podCIDRIPv4) {
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
		ruleNotTrackSrcPodCidr := []string{"-p", "UDP", "-m", "udp", "-s", podCIDRIPv4, "--sport", "53", "-j", "CT", "--notrack"}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableRaw, ChainRawNotTrackDNSPrerouting, "", ruleNotTrackSrcPodCidr...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableRaw, ChainRawNotTrackDNSOutput, "", ruleNotTrackSrcPodCidr...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleNotTrackDestPodCidr := []string{"-p", "UDP", "-m", "udp", "-d", podCIDRIPv4, "--dport", "53", "-j", "CT", "--notrack"}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableRaw, ChainRawNotTrackDNSPrerouting, "", ruleNotTrackDestPodCidr...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableRaw, ChainRawNotTrackDNSOutput, "", ruleNotTrackDestPodCidr...)
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
	if ip.IsIPv6CIDR(podCIDRIPv6) {
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
		ruleNotTrackSrcPodCidr := []string{"-p", "UDP", "-m", "udp", "-s", podCIDRIPv6, "--sport", "53", "-j", "CT", "--notrack"}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableRaw, ChainRawNotTrackDNSPrerouting, "", ruleNotTrackSrcPodCidr...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableRaw, ChainRawNotTrackDNSOutput, "", ruleNotTrackSrcPodCidr...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleNotTrackDestPodCidr := []string{"-p", "UDP", "-m", "udp", "-d", podCIDRIPv6, "--dport", "53", "-j", "CT", "--notrack"}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableRaw, ChainRawNotTrackDNSPrerouting, "", ruleNotTrackDestPodCidr...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableRaw, ChainRawNotTrackDNSOutput, "", ruleNotTrackDestPodCidr...)
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

func CleanupRulesNotTrackDNS(logger logr.Logger) error {
	// IPv4
	if ip.IsIPv4CIDR(podCIDRIPv4) {
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
	if ip.IsIPv6CIDR(podCIDRIPv6) {
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

func CreateRulesNotTrackDNS(logger logr.Logger, dnsClusterIP string) error {
	// Init base chains
	if err := initBaseChains(logger); err != nil {
		logger.Error(err, "failed to init base chain for externalIP to clusterIP Rules")
		return err
	}

	// IPv4
	if ip.IsIPv4CIDR(podCIDRIPv4) && ip.IsIPv4Addr(dnsClusterIP) {
		// Set not track rules
		ruleNotTrackDport := []string{"-p", "UDP", "-m", "udp", "-d", dnsClusterIP, "--dport", "53", "-j", "CT", "--notrack"}
		out, err := iptables.CreateRuleFirstIPv4(iptables.TableRaw, ChainRawNotTrackDNSPrerouting, "", ruleNotTrackDport...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableRaw, ChainRawNotTrackDNSOutput, "", ruleNotTrackDport...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}
	// IPv6
	if ip.IsIPv6CIDR(podCIDRIPv6) && ip.IsIPv6Addr(dnsClusterIP) {
		// Set not track rules
		ruleNotTrackDport := []string{"-p", "UDP", "-m", "udp", "-d", dnsClusterIP, "--dport", "53", "-j", "CT", "--notrack"}
		out, err := iptables.CreateRuleFirstIPv6(iptables.TableRaw, ChainRawNotTrackDNSPrerouting, "", ruleNotTrackDport...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableRaw, ChainRawNotTrackDNSOutput, "", ruleNotTrackDport...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	return nil
}
