package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpull/internal/audit"
)

func init() {
	policySaveCmd := &cobra.Command{
		Use:   "policy-save [name] [path:allow|deny] ...",
		Short: "Save a path policy",
		Args:  cobra.MinimumNArgs(2),
		RunE:  runPolicySave,
	}

	policyCheckCmd := &cobra.Command{
		Use:   "policy-check [name] [path ...]",
		Short: "Check paths against a saved policy",
		Args:  cobra.MinimumNArgs(2),
		RunE:  runPolicyCheck,
	}

	rootCmd.AddCommand(policySaveCmd)
	rootCmd.AddCommand(policyCheckCmd)
}

func runPolicySave(cmd *cobra.Command, args []string) error {
	name := args[0]
	var rules []audit.PolicyRule
	for _, arg := range args[1:] {
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid rule %q: expected path:allow or path:deny", arg)
		}
		path, action := parts[0], strings.ToLower(parts[1])
		allowed := action == "allow"
		if action != "allow" && action != "deny" {
			return fmt.Errorf("invalid action %q: must be allow or deny", action)
		}
		rules = append(rules, audit.PolicyRule{Path: path, Allowed: allowed})
	}
	dir, _ := os.Getwd()
	if err := audit.SavePolicy(dir, name, rules); err != nil {
		return fmt.Errorf("saving policy: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Policy %q saved with %d rule(s).\n", name, len(rules))
	return nil
}

func runPolicyCheck(cmd *cobra.Command, args []string) error {
	name := args[0]
	paths := args[1:]
	dir, _ := os.Getwd()
	p, err := audit.LoadPolicy(dir, name)
	if err != nil {
		return fmt.Errorf("loading policy %q: %w", name, err)
	}
	violations := audit.EnforcePolicy(p, paths)
	if len(violations) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "All paths comply with policy %q.\n", name)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%d violation(s) found:\n", len(violations))
	for _, v := range violations {
		fmt.Fprintf(cmd.OutOrStdout(), "  DENIED  %s  (%s)\n", v.Path, v.Reason)
	}
	return fmt.Errorf("policy check failed: %d violation(s)", len(violations))
}
