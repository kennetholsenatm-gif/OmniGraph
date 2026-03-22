package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/coerce"
	"github.com/kennetholsenatm-gif/omnigraph/internal/project"
	"github.com/kennetholsenatm-gif/omnigraph/internal/schema"
	"github.com/spf13/cobra"
)

func newCoerceCmd() *cobra.Command {
	var path, format string
	cmd := &cobra.Command{
		Use:   "coerce [path]",
		Short: "Coerce a validated .omnigraph.schema into in-memory tool representations",
		Long: strings.TrimSpace(`
Prints Terraform tfvars JSON, Ansible group_vars/all YAML, and environment lines.
Secrets are never emitted by this command; inject credentials via a secret resolver at run time.`),
		Args: cobra.MaximumNArgs(1),
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
			art, err := coerce.FromDocument(doc)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			f := strings.ToLower(strings.TrimSpace(format))
			switch f {
			case "tfvars", "terraform":
				b, err := coerce.FormatTerraformTfvarsJSON(art)
				if err != nil {
					return err
				}
				fmt.Fprintln(out, string(b))
			case "groupvars", "ansible":
				fmt.Fprint(out, string(art.GroupVarsAllYAML))
			case "env":
				fmt.Fprint(out, coerce.FormatEnvLines(art))
			case "all":
				fmt.Fprintln(out, "### terraform.tfvars.json")
				b, err := coerce.FormatTerraformTfvarsJSON(art)
				if err != nil {
					return err
				}
				fmt.Fprintln(out, string(b))
				fmt.Fprintln(out, "### group_vars/all.yml")
				fmt.Fprint(out, string(art.GroupVarsAllYAML))
				fmt.Fprintln(out, "### env")
				fmt.Fprint(out, coerce.FormatEnvLines(art))
			default:
				return fmt.Errorf("unknown --format %q (use tfvars, groupvars, env, all)", format)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&path, "file", "f", "", "path to schema file (default .omnigraph.schema)")
	cmd.Flags().StringVar(&format, "format", "all", "output format: tfvars | groupvars | env | all")
	return cmd
}
