package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kennetholsenatm-gif/omnigraph/internal/coerce"
	"github.com/kennetholsenatm-gif/omnigraph/internal/graph"
	"github.com/kennetholsenatm-gif/omnigraph/internal/project"
	"github.com/kennetholsenatm-gif/omnigraph/internal/schema"
	"github.com/kennetholsenatm-gif/omnigraph/internal/security"
	"github.com/kennetholsenatm-gif/omnigraph/internal/state"
	"github.com/spf13/cobra"
)

func newParseCmd() *cobra.Command {
	var path string
	parse := &cobra.Command{
		Use:   "parse [path]",
		Short: "Parse omnigraph/graph/v1 JSON and construct graph structure",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := path
			if len(args) > 0 {
				p = args[0]
			}
			if p == "" {
				return fmt.Errorf("path to JSON file is required")
			}
			raw, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			doc, err := graph.ParseDocument(raw)
			if err != nil {
				return err
			}
			graph, err := graph.ConstructFromDocument(doc)
			if err != nil {
				return err
			}
			b, err := json.MarshalIndent(graph, "", "  ")
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
	parse.Flags().StringVarP(&path, "file", "f", "", "path to JSON file (default stdin)")
	return parse
}

func newGraphCmd() *cobra.Command {
	var path, planJSON, tfState, telemetryFile, securityFile string
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
			opts := graph.EmitOptions{PlanJSONPath: planJSON, TelemetryPath: telemetryFile}
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
			if securityFile != "" {
				secdoc, err := security.LoadDocument(securityFile)
				if err != nil {
					return err
				}
				graph.MergeSecurity(gdoc, secdoc)
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
	emit.Flags().StringVar(&telemetryFile, "telemetry-file", "", "path to omnigraph/telemetry/v1 JSON to merge into the graph")
	emit.Flags().StringVar(&securityFile, "security-file", "", "path to omnigraph/security/v1 JSON to merge securityPosture into host nodes")
	root := &cobra.Command{Use: "graph", Short: "Build and parse graph documents"}
	root.AddCommand(emit)
	root.AddCommand(newParseCmd())
	return root
}
