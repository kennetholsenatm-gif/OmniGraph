package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/ir"
	"github.com/spf13/cobra"
)

func newIRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ir",
		Short: "Infrastructure intent (omnigraph/ir/v1) — validate and list backends",
	}
	cmd.AddCommand(newIRValidateCmd(), newIRFormatsCmd(), newIREmitCmd())
	return cmd
}

func newIRValidateCmd() *cobra.Command {
	var file string
	c := &cobra.Command{
		Use:   "validate",
		Short: "Validate an IR document (JSON or YAML) against omnigraph/ir/v1",
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}
			raw, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			doc, err := ir.ParseDocument(raw)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ok: %s (%d targets, %d components, %d relations)\n",
				doc.Metadata.Name, len(doc.Spec.Targets), len(doc.Spec.Components), len(doc.Spec.Relations))
			return nil
		},
	}
	c.Flags().StringVar(&file, "file", "", "path to IR YAML or JSON")
	return c
}

func newIRFormatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "formats",
		Short: "List registered IaC backend format ids (emitters phased)",
		Run: func(cmd *cobra.Command, args []string) {
			for _, f := range ir.AllFormats() {
				fmt.Fprintln(cmd.OutOrStdout(), f)
			}
		},
	}
}

func newIREmitCmd() *cobra.Command {
	var file, format, out string
	c := &cobra.Command{
		Use:   "emit",
		Short: "Emit artifacts from an IR document using a backend (--format)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" || format == "" {
				return fmt.Errorf("--file and --format are required")
			}
			raw, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			doc, err := ir.ParseDocument(raw)
			if err != nil {
				return err
			}
			ctx := context.Background()
			arts, err := ir.DefaultRegistry().Emit(ctx, format, doc)
			if err != nil {
				return err
			}
			if len(arts) == 0 {
				return fmt.Errorf("emitter returned no artifacts")
			}
			if out == "" || out == "-" {
				for _, a := range arts {
					if _, err := cmd.OutOrStdout().Write(a.Content); err != nil {
						return err
					}
				}
				return nil
			}
			st, err := os.Stat(out)
			if err != nil || !st.IsDir() {
				if len(arts) != 1 {
					return fmt.Errorf("--out must be a directory when the backend emits multiple files")
				}
				if err := os.WriteFile(out, arts[0].Content, 0o644); err != nil {
					return err
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "wrote %s (%d bytes)\n", out, len(arts[0].Content))
				return nil
			}
			for _, a := range arts {
				p := filepath.Join(out, filepath.FromSlash(strings.TrimPrefix(a.Path, "/")))
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(p, a.Content, 0o644); err != nil {
					return err
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "wrote %s\n", p)
			}
			return nil
		},
	}
	c.Flags().StringVar(&file, "file", "", "path to IR YAML or JSON")
	c.Flags().StringVar(&format, "format", "", "backend id (see: omnigraph ir formats)")
	c.Flags().StringVar(&out, "out", "-", "output path: \"-\" for stdout, a file path for single artifact, or a directory for multiple")
	return c
}
