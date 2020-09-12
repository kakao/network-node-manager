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
	iptablesCmdIPv4   = "iptables"
	iptablesCmdIPv6   = "ip6tables"
	iptablesErrNoRule = "No chain/target/match by that name"

	TableNAT    Table = "nat"
	TableFilter Table = "filter"
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

	// Set Common args
	args := []string{"-t", string(table)}

	// Check chain
	cmd := exec.Command(iptablesCmd, append(args, "-L", chain)...)
	_, err := cmd.CombinedOutput()
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

	// Set Common args
	args := []string{"-t", string(table)}

	// Check chain
	cmd := exec.Command(iptablesCmd, append(args, "-L", chain)...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		// If already exists, return success
		return string(out), nil
	}

	// Create chain
	cmd = exec.Command(iptablesCmd, append(args, "-N", chain)...)
	out, err = cmd.CombinedOutput()
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

	// Set Common args
	args := []string{"-t", string(table)}

	// Check chain
	cmd := exec.Command(iptablesCmd, append(args, "-L", chain)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// If chain isn't exist, return success
		return string(out), nil
	}

	// Delete chain
	cmd = exec.Command(iptablesCmd, append(args, "-X", chain)...)
	out, err = cmd.CombinedOutput()
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
	args := []string{"-t", string(table)}
	if comment != "" {
		args = append(args, "-m", "comment", "--comment", comment)
	}

	// Check rule
	cmd := exec.Command(iptablesCmd, append(append(args, "-C", chain), rule...)...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return true
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
	args := []string{"-t", string(table)}
	if comment != "" {
		args = append(args, "-m", "comment", "--comment", comment)
	}

	// Check rule
	cmd := exec.Command(iptablesCmd, append(append(args, "-C", chain), rule...)...)
	out, err := cmd.CombinedOutput()
	if err == nil { // If already exists, return success
		return string(out), nil
	}

	// Create rule
	cmd = exec.Command(iptablesCmd, append(append(args, "-I", chain, "1"), rule...)...)
	out, err = cmd.CombinedOutput()
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
	args := []string{"-t", string(table)}
	if comment != "" {
		args = append(args, "-m", "comment", "--comment", comment)
	}

	// Check rule
	cmd := exec.Command(iptablesCmd, append(append(args, "-C", chain), rule...)...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		// If already exists, return success
		return string(out), nil
	}

	// Create rule
	cmd = exec.Command(iptablesCmd, append(append(args, "-A", chain), rule...)...)
	out, err = cmd.CombinedOutput()
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
	args := []string{"-t", string(table)}
	if comment != "" {
		args = append(args, "-m", "comment", "--comment", comment)
	}

	// Check rule
	cmd := exec.Command(iptablesCmd, append(append(args, "-C", chain), rule...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), iptablesErrNoRule) {
			// If rule isn't exist, return success
			return string(out), nil
		}
		return string(out), err
	}

	// Delete rule
	cmd = exec.Command(iptablesCmd, append(append(args, "-D", chain), rule...)...)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}
