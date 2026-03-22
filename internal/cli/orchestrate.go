package cli

import (
	"context"
	"fmt"

	"github.com/kennetholsenatm-gif/omnigraph/internal/orchestrate"
	"github.com/spf13/cobra"
)

func newOrchestrateCmd() *cobra.Command {
	var o orchestrate.Options
	var containerRuntime string

	cmd := &cobra.Command{
		Use:     "orchestrate",
		Aliases: []string{"pipeline"},
		Short:   "Run plan → ansible check → approve → apply → ansible apply (magic handoff)",
		Long: `Chains validation, OpenTofu/Terraform plan, projected inventory, ansible-playbook --check,
human approval (TTY) or --auto-approve, tofu apply, live inventory, and ansible-playbook.

Use --runner=container to execute tofu and Ansible inside Docker/Podman (see docs/execution-matrix.md).
Secrets must be passed via environment variables only (ADR 003); use your shell or a secret resolver.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			o.ContainerRuntime = containerRuntime
			r := orchestrate.NewRunner(o.Runner, containerRuntime)
			log := func(phase, detail string) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[%s] %s\n", phase, detail)
			}
			return orchestrate.Run(ctx, r, o, log)
		},
	}

	cmd.Flags().StringVar(&o.Workdir, "workdir", "", "OpenTofu/Terraform root (required)")
	cmd.Flags().StringVar(&o.SchemaPath, "schema", ".omnigraph.schema", "path to .omnigraph.schema (relative to --workdir if not absolute)")
	cmd.Flags().StringVar(&o.Playbook, "playbook", "", "Ansible playbook path relative to --workdir (required unless --skip-ansible)")
	cmd.Flags().StringVar(&o.TFBinary, "tf-binary", "tofu", "terraform or tofu binary name (exec) or first container argv token")
	cmd.Flags().StringVar(&o.PlanFile, "plan-file", "tfplan", "plan file name relative to --workdir")
	cmd.Flags().StringVar(&o.StateFile, "state-file", "terraform.tfstate", "state file relative to --workdir")
	cmd.Flags().StringVar(&o.Runner, "runner", "exec", "exec (host) or container (docker/podman)")
	cmd.Flags().StringVar(&containerRuntime, "container-runtime", "", "docker or podman (default: first found on PATH)")
	cmd.Flags().StringVar(&o.TofuImage, "tofu-image", "", "override OpenTofu/Terraform image (container runner)")
	cmd.Flags().StringVar(&o.AnsibleImage, "ansible-image", "", "override Ansible image (container runner)")
	cmd.Flags().BoolVar(&o.AutoApprove, "auto-approve", false, "skip interactive apply confirmation (required when stdin is not a TTY)")
	cmd.Flags().StringVar(&o.GraphOut, "graph-out", "", "write final omnigraph/graph/v1 JSON to this path")
	cmd.Flags().BoolVar(&o.SkipAnsible, "skip-ansible", false, "skip ansible-playbook steps (e.g. tofu-only workspaces)")
	_ = cmd.MarkFlagRequired("workdir")

	return cmd
}
