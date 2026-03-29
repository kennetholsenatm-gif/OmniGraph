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
	var manifest bool
	scan := &cobra.Command{
		Use:   "scan",
		Short: "List IaC paths under a repo (automation; mirrors serve discovery)",
		Long: `Recursively scans a working tree (e.g. a Git checkout), skipping common vendor/cache dirs,
and emits JSON describing discovered Terraform/OpenTofu state, .tf/.tofu files, .omnigraph.schema,
Ansible cfg/playbooks/inventory-style paths, and tfplan files.

With --manifest, performs a deeper scan (Terraform state + Ansible INI/YAML inventories) and emits
connector-style graph manifests (parity with the former omnigraph-agent discover --json output shape).

Use this as the machine-facing view of "the whole repo" for tooling and UIs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if scanPath == "" {
				scanPath = "."
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			if manifest {
				md, err := repo.DiscoverManifests([]string{scanPath}, repo.DefaultManifestDiscoverOptions())
				if err != nil {
					return err
				}
				if err := enc.Encode(md); err != nil {
					return err
				}
				return nil
			}
			res, err := repo.Discover(scanPath)
			if err != nil {
				return err
			}
			if err := enc.Encode(res); err != nil {
				return err
			}
			return nil
		},
	}
	scan.Flags().StringVar(&scanPath, "path", ".", "repository root to scan")
	scan.Flags().BoolVar(&manifest, "manifest", false, "emit graph manifests from Terraform state and Ansible inventories (deep scan)")

	root := &cobra.Command{
		Use:   "repo",
		Short: "Repository-wide discovery commands for CI and integrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("use a subcommand: %s repo scan --path <dir>", os.Args[0])
		},
	}
	root.AddCommand(scan)
	return root
}
