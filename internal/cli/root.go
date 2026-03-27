package cli

import (
	"fmt"
	"os"

	"github.com/kennetholsenatm-gif/omnigraph/internal/version"
	"github.com/spf13/cobra"
)

// Execute runs the Cobra command tree.
func Execute() {
	if err := newRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "omnigraph",
		Short: "OmniGraph — graph workspace, validation, and automation (CLI feeds the web UI and CI)",
		Long: `OmniGraph is a graph-forward infrastructure workspace: the React UI is where most people
explore omnigraph/graph/v1 and related context. This binary validates schema, runs policy,
orchestrates OpenTofu/Terraform and Ansible when needed, scans posture, serves HTTP APIs,
and emits the JSON artifacts the dashboard consumes—plus headless CI use cases.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.String(),
	}
	root.SetVersionTemplate("{{.Version}}\n")
	root.AddCommand(
		newValidateCmd(),
		newPolicyCmd(),
		newCoerceCmd(),
		newOrchestrateCmd(),
		newStateCmd(),
		newInventoryCmd(),
		newGraphCmd(),
		newRunCmd(),
		newNetBoxCmd(),
		newRepoCmd(),
		newServeCmd(),
		newSecurityCmd(),
		newIRCmd(),
		newApplyCmd(),
		newGetCmd(),
		newDescribeCmd(),
		newDiffCmd(),
		newEnclaveCmd(),
	)
	return root
}
