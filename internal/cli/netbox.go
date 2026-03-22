package cli

import (
	"context"
	"fmt"

	"github.com/kennetholsenatm-gif/omnigraph/internal/netbox"
	"github.com/spf13/cobra"
)

func newNetBoxCmd() *cobra.Command {
	var url, action, ip, role string
	sync := &cobra.Command{
		Use:   "sync",
		Short: "POST an illustrative sync payload to a NetBox-compatible webhook URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			if url == "" {
				return fmt.Errorf("--url is required")
			}
			if ip == "" {
				return fmt.Errorf("--ip is required")
			}
			c := &netbox.Client{}
			return c.PostWebhook(context.Background(), url, netbox.SyncPayload{
				Action: action,
				IP:     ip,
				Role:   role,
			})
		},
	}
	sync.Flags().StringVar(&url, "url", "", "webhook URL (required)")
	sync.Flags().StringVar(&action, "action", "create", "action field in JSON payload")
	sync.Flags().StringVar(&ip, "ip", "", "IP address (required)")
	sync.Flags().StringVar(&role, "role", "", "optional role label")
	root := &cobra.Command{Use: "netbox", Short: "CMDB integration helpers"}
	root.AddCommand(sync)
	return root
}
