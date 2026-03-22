package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/kennetholsenatm-gif/omnigraph/internal/state"
	"github.com/spf13/cobra"
)

func newStateCmd() *cobra.Command {
	parse := &cobra.Command{
		Use:   "parse <tfstate.json>",
		Short: "Parse a Terraform/OpenTofu JSON state file and print a normalized summary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := state.Load(args[0])
			if err != nil {
				return err
			}
			hosts := state.ExtractHosts(st)
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]any{
				"hosts": hosts,
			})
		},
	}
	hosts := &cobra.Command{
		Use:   "hosts <tfstate.json>",
		Short: "Print extracted ansible_host candidates (address -> IP)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := state.Load(args[0])
			if err != nil {
				return err
			}
			h := state.ExtractHosts(st)
			keys := make([]string, 0, len(h))
			for k := range h {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", k, h[k])
			}
			return nil
		},
	}
	root := &cobra.Command{Use: "state", Short: "Inspect Terraform/OpenTofu state"}
	root.AddCommand(parse, hosts)
	return root
}
