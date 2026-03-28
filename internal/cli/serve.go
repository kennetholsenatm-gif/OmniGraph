package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/kennetholsenatm-gif/omnigraph/internal/serve"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var listen, root, webDist, authToken string
	var oidcIssuer, oidcClientID, oidcRequiredRoles string
	var oidcSkipTLS bool
	var enableSecurityScan, enableHostOps, enableInventoryAPI, hostOpsAllowWrites, enableMetrics bool
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run local HTTP API and optional web UI for repository-wide IaC views",
		Long: `Starts an HTTP server on loopback by default.

API:
  GET  /api/v1/health
  GET  /api/v1/workspace/stream  query: path=.  — SSE stream; event workspace_summary (JSON) + periodic ping
  POST /api/v1/repo/scan         body: {"path":"."}  — same discovery as omnigraph repo scan
  POST /api/v1/workspace/summary body: {"path":"."}  — discovery + aggregated state inventory + omnigraph INI

With --web-dist pointing at a Vite build (e.g. packages/web/dist after npm run build), static assets are served at /
and the UI can call the API same-origin without CORS setup.

Without --web-dist, only /api/v1/* is available; opening / in a browser will 404.

Default --listen uses 127.0.0.1 and, when possible, [::1] as well so http://localhost matches
browsers that prefer IPv6 loopback.

Security: bind address defaults to loopback only; do not expose without authentication on untrusted networks.

Experimental APIs (require --auth-token or OMNIGRAPH_SERVE_TOKEN):
  POST /api/v1/security/scan
  POST /api/v1/host-ops/systemd/units
  POST /api/v1/host-ops/journal/tail
  POST /api/v1/host-ops/systemd/restart (with --host-ops-allow-writes)
  GET  /api/v1/audit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			tok := strings.TrimSpace(authToken)
			if tok == "" {
				tok = strings.TrimSpace(os.Getenv("OMNIGRAPH_SERVE_TOKEN"))
			}

			fmt.Fprintln(cmd.ErrOrStderr(), "omnigraph serve: starting (leave this running; Ctrl+C to stop)")
			return serve.Run(ctx, serve.Options{
				Listen:                listen,
				Root:                  root,
				WebDist:               webDist,
				EnableSecurityScanAPI: enableSecurityScan,
				EnableHostOpsAPI:      enableHostOps,
				EnableInventoryAPI:    enableInventoryAPI,
				HostOpsAllowWrites:    hostOpsAllowWrites,
				AuthToken:             tok,
				OIDCIssuerURL:         strings.TrimSpace(oidcIssuer),
				OIDCClientID:          strings.TrimSpace(oidcClientID),
				OIDCRequiredRoles:     strings.TrimSpace(oidcRequiredRoles),
				OIDCSkipTLSVerify:     oidcSkipTLS,
				OnBound: func(addrs []net.Addr) {
					for _, a := range addrs {
						fmt.Fprintf(cmd.ErrOrStderr(), "  listening on %s\n", a.String())
					}
					if len(addrs) > 0 {
						host, port, err := net.SplitHostPort(addrs[0].String())
						if err == nil && host != "" && port != "" {
							v4 := net.JoinHostPort("127.0.0.1", port)
							fmt.Fprintf(cmd.ErrOrStderr(), "  health check (other terminal): curl http://%s/api/v1/health\n", v4)
						}
					}
					fmt.Fprintln(cmd.ErrOrStderr(), "  POST /api/v1/workspace/summary  body: {\"path\":\".\"}")
					if webDist != "" {
						p := mustPort(addrs)
						fmt.Fprintf(cmd.ErrOrStderr(), "  UI: http://127.0.0.1:%s/ (and http://localhost:%s/ when IPv6 loopback is bound)\n", p, p)
					}
				},
			})
		},
	}
	cmd.Flags().StringVar(&listen, "listen", "127.0.0.1:38671", "listen address (use 127.0.0.1 for local only)")
	cmd.Flags().StringVar(&root, "root", "", "default filesystem root for relative {\"path\"} in API requests (empty = process cwd)")
	cmd.Flags().StringVar(&webDist, "web-dist", "", "path to built web app directory (e.g. packages/web/dist) to serve at /")
	cmd.Flags().BoolVar(&enableSecurityScan, "enable-security-scan", false, "register POST /api/v1/security/scan (local scans; requires --auth-token)")
	cmd.Flags().BoolVar(&enableHostOps, "enable-host-ops", false, "register SSH host-ops endpoints (requires --auth-token)")
	cmd.Flags().BoolVar(&enableInventoryAPI, "enable-inventory-api", false, "register GET /api/v1/inventory (requires --auth-token)")
	cmd.Flags().BoolVar(&hostOpsAllowWrites, "host-ops-allow-writes", false, "allow systemd restart API (requires --enable-host-ops)")
	cmd.Flags().StringVar(&authToken, "auth-token", "", "Bearer token for experimental APIs (or env OMNIGRAPH_SERVE_TOKEN)")
	cmd.Flags().BoolVar(&enableMetrics, "enable-metrics", false, "enable Prometheus metrics endpoint at /metrics")
	return cmd
}

func mustPort(addrs []net.Addr) string {
	if len(addrs) == 0 {
		return "38671"
	}
	_, port, err := net.SplitHostPort(addrs[0].String())
	if err != nil {
		return "38671"
	}
	return port
}
