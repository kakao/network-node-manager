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
	iptablesCmd       = "iptables"
	iptablesErrNoRule = "No chain/target/match by that name"

	TableNAT    Table = "nat"
	TableFilter Table = "filter"
)

// Var
var (
	lock = &sync.Mutex{}
)

func CreateChain(table Table, chain string) (string, error) {
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

func CreateRuleFirst(table Table, chain string, comment string, rule ...string) (string, error) {
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
	cmd = exec.Command(iptablesCmd, append(append(args, "-I", chain, "1"), rule...)...)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}

func CreateRuleLast(table Table, chain string, comment string, rule ...string) (string, error) {
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

func DeleteRule(table Table, chain string, comment string, rule ...string) (string, error) {
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
