package cli

import (
	"os"

	"github.com/kennetholsenatm-gif/omnigraph/internal/coerce"
	"github.com/kennetholsenatm-gif/omnigraph/internal/graph"
	"github.com/kennetholsenatm-gif/omnigraph/internal/project"
	"github.com/kennetholsenatm-gif/omnigraph/internal/schema"
	"github.com/kennetholsenatm-gif/omnigraph/internal/state"
	"github.com/spf13/cobra"
)

func newGraphCmd() *cobra.Command {
	var path, planJSON, tfState string
	emit := &cobra.Command{
		Use:   "emit [path]",
		Short: "Emit omnigraph/graph/v1 JSON for UI and CI artifacts",
		Args:  cobra.MaximumNArgs(1),
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
				return err
			}
			if _, err := schema.ValidateRawDocument(raw); err != nil {
				return err
			}
			doc, err := project.ParseDocument(raw)
			if err != nil {
				return err
			}
			art, err := coerce.FromDocument(doc)
			if err != nil {
				return err
			}
			opts := graph.EmitOptions{PlanJSONPath: planJSON}
			if tfState != "" {
				st, err := state.Load(tfState)
				if err != nil {
					return err
				}
				opts.TerraformState = st
			}
			gdoc, err := graph.Emit(doc, art, opts)
			if err != nil {
				return err
			}
			b, err := graph.EncodeIndent(gdoc)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(b)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write([]byte("\n"))
			return err
		},
	}
	emit.Flags().StringVarP(&path, "file", "f", "", "path to .omnigraph.schema (default .omnigraph.schema)")
	emit.Flags().StringVar(&planJSON, "plan-json", "", "path to `terraform show -json` plan output")
	emit.Flags().StringVar(&tfState, "tfstate", "", "path to Terraform/OpenTofu JSON state after apply")
	root := &cobra.Command{Use: "graph", Short: "Build graph documents for visualization"}
	root.AddCommand(emit)
	return root
}
