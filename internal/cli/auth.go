package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/cloudauth"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	var issuer, clientID, scope string
	device := &cobra.Command{
		Use:   "oidc-device",
		Short: "Run OIDC device authorization flow and print short-lived tokens as JSON",
		Long: `Performs RFC 8628 device authorization against an OIDC issuer (e.g. Keycloak, Azure AD, Okta).
Use the printed access_token or id_token with cloud APIs (e.g. AWS STS AssumeRoleWithWebIdentity) instead of
long-lived static keys. The issuer must advertise device_authorization_endpoint in its OpenID configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if issuer == "" || clientID == "" {
				return fmt.Errorf("--issuer and --client-id are required")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
			defer cancel()
			meta, err := cloudauth.FetchOIDCMetadata(ctx, issuer, nil)
			if err != nil {
				return err
			}
			dev, err := cloudauth.StartDeviceFlow(ctx, meta, clientID, scope, nil)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.ErrOrStderr(), "Open the verification URL and approve the request:")
			if dev.VerificationURIComplete != "" {
				fmt.Fprintln(cmd.ErrOrStderr(), dev.VerificationURIComplete)
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s  user_code=%s\n", dev.VerificationURI, dev.UserCode)
			}
			tok, err := cloudauth.PollDeviceToken(ctx, meta, clientID, dev.DeviceCode, dev, nil)
			if err != nil {
				return err
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(tok)
		},
	}
	device.Flags().StringVar(&issuer, "issuer", "", "OIDC issuer URL (e.g. https://keycloak.example/realms/realm)")
	device.Flags().StringVar(&clientID, "client-id", "", "OAuth2 public client id enabled for device flow")
	device.Flags().StringVar(&scope, "scope", "openid profile email offline_access", "OAuth2 scopes")

	root := &cobra.Command{
		Use:   "auth",
		Short: "OIDC and token helpers for automation (cloud identity)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("use a subcommand, e.g. %s auth oidc-device --help", os.Args[0])
		},
	}
	root.AddCommand(device)
	return root
}
