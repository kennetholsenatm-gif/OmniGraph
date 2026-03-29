package cli

import (
	"fmt"

	"github.com/kennetholsenatm-gif/omnigraph/internal/inventory"
	"github.com/spf13/cobra"
)

func newInventoryCmd() *cobra.Command {
	fromState := &cobra.Command{
		Use:   "from-state <tfstate.json>",
		Short: "Render an Ansible INI inventory from Terraform/OpenTofu JSON state",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := inventory.FromStateFile(args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), s)
			return err
		},
	}
	root := &cobra.Command{Use: "inventory", Short: "Generate Ansible inventory snippets for pipelines (workspace has its own Inventory tab)"}
	root.AddCommand(fromState)
	return root
}
