package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/runner"
	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	var timeout time.Duration
	var workdir string
	cmd := &cobra.Command{
		Use:                   "run -- <command> [args...]",
		Short:                 "Run a command with optional extra env (local exec runner)",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("provide a command after --")
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			} else {
				sig := make(chan os.Signal, 1)
				signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
				defer signal.Stop(sig)
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				defer cancel()
				go func() {
					select {
					case <-sig:
						cancel()
					case <-ctx.Done():
					}
				}()
			}
			r := runner.ExecRunner{}
			res, err := r.Run(ctx, runner.Step{
				Argv:    args,
				Dir:     workdir,
				Timeout: timeout,
			})
			if err != nil {
				return err
			}
			_, _ = cmd.OutOrStdout().Write(res.Stdout)
			_, _ = cmd.ErrOrStderr().Write(res.Stderr)
			if res.ExitCode != 0 {
				os.Exit(res.ExitCode)
			}
			return nil
		},
	}
	cmd.Flags().DurationVar(&timeout, "timeout", 0, "optional timeout for the command")
	cmd.Flags().StringVar(&workdir, "workdir", "", "working directory")
	return cmd
}
