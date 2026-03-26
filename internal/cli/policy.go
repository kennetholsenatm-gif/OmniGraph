package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/policy"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage and evaluate policies",
		Long: `Policy-as-Code commands for governance and compliance.

Policies are defined using Rego (Open Policy Agent) and can be used to:
- Validate infrastructure configurations
- Enforce security best practices
- Check compliance requirements
- Gate pipeline execution based on policy violations

Examples:
  # Check a configuration against policies
  omnigraph policy check config.yaml --policy-dir ./policies

  # Dry-run policy evaluation
  omnigraph policy dry-run --policy security-baseline --input terraform.tfstate

  # Generate policy report
  omnigraph policy report --format json --output policy-report.json`,
	}

	cmd.AddCommand(newPolicyCheckCmd())
	cmd.AddCommand(newPolicyDryRunCmd())
	cmd.AddCommand(newPolicyReportCmd())
	cmd.AddCommand(newPolicyListCmd())
	cmd.AddCommand(newPolicyValidateCmd())

	return cmd
}

func newPolicyCheckCmd() *cobra.Command {
	var policyDir string
	var policySet string
	var enforcement string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "check [file...]",
		Short: "Check configuration against policies",
		Long: `Check one or more configuration files against loaded policies.

The check command evaluates configurations against policy sets and reports
any violations found. By default, violations cause a non-zero exit code.

Examples:
  # Check a single file
  omnigraph policy check config.yaml --policy-dir ./policies

  # Check multiple files
  omnigraph policy check *.yaml --policy-dir ./policies

  # Check with specific policy set
  omnigraph policy check config.yaml --policy security-baseline

  # Warn on violations instead of failing
  omnigraph policy check config.yaml --enforcement warn`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPolicyCheck(cmd, args, policyDir, policySet, enforcement, outputFormat)
		},
	}

	cmd.Flags().StringVar(&policyDir, "policy-dir", "", "Directory containing policy files")
	cmd.Flags().StringVar(&policySet, "policy", "", "Specific policy set to use")
	cmd.Flags().StringVar(&enforcement, "enforcement", "", "Enforcement level (warn, deny)")
	cmd.Flags().StringVar(&outputFormat, "output", "text", "Output format (text, json, yaml)")

	return cmd
}

func runPolicyCheck(cmd *cobra.Command, args []string, policyDir, policySet, enforcement, outputFormat string) error {
	ctx := context.Background()
	engine := policy.NewEngine()

	// Load policies
	if policyDir != "" {
		if err := loadPoliciesFromDir(engine, policyDir); err != nil {
			return fmt.Errorf("failed to load policies: %w", err)
		}
	}

	// Get policy set name
	psName := policySet
	if psName == "" {
		sets := engine.ListPolicySets()
		if len(sets) == 0 {
			return fmt.Errorf("no policy sets loaded")
		}
		psName = sets[0].Metadata.Name
	}

	// Override enforcement if specified
	if enforcement != "" {
		ps, err := engine.GetPolicySet(psName)
		if err != nil {
			return err
		}
		ps.Spec.Enforcement = enforcement
	}

	// Check each file
	allPassed := true
	var allReports []*policy.PolicyReport

	for _, file := range args {
		report, err := engine.EvaluateFile(ctx, psName, file)
		if err != nil {
			return fmt.Errorf("failed to evaluate %s: %w", file, err)
		}

		allReports = append(allReports, report)

		if report.Failed > 0 {
			allPassed = false
		}
	}

	// Output results
	switch outputFormat {
	case "json":
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		if len(allReports) == 1 {
			return enc.Encode(allReports[0])
		}
		return enc.Encode(allReports)

	case "yaml":
		enc := yaml.NewEncoder(cmd.OutOrStdout())
		if len(allReports) == 1 {
			return enc.Encode(allReports[0])
		}
		return enc.Encode(allReports)

	default:
		// Text output
		for i, file := range args {
			report := allReports[i]
			if i > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "---")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "File: %s\n", file)
			fmt.Fprintf(cmd.OutOrStdout(), "Policy Set: %s\n", report.PolicySet)
			fmt.Fprintf(cmd.OutOrStdout(), "Enforcement: %s\n", report.Enforcement)
			fmt.Fprintf(cmd.OutOrStdout(), "Passed: %d\n", report.Passed)
			fmt.Fprintf(cmd.OutOrStdout(), "Failed: %d\n", report.Failed)
			fmt.Fprintf(cmd.OutOrStdout(), "Warnings: %d\n", report.Warnings)

			if len(report.Violations) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "\nViolations:")
				for _, v := range report.Violations {
					severity := strings.ToUpper(v.Severity)
					fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s: %s\n", severity, v.Policy, v.Message)
					if v.Path != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "    Path: %s\n", v.Path)
					}
					if v.Description != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "    Description: %s\n", v.Description)
					}
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "\nNo violations found.")
			}
		}
	}

	if !allPassed {
		return fmt.Errorf("policy check failed with violations")
	}

	return nil
}

func newPolicyDryRunCmd() *cobra.Command {
	var policySet string
	var inputPath string

	cmd := &cobra.Command{
		Use:   "dry-run",
		Short: "Dry-run policy evaluation",
		Long: `Evaluate policies against input without side effects.

The dry-run command shows what violations would be detected without
actually applying any enforcement. Useful for testing policy changes.

Examples:
  # Dry-run with a policy set
  omnigraph policy dry-run --policy security-baseline --input config.yaml

  # Dry-run with inline input
  echo '{"apiVersion":"omnigraph/ir/v1","kind":"InfrastructureIntent"}' | omnigraph policy dry-run --policy security-baseline`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPolicyDryRun(cmd, policySet, inputPath)
		},
	}

	cmd.Flags().StringVar(&policySet, "policy", "", "Policy set to evaluate (required)")
	cmd.Flags().StringVar(&inputPath, "input", "", "Input file to evaluate")
	cmd.MarkFlagRequired("policy")

	return cmd
}

func runPolicyDryRun(cmd *cobra.Command, policySet, inputPath string) error {
	ctx := context.Background()
	engine := policy.NewEngine()

	// Load policies from current directory
	if err := loadPoliciesFromDir(engine, "."); err != nil {
		return fmt.Errorf("failed to load policies: %w", err)
	}

	// Read input
	var input interface{}
	if inputPath != "" {
		data, err := os.ReadFile(inputPath)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		ext := filepath.Ext(inputPath)
		if ext == ".yaml" || ext == ".yml" {
			if err := yaml.Unmarshal(data, &input); err != nil {
				return fmt.Errorf("failed to parse YAML: %w", err)
			}
		} else {
			if err := json.Unmarshal(data, &input); err != nil {
				return fmt.Errorf("failed to parse JSON: %w", err)
			}
		}
	} else {
		// Read from stdin
		if err := json.NewDecoder(cmd.InOrStdin()).Decode(&input); err != nil {
			return fmt.Errorf("failed to parse stdin: %w", err)
		}
	}

	// Evaluate
	report, err := engine.Evaluate(ctx, policySet, input)
	if err != nil {
		return fmt.Errorf("failed to evaluate: %w", err)
	}

	// Output
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func newPolicyReportCmd() *cobra.Command {
	var policyDir string
	var format string
	var output string

	cmd := &cobra.Command{
		Use:   "report [file...]",
		Short: "Generate policy evaluation report",
		Long: `Generate a comprehensive policy evaluation report.

The report command evaluates multiple files and generates a consolidated
report with all violations and statistics.

Examples:
  # Generate text report
  omnigraph policy report *.yaml --policy-dir ./policies

  # Generate JSON report
  omnigraph policy report *.yaml --output report.json --format json

  # Generate YAML report
  omnigraph policy report *.yaml --output report.yaml --format yaml`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPolicyReport(cmd, args, policyDir, format, output)
		},
	}

	cmd.Flags().StringVar(&policyDir, "policy-dir", "", "Directory containing policy files")
	cmd.Flags().StringVar(&format, "format", "text", "Report format (text, json, yaml)")
	cmd.Flags().StringVar(&output, "output", "", "Output file (default: stdout)")

	return cmd
}

func runPolicyReport(cmd *cobra.Command, args []string, policyDir, format, output string) error {
	ctx := context.Background()
	engine := policy.NewEngine()

	// Load policies
	if policyDir != "" {
		if err := loadPoliciesFromDir(engine, policyDir); err != nil {
			return fmt.Errorf("failed to load policies: %w", err)
		}
	}

	// Get policy set
	sets := engine.ListPolicySets()
	if len(sets) == 0 {
		return fmt.Errorf("no policy sets loaded")
	}

	// Evaluate all files
	type fileReport struct {
		File   string               `json:"file" yaml:"file"`
		Report *policy.PolicyReport `json:"report" yaml:"report"`
	}

	var reports []fileReport
	for _, file := range args {
		report, err := engine.EvaluateFile(ctx, sets[0].Metadata.Name, file)
		if err != nil {
			return fmt.Errorf("failed to evaluate %s: %w", file, err)
		}
		reports = append(reports, fileReport{File: file, Report: report})
	}

	// Output
	var out *os.File
	if output != "" {
		var err error
		out, err = os.Create(output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	switch format {
	case "json":
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(reports)

	case "yaml":
		enc := yaml.NewEncoder(out)
		return enc.Encode(reports)

	default:
		// Text format
		for _, fr := range reports {
			fmt.Fprintf(out, "=== %s ===\n", fr.File)
			fmt.Fprintf(out, "Policy Set: %s\n", fr.Report.PolicySet)
			fmt.Fprintf(out, "Passed: %d, Failed: %d, Warnings: %d\n",
				fr.Report.Passed, fr.Report.Failed, fr.Report.Warnings)

			if len(fr.Report.Violations) > 0 {
				fmt.Fprintln(out, "\nViolations:")
				for _, v := range fr.Report.Violations {
					fmt.Fprintf(out, "  [%s] %s: %s\n", v.Severity, v.Policy, v.Message)
				}
			}
			fmt.Fprintln(out)
		}
	}

	return nil
}

func newPolicyListCmd() *cobra.Command {
	var policyDir string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available policy sets",
		Long: `List all loaded policy sets.

Examples:
  # List policy sets from directory
  omnigraph policy list --policy-dir ./policies`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPolicyList(cmd, policyDir)
		},
	}

	cmd.Flags().StringVar(&policyDir, "policy-dir", "", "Directory containing policy files")

	return cmd
}

func runPolicyList(cmd *cobra.Command, policyDir string) error {
	engine := policy.NewEngine()

	// Load policies
	if policyDir != "" {
		if err := loadPoliciesFromDir(engine, policyDir); err != nil {
			return fmt.Errorf("failed to load policies: %w", err)
		}
	}

	sets := engine.ListPolicySets()
	if len(sets) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No policy sets found")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-10s %-40s\n", "NAME", "VERSION", "DESCRIPTION")
	fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-10s %-40s\n", "----", "-------", "-----------")

	for _, ps := range sets {
		version := ps.Metadata.Version
		if version == "" {
			version = "-"
		}
		desc := ps.Metadata.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%-30s %-10s %-40s\n", ps.Metadata.Name, version, desc)
	}

	return nil
}

func newPolicyValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [file...]",
		Short: "Validate policy files",
		Long: `Validate policy set files for syntax and schema errors.

Examples:
  # Validate a single policy file
  omnigraph policy validate policy.yaml

  # Validate all policy files in directory
  omnigraph policy validate policies/*.yaml`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPolicyValidate(cmd, args)
		},
	}

	return cmd
}

func runPolicyValidate(cmd *cobra.Command, args []string) error {
	engine := policy.NewEngine()

	allValid := true
	for _, file := range args {
		_, err := engine.LoadPolicySet(file)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "INVALID %s: %v\n", file, err)
			allValid = false
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "VALID %s\n", file)
		}
	}

	if !allValid {
		return fmt.Errorf("some policy files are invalid")
	}

	return nil
}

// Helper function to load policies from directory
func loadPoliciesFromDir(engine *policy.Engine, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if _, err := engine.LoadPolicySet(path); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load %s: %v\n", path, err)
		}
	}

	return nil
}