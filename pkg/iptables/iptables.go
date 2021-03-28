// Tested with iptables version 1.8.3-r2 in alpine
package iptables

import (
	"os/exec"
	"strings"
	"sync"
)

// Type
type Table string

// Const
const (
	iptablesCmdIPv4     = "iptables"
	iptablesCmdIPv6     = "ip6tables"
	iptablesSaveCmdIPv4 = "iptables-save"
	iptablesSaveCmdIPv6 = "ip6tables-save"

	// kube-proxy defaults
	iptablesWaitSeconds          = "5"
	iptablesWaitIntervalUseconds = "100000"

	iptablesErrNoRule   = "No chain/target/match by that name"
	iptablesErrNoTarget = "Couldn't load target"

	TableNAT    Table = "nat"
	TableFilter Table = "filter"
	TableRaw    Table = "raw"
)

// Var
var (
	lock = &sync.Mutex{}
)

// IsExistChain
func IsExistChainIPv4(table Table, chain string) bool {
	return isExistChain(iptablesCmdIPv4, table, chain)
}

func IsExistChainIPv6(table Table, chain string) bool {
	return isExistChain(iptablesCmdIPv6, table, chain)
}

func isExistChain(iptablesCmd string, table Table, chain string) bool {
	// Lock
	lock.Lock()
	defer lock.Unlock()

	// Check chain
	_, err := runIptables(iptablesCmd, table, "-L", chain)
	if err != nil {
		return false
	}
	return true
}

// CreateChain
func CreateChainIPv4(table Table, chain string) (string, error) {
	return createChain(iptablesCmdIPv4, table, chain)
}

func CreateChainIPv6(table Table, chain string) (string, error) {
	return createChain(iptablesCmdIPv6, table, chain)
}

func createChain(iptablesCmd string, table Table, chain string) (string, error) {
	// Lock
	lock.Lock()
	defer lock.Unlock()

	// Check chain
	out, err := runIptables(iptablesCmd, table, "-L", chain)
	if err == nil {
		// If already exists, return success
		return string(out), nil
	}

	// Create chain
	out, err = runIptables(iptablesCmd, table, "-N", chain)
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}

// DeleteChain
func DeleteChainIPv4(table Table, chain string) (string, error) {
	return deleteChain(iptablesCmdIPv4, table, chain)
}

func DeleteChainIPv6(table Table, chain string) (string, error) {
	return deleteChain(iptablesCmdIPv6, table, chain)
}

func deleteChain(iptablesCmd string, table Table, chain string) (string, error) {
	// Lock
	lock.Lock()
	defer lock.Unlock()

	// Check chain
	out, err := runIptables(iptablesCmd, table, "-L", chain)
	if err != nil {
		// If chain isn't exist, return success
		return string(out), nil
	}

	// Flush chain
	out, err = runIptables(iptablesCmd, table, "-F", chain)
	if err != nil {
		return string(out), err
	}

	// Delete chain
	out, err = runIptables(iptablesCmd, table, "-X", chain)
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}

// IsExistRule
func IsExistRuleIPv4(table Table, chain string, comment string, rule ...string) bool {
	return isExistRule(iptablesCmdIPv4, table, chain, comment, rule...)
}

func IsExistRuleIPv6(table Table, chain string, comment string, rule ...string) bool {
	return isExistRule(iptablesCmdIPv6, table, chain, comment, rule...)
}

func isExistRule(iptablesCmd string, table Table, chain string, comment string, rule ...string) bool {
	// Lock
	lock.Lock()
	defer lock.Unlock()

	// Set Common args
	args := []string{"-C", chain}
	if comment != "" {
		args = append(args, "-m", "comment", "--comment", comment)
	}

	// Check rule
	_, err := runIptables(iptablesCmd, table, append(args, rule...)...)
	if err != nil {
		return false
	}
	return true
}

// GetRules
func GetRulesIPv4(table Table, chain string) ([]string, error) {
	return getRules(iptablesSaveCmdIPv4, table, chain)
}

func GetRulesIPv6(table Table, chain string) ([]string, error) {
	return getRules(iptablesSaveCmdIPv6, table, chain)
}

func getRules(iptablesSaveCmd string, table Table, chain string) ([]string, error) {
	// Lock
	lock.Lock()
	defer lock.Unlock()

	// Set Common args
	args := []string{"-t", string(table)}

	// Check rule
	cmd := exec.Command(iptablesSaveCmd, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var result []string
	for _, rule := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(rule, "-A "+chain) {
			result = append(result, rule)
		}
	}

	return result, nil
}

// CreateRuleFirst
func CreateRuleFirstIPv4(table Table, chain string, comment string, rule ...string) (string, error) {
	return createRuleFirst(iptablesCmdIPv4, table, chain, comment, rule...)
}

func CreateRuleFirstIPv6(table Table, chain string, comment string, rule ...string) (string, error) {
	return createRuleFirst(iptablesCmdIPv6, table, chain, comment, rule...)
}

func createRuleFirst(iptablesCmd string, table Table, chain string, comment string, rule ...string) (string, error) {
	// Lock
	lock.Lock()
	defer lock.Unlock()

	// Set Common args
	args := []string{}
	if comment != "" {
		args = append(args, "-m", "comment", "--comment", comment)
	}

	// Check rule
	out, err := runIptables(iptablesCmd, table, append(append(args, "-C", chain), rule...)...)
	if err == nil { // If already exists, return success
		return string(out), nil
	}

	// Create rule
	out, err = runIptables(iptablesCmd, table, append(append(args, "-I", chain, "1"), rule...)...)
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}

// CreateRuleLast
func CreateRuleLastIPv4(table Table, chain string, comment string, rule ...string) (string, error) {
	return createRuleLast(iptablesCmdIPv4, table, chain, comment, rule...)
}

func CreateRuleLastIPv6(table Table, chain string, comment string, rule ...string) (string, error) {
	return createRuleLast(iptablesCmdIPv6, table, chain, comment, rule...)
}

func createRuleLast(iptablesCmd string, table Table, chain string, comment string, rule ...string) (string, error) {
	// Lock
	lock.Lock()
	defer lock.Unlock()

	// Set Common args
	args := []string{}
	if comment != "" {
		args = append(args, "-m", "comment", "--comment", comment)
	}

	// Check rule
	out, err := runIptables(iptablesCmd, table, append(append(args, "-C", chain), rule...)...)
	if err == nil {
		// If already exists, return success
		return string(out), nil
	}

	// Create rule
	out, err = runIptables(iptablesCmd, table, append(append(args, "-A", chain), rule...)...)
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}

// DeleteRule
func DeleteRuleIPv4(table Table, chain string, comment string, rule ...string) (string, error) {
	return deleteRule(iptablesCmdIPv4, table, chain, comment, rule...)
}

func DeleteRuleIPv6(table Table, chain string, comment string, rule ...string) (string, error) {
	return deleteRule(iptablesCmdIPv6, table, chain, comment, rule...)
}

func deleteRule(iptablesCmd string, table Table, chain string, comment string, rule ...string) (string, error) {
	// Lock
	lock.Lock()
	defer lock.Unlock()

	// Set Common args
	args := []string{}
	if comment != "" {
		args = append(args, "-m", "comment", "--comment", comment)
	}

	// Check rule
	out, err := runIptables(iptablesCmd, table, append(append(args, "-C", chain), rule...)...)
	if err != nil {
		if strings.Contains(string(out), iptablesErrNoRule) {
			// If rule isn't exist, return success
			return string(out), nil
		} else if strings.Contains(string(out), iptablesErrNoTarget) {
			// If target isn't exit, return success
			return string(out), nil
		}
		return string(out), err
	}

	// Delete rule
	out, err = runIptables(iptablesCmd, table, append(append(args, "-D", chain), rule...)...)
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}

// DeleteRuleRaw
func DeleteRuleRawIPv4(table Table, rule ...string) (string, error) {
	return deleteRuleRaw(iptablesCmdIPv4, table, rule...)
}

func DeleteRuleRawIPv6(table Table, rule ...string) (string, error) {
	return deleteRuleRaw(iptablesCmdIPv6, table, rule...)
}

func deleteRuleRaw(iptablesCmd string, table Table, rule ...string) (string, error) {
	// Lock
	lock.Lock()
	defer lock.Unlock()

	// Check rule
	out, err := runIptables(iptablesCmd, table, append([]string{"-C"}, rule...)...)
	if err != nil {
		if strings.Contains(string(out), iptablesErrNoRule) {
			// If rule isn't exist, return success
			return string(out), nil
		}
		return string(out), err
	}

	// Delete rule
	out, err = runIptables(iptablesCmd, table, append([]string{"-D"}, rule...)...)
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}

// Run iptables with lock
func runIptables(iptablesCmd string, table Table, args ...string) ([]byte, error) {
	// Build arguments list
	fullArgs := []string{
		"-w", iptablesWaitSeconds,
		"-W", iptablesWaitIntervalUseconds,
		"-t", string(table),
	}
	fullArgs = append(fullArgs, args...)

	// Check chain
	cmd := exec.Command(iptablesCmd, fullArgs...)
	out, err := cmd.CombinedOutput()

	return out, err
}
