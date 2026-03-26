package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kennetholsenatm-gif/omnigraph/internal/netbox"
	"github.com/spf13/cobra"
)

func newNetBoxCmd() *cobra.Command {
	var (
		url, payloadVersion, action, ip, role, cidr, environment, idempotencyKey string
		siteID, deviceID                                                         int
	)
	sync := &cobra.Command{
		Use:   "sync",
		Short: "POST a NetBox-compatible sync payload to a webhook URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			if url == "" {
				return fmt.Errorf("--url is required")
			}
			c := &netbox.Client{}
			ctx := context.Background()
			switch payloadVersion {
			case "v1":
				w := &netbox.WebhookV1{
					Action:         action,
					IP:             ip,
					CIDR:           cidr,
					Role:           role,
					SiteID:         siteID,
					DeviceID:       deviceID,
					Environment:    environment,
					IdempotencyKey: idempotencyKey,
				}
				b, err := json.Marshal(w)
				if err != nil {
					return err
				}
				hdr := map[string]string{}
				if idempotencyKey != "" {
					hdr["X-Omnigraph-Idempotency-Key"] = idempotencyKey
				}
				return c.PostJSON(ctx, url, b, hdr)
			case "legacy":
				if ip == "" {
					return fmt.Errorf("legacy payload: --ip is required")
				}
				return c.PostWebhook(ctx, url, netbox.SyncPayload{
					Action: action,
					IP:     ip,
					Role:   role,
				})
			default:
				return fmt.Errorf("--payload-version must be v1 or legacy")
			}
		},
	}
	sync.Flags().StringVar(&url, "url", "", "webhook URL (required)")
	sync.Flags().StringVar(&payloadVersion, "payload-version", "v1", "v1 (versioned JSON) or legacy (illustrative shape without apiVersion)")
	sync.Flags().StringVar(&action, "action", "create", "action field in JSON payload")
	sync.Flags().StringVar(&ip, "ip", "", "host IP (v1: optional if --cidr set; legacy: required)")
	sync.Flags().StringVar(&cidr, "cidr", "", "CIDR prefix (v1 only)")
	sync.Flags().StringVar(&role, "role", "", "optional role label")
	sync.Flags().IntVar(&siteID, "site-id", 0, "optional NetBox site id (v1)")
	sync.Flags().IntVar(&deviceID, "device-id", 0, "optional NetBox device id (v1)")
	sync.Flags().StringVar(&environment, "environment", "", "optional environment label (v1)")
	sync.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "optional idempotency token (also sent as X-Omnigraph-Idempotency-Key)")

	root := &cobra.Command{Use: "netbox", Short: "CMDB integration helpers"}
	root.AddCommand(sync)
	return root
}
