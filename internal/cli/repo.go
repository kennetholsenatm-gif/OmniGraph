package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kennetholsenatm-gif/omnigraph/internal/repo"
	"github.com/spf13/cobra"
)

func newRepoCmd() *cobra.Command {
	var scanPath string
	scan := &cobra.Command{
		Use:   "scan",
		Short: "Walk a repository tree and list IaC artifacts (state, HCL, Ansible, schema)",
		Long: `Recursively scans a working tree (e.g. a Git checkout), skipping common vendor/cache dirs,
and emits JSON describing discovered Terraform/OpenTofu state, .tf/.tofu files, .omnigraph.schema,
Ansible cfg/playbooks/inventory-style paths, and tfplan files.

Use this as the machine-facing view of "the whole repo" for tooling and UIs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if scanPath == "" {
				scanPath = "."
			}
			res, err := repo.Discover(scanPath)
			if err != nil {
				return err
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			if err := enc.Encode(res); err != nil {
				return err
			}
			return nil
		},
	}
	scan.Flags().StringVar(&scanPath, "path", ".", "repository root to scan")

	root := &cobra.Command{
		Use:   "repo",
		Short: "Repository-wide discovery (IaC layout under a checkout)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("use a subcommand: %s repo scan --path <dir>", os.Args[0])
		},
	}
	root.AddCommand(scan)
	return root
}
