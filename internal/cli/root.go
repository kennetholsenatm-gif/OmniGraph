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
		Use:           "omnigraph",
		Short:         "OmniGraph control plane — orchestrates provisioning, configuration, and handoff",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.String(),
	}
	root.SetVersionTemplate("{{.Version}}\n")
	root.AddCommand(
		newValidateCmd(),
		newCoerceCmd(),
		newStateCmd(),
		newInventoryCmd(),
		newGraphCmd(),
		newRunCmd(),
		newNetBoxCmd(),
	)
	return root
}
