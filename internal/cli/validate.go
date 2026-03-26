package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kennetholsenatm-gif/omnigraph/internal/policy"
	"github.com/kennetholsenatm-gif/omnigraph/internal/project"
	"github.com/kennetholsenatm-gif/omnigraph/internal/schema"
	"github.com/kennetholsenatm-gif/omnigraph/internal/serve"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	var path string
	var policyDir string
	var enforcePolicy bool
	cmd := &cobra.Command{
		Use:   "validate [path]",
		Short: "Validate an .omnigraph.schema document against the JSON Schema",
		Long: `Validate an .omnigraph.schema document against the JSON Schema.

By default, validates only the schema syntax. Use --policy-dir to also
validate against policy-as-code rules.

Examples:
  # Validate schema only
  omnigraph validate .omnigraph.schema

  # Validate with policy checks
  omnigraph validate .omnigraph.schema --policy-dir ./policies

  # Validate with policy enforcement (fail on violations)
  omnigraph validate .omnigraph.schema --policy-dir ./policies --enforce`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := path
			if len(args) > 0 {
				p = args[0]
			}
			if p == "" {
				p = ".omnigraph.schema"
			}
			raw, err := os.ReadFile(p)
			if err != nil {
				return fmt.Errorf("read %q: %w", p, err)
			}
			if _, err := schema.ValidateRawDocument(raw); err != nil {
				return err
			}
			doc, err := project.ParseDocument(raw)
			if err != nil {
				return err
			}
			if doc.Metadata.Name == "" {
				return fmt.Errorf("validated document missing metadata.name")
			}

			// Policy validation if requested
			if policyDir != "" {
				ctx := context.Background()
				engine := policy.NewEngine()

				// Load policies
				if err := loadPoliciesFromDir(engine, policyDir); err != nil {
					return fmt.Errorf("failed to load policies: %w", err)
				}

				// Get policy set
				sets := engine.ListPolicySets()
				if len(sets) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "Warning: no policy sets found in", policyDir)
				} else {
					// Evaluate document against policies
					var input interface{}
					if err := json.Unmarshal(raw, &input); err != nil {
						return fmt.Errorf("failed to parse document for policy evaluation: %w", err)
					}

					for _, ps := range sets {
						report, err := engine.Evaluate(ctx, ps.Metadata.Name, input)
						if err != nil {
							return fmt.Errorf("failed to evaluate policies: %w", err)
						}

						// Record policy metrics
						serve.GetMetricsCollector().RecordPolicyMetrics(report)

						if len(report.Violations) > 0 {
							if enforcePolicy && report.Enforcement == "deny" {
								// Fail validation
								fmt.Fprintf(cmd.OutOrStderr(), "Policy violations found (%s):\n", ps.Metadata.Name)
								for _, v := range report.Violations {
									fmt.Fprintf(cmd.OutOrStderr(), "  [%s] %s: %s\n", v.Severity, v.Policy, v.Message)
								}
								return fmt.Errorf("validation failed due to policy violations")
							} else {
								// Warn only
								fmt.Fprintf(cmd.OutOrStdout(), "Policy warnings (%s):\n", ps.Metadata.Name)
								for _, v := range report.Violations {
									fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s: %s\n", v.Severity, v.Policy, v.Message)
								}
							}
						} else {
							fmt.Fprintf(cmd.OutOrStdout(), "Policy check passed (%s): %d policies evaluated\n",
								ps.Metadata.Name, report.Passed)
						}
					}
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Schema validation: ok")
			return nil
		},
	}
	cmd.Flags().StringVarP(&path, "file", "f", "", "path to schema file (default .omnigraph.schema)")
	cmd.Flags().StringVar(&policyDir, "policy-dir", "", "directory containing policy files for validation")
	cmd.Flags().BoolVar(&enforcePolicy, "enforce", false, "enforce policy validation (fail on violations)")
	return cmd
}
