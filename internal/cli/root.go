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
		Short: "Automation and integration CLI for OmniGraph (supports the browser workspace and CI)",
		Long: `The primary OmniGraph experience is the browser workspace (topology, inventory, posture).
This binary is the headless control plane: validate schema and policy, emit graph JSON for pipelines,
orchestrate OpenTofu/Terraform and Ansible when needed, scan posture, serve local HTTP APIs the UI can call,
and produce the same versioned artifacts CI and integrations consume.`,
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
		newAuthCmd(),
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
