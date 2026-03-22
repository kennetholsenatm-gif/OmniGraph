package cli

import (
	"fmt"
	"os"

	"github.com/kennetholsenatm-gif/omnigraph/internal/project"
	"github.com/kennetholsenatm-gif/omnigraph/internal/schema"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "validate [path]",
		Short: "Validate an .omnigraph.schema document against the JSON Schema",
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
			fmt.Fprintln(cmd.OutOrStdout(), "ok")
			return nil
		},
	}
	cmd.Flags().StringVarP(&path, "file", "f", "", "path to schema file (default .omnigraph.schema)")
	return cmd
}
