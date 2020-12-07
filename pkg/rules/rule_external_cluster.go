package rules

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"
	"github.com/kakao/network-node-manager/pkg/ip"
	"github.com/kakao/network-node-manager/pkg/iptables"
)

func InitRulesExternalCluster(logger logr.Logger) error {
	// Init base chains
	if err := initBaseChains(logger); err != nil {
		logger.Error(err, "failed to init base chain for externalIP to clusterIP Rules")
		return err
	}

	// IPv4
	if configIPv4Enabled {
		// Create chain in nat table
		out, err := iptables.CreateChainIPv4(iptables.TableNAT, ChainNATExternalClusterPrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv4(iptables.TableNAT, ChainNATExternalClusterOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set jump rule to each chain in nat table
		ruleJumpPre := []string{"-j", ChainNATExternalClusterPrerouting}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableNAT, ChainBasePrerouting, "", ruleJumpPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpOut := []string{"-j", ChainNATExternalClusterOutput}
		out, err = iptables.CreateRuleFirstIPv4(iptables.TableNAT, ChainBaseOutput, "", ruleJumpOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}
	// IPv6
	if configIPv6Enabled {
		// Create chain in nat table
		out, err := iptables.CreateChainIPv6(iptables.TableNAT, ChainNATExternalClusterPrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.CreateChainIPv6(iptables.TableNAT, ChainNATExternalClusterOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set jump rule to each chain in nat table
		ruleJumpPre := []string{"-j", ChainNATExternalClusterPrerouting}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableNAT, ChainBasePrerouting, "", ruleJumpPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpOut := []string{"-j", ChainNATExternalClusterOutput}
		out, err = iptables.CreateRuleFirstIPv6(iptables.TableNAT, ChainBaseOutput, "", ruleJumpOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	return nil
}

func DestoryRulesExternalCluster(logger logr.Logger) error {
	// IPv4
	if configIPv4Enabled {
		// Delete jump rule to each chain in nat table
		ruleJumpPre := []string{"-j", ChainNATExternalClusterPrerouting}
		out, err := iptables.DeleteRuleIPv4(iptables.TableNAT, ChainBasePrerouting, "", ruleJumpPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpOut := []string{"-j", ChainNATExternalClusterOutput}
		out, err = iptables.DeleteRuleIPv4(iptables.TableNAT, ChainBaseOutput, "", ruleJumpOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Delete chain in nat table
		out, err = iptables.DeleteChainIPv4(iptables.TableNAT, ChainNATExternalClusterPrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.DeleteChainIPv4(iptables.TableNAT, ChainNATExternalClusterOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}
	// IPv6
	if configIPv6Enabled {
		// Delete jump rule to each chain in nat table
		ruleJumpPre := []string{"-j", ChainNATExternalClusterPrerouting}
		out, err := iptables.DeleteRuleIPv6(iptables.TableNAT, ChainBasePrerouting, "", ruleJumpPre...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleJumpOut := []string{"-j", ChainNATExternalClusterOutput}
		out, err = iptables.DeleteRuleIPv6(iptables.TableNAT, ChainBaseOutput, "", ruleJumpOut...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Delete chain in nat table
		out, err = iptables.DeleteChainIPv6(iptables.TableNAT, ChainNATExternalClusterPrerouting)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		out, err = iptables.DeleteChainIPv6(iptables.TableNAT, ChainNATExternalClusterOutput)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	}

	return nil
}

func CleanupRulesExternalCluster(logger logr.Logger, svcs *corev1.ServiceList, podCIDRIPv4, podCIDRIPv6 string) error {
	// IPv4
	if configIPv4Enabled {
		// Make up service map
		svcMap := make(map[string]*corev1.Service)
		for _, svc := range svcs.Items {
			if ip.IsIPv4Addr(svc.Spec.ClusterIP) && svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
				svcMap[svc.Namespace+"/"+svc.Name] = svc.DeepCopy()
			}
		}

		// Cleanup prerouting chain
		preRules, err := iptables.GetRulesIPv4(iptables.TableNAT, ChainNATExternalClusterPrerouting)
		if err != nil {
			return err
		}
		for _, rule := range preRules {
			// Get service info from rule and k8s, and delete iptables rules
			nsName, src, dest, jump, dnatDest := getSvcInfoFromRule(rule)
			svc, ok := svcMap[nsName]
			if !ok {
				logger.WithValues("rule", rule).Info("there is no service info in k8s. cleanup prerouting chain IPv4 rule")
				out, err := iptables.DeleteRuleRawIPv4(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				if err != nil {
					logger.Error(err, out)
					return err
				}
				continue
			}

			// Compare service info and delete iptables rules
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				externalIP := ingress.IP + "/32"
				if (jump == ChainNATKubeMarkMasq && (src == podCIDRIPv4 && dest == externalIP)) ||
					(jump == "DNAT" && (src == podCIDRIPv4 && dest == externalIP && dnatDest == svc.Spec.ClusterIP)) {
					continue
				}
				logger.WithValues("rule", rule).Info("service info is diff. cleanup prerouting chain IPv4 rule")
				out, err := iptables.DeleteRuleRawIPv4(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				if err != nil {
					logger.Error(err, out)
					return err
				}
			}
		}

		// Cleanup output chain
		outRules, err := iptables.GetRulesIPv4(iptables.TableNAT, ChainNATExternalClusterOutput)
		if err != nil {
			return err
		}
		for _, rule := range outRules {
			// Get service info from rule and k8s, and delete iptables rules
			nsName, _, dest, jump, dnatDest := getSvcInfoFromRule(rule)
			svc, ok := svcMap[nsName]
			if !ok {
				logger.WithValues("rule", rule).Info("there is no service info in k8s. cleanup output chain IPv4 rule")
				out, err := iptables.DeleteRuleRawIPv4(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				if err != nil {
					logger.Error(err, out)
					return err
				}
				continue
			}

			// Compare service info and delete diff iptables rules
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				externalIP := ingress.IP + "/32"
				if (jump == ChainNATKubeMarkMasq && dest == externalIP) ||
					(jump == "DNAT" && (dest == externalIP && dnatDest == svc.Spec.ClusterIP)) {
					continue
				}
				logger.WithValues("rule", rule).Info("service info is diff. cleanup output chain IPv4 rule")
				out, err := iptables.DeleteRuleRawIPv4(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				if err != nil {
					logger.Error(err, out)
					return err
				}
			}
		}
	}
	// IPv6
	if configIPv6Enabled {
		// Make up service map
		svcMap := make(map[string]*corev1.Service)
		for _, svc := range svcs.Items {
			if ip.IsIPv6Addr(svc.Spec.ClusterIP) && svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
				svcMap[svc.Namespace+"/"+svc.Name] = svc.DeepCopy()
			}
		}

		// Cleanup prerouting chain
		preRules, err := iptables.GetRulesIPv6(iptables.TableNAT, ChainNATExternalClusterPrerouting)
		if err != nil {
			return err
		}
		for _, rule := range preRules {
			// Get service info from rule and k8s, and delete iptables rules
			nsName, src, dest, jump, dnatDest := getSvcInfoFromRule(rule)
			svc, ok := svcMap[nsName]
			if !ok {
				logger.WithValues("rule", rule).Info("there is no service info in k8s. cleanup prerouting chain IPv6 rule")
				out, err := iptables.DeleteRuleRawIPv6(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				if err != nil {
					logger.Error(err, out)
					return err
				}
				continue
			}

			// Compare service info and delete iptables rules
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				externalIP := ingress.IP + "/128"
				if (jump == ChainNATKubeMarkMasq && (src == podCIDRIPv6 && dest == externalIP)) ||
					(jump == "DNAT" && (src == podCIDRIPv6 && dest == externalIP && dnatDest == svc.Spec.ClusterIP)) {
					continue
				}
				logger.WithValues("rule", rule).Info("service info is diff. cleanup prerouting chain IPv6 rule")
				out, err := iptables.DeleteRuleRawIPv6(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				if err != nil {
					logger.Error(err, out)
					return err
				}
			}
		}

		// Cleanup output
		outRules, err := iptables.GetRulesIPv6(iptables.TableNAT, ChainNATExternalClusterOutput)
		if err != nil {
			return err
		}
		for _, rule := range outRules {
			// Get service info from rule and k8s, and delete iptables rules
			nsName, _, dest, jump, dnatDest := getSvcInfoFromRule(rule)
			svc, ok := svcMap[nsName]
			if !ok {
				logger.WithValues("rule", rule).Info("there is no service info in k8s. cleanup output chain IPv6 rule")
				out, err := iptables.DeleteRuleRawIPv6(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				if err != nil {
					logger.Error(err, out)
					return err
				}
				continue
			}

			// Compare service info and delete diff iptables rules
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				externalIP := ingress.IP + "/128"
				if (jump == ChainNATKubeMarkMasq && dest == externalIP) ||
					(jump == "DNAT" && (dest == externalIP && dnatDest == svc.Spec.ClusterIP)) {
					continue
				}
				logger.WithValues("rule", rule).Info("service info is diff. cleanup output chain IPv6 rule")
				out, err := iptables.DeleteRuleRawIPv6(iptables.TableNAT, iptables.ChangeRuleToDelete(rule)...)
				if err != nil {
					logger.Error(err, out)
					return err
				}
			}
		}
	}

	return nil
}

func CreateRulesExternalCluster(logger logr.Logger, req *ctrl.Request, clusterIP, externalIP, podCIDRIPv4, podCIDRIPv6 string) error {
	// Don't use spec.ipFamily to distingush between IPv4 and IPv6 Address
	// for kubernetes version that dosen't support IPv6 dualstack
	if configIPv4Enabled && ip.IsIPv4Addr(clusterIP) {
		// IPv4
		// Set prerouting
		rulePreMasq := []string{"-s", podCIDRIPv4, "-d", externalIP, "-j", ChainNATKubeMarkMasq}
		out, err := iptables.CreateRuleLastIPv4(iptables.TableNAT, ChainNATExternalClusterPrerouting, req.String(), rulePreMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		rulePreDNAT := []string{"-s", podCIDRIPv4, "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.CreateRuleLastIPv4(iptables.TableNAT, ChainNATExternalClusterPrerouting, req.String(), rulePreDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set output
		ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", ChainNATKubeMarkMasq}
		out, err = iptables.CreateRuleLastIPv4(iptables.TableNAT, ChainNATExternalClusterOutput, req.String(), ruleOutMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.CreateRuleLastIPv4(iptables.TableNAT, ChainNATExternalClusterOutput, req.String(), ruleOutDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	} else if configIPv6Enabled && ip.IsIPv6Addr(clusterIP) {
		// IPv6
		// Set prerouting
		rulePreMasq := []string{"-s", podCIDRIPv6, "-d", externalIP, "-j", ChainNATKubeMarkMasq}
		out, err := iptables.CreateRuleLastIPv6(iptables.TableNAT, ChainNATExternalClusterPrerouting, req.String(), rulePreMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		rulePreDNAT := []string{"-s", podCIDRIPv6, "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.CreateRuleLastIPv6(iptables.TableNAT, ChainNATExternalClusterPrerouting, req.String(), rulePreDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Set output
		ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", ChainNATKubeMarkMasq}
		out, err = iptables.CreateRuleLastIPv6(iptables.TableNAT, ChainNATExternalClusterOutput, req.String(), ruleOutMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.CreateRuleLastIPv6(iptables.TableNAT, ChainNATExternalClusterOutput, req.String(), ruleOutDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	} else {
		if ip.IsVaildIP(clusterIP) {
			logger.WithValues("clusterIP", clusterIP).Error(errors.New("invalid IP"), "invaild IP")
		}
	}

	return nil
}

func DeleteRulesExternalCluster(logger logr.Logger, req *ctrl.Request, clusterIP, externalIP, podCIDRIPv4, podCIDRIPv6 string) error {
	// Don't use spec.ipFamily to distingush between IPv4 and IPv6 Address
	// for kubernetes version that dosen't support IPv6 dualstack
	if configIPv4Enabled && ip.IsIPv4Addr(clusterIP) {
		// IPv4
		// Unset prerouting
		rulePreMasq := []string{"-s", podCIDRIPv4, "-d", externalIP, "-j", ChainNATKubeMarkMasq}
		out, err := iptables.DeleteRuleIPv4(iptables.TableNAT, ChainNATExternalClusterPrerouting, req.String(), rulePreMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		rulePreDNAT := []string{"-s", podCIDRIPv4, "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.DeleteRuleIPv4(iptables.TableNAT, ChainNATExternalClusterPrerouting, req.String(), rulePreDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Unset output
		ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", ChainNATKubeMarkMasq}
		out, err = iptables.DeleteRuleIPv4(iptables.TableNAT, ChainNATExternalClusterOutput, req.String(), ruleOutMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.DeleteRuleIPv4(iptables.TableNAT, ChainNATExternalClusterOutput, req.String(), ruleOutDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	} else if configIPv6Enabled && ip.IsIPv6Addr(clusterIP) {
		// IPv6
		// Unset prerouting
		rulePreMasq := []string{"-s", podCIDRIPv6, "-d", externalIP, "-j", ChainNATKubeMarkMasq}
		out, err := iptables.DeleteRuleIPv6(iptables.TableNAT, ChainNATExternalClusterPrerouting, req.String(), rulePreMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		rulePreDNAT := []string{"-s", podCIDRIPv6, "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.DeleteRuleIPv6(iptables.TableNAT, ChainNATExternalClusterPrerouting, req.String(), rulePreDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}

		// Unset output
		ruleOutMasq := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", ChainNATKubeMarkMasq}
		out, err = iptables.DeleteRuleIPv6(iptables.TableNAT, ChainNATExternalClusterOutput, req.String(), ruleOutMasq...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
		ruleOutDNAT := []string{"-m", "addrtype", "--src-type", "LOCAL", "-d", externalIP, "-j", "DNAT", "--to-destination", clusterIP}
		out, err = iptables.DeleteRuleIPv6(iptables.TableNAT, ChainNATExternalClusterOutput, req.String(), ruleOutDNAT...)
		if err != nil {
			logger.Error(err, out)
			return err
		}
	} else {
		if ip.IsVaildIP(clusterIP) {
			logger.WithValues("clusterIP", clusterIP).Error(errors.New("invalid IP"), "invaild IP")
		}
	}
	return nil
}

func getSvcInfoFromRule(rule string) (nsName, src, dest, jump, dnatDest string) {
	nsName = iptables.GetRuleComment(rule)
	src = iptables.GetRuleSrc(rule)
	dest = iptables.GetRuleDest(rule)
	jump = iptables.GetRuleJump(rule)
	dnatDest = iptables.GetRuleDNATDest(rule)
	return
}
