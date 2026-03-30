package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/kennetholsenatm-gif/omnigraph/internal/serve"
	"github.com/kennetholsenatm-gif/omnigraph/internal/version"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "integration-run" {
		if err := runIntegrationRun(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	fs := flag.NewFlagSet("omnigraph", flag.ExitOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: omnigraph [flags]\n       omnigraph integration-run --wasm=REL_PATH < run.json\n\nWorkspace server — HTTP API for the web UI. Subcommand integration-run: --wasm relative to cwd; stdin omnigraph/integration-run/v1.\n\nServer flags:\n")
		fs.PrintDefaults()
	}
	var (
		listen, root, webDist, authToken                                                         string
		oidcIssuer, oidcClientID, oidcRequiredRoles                                              string
		oidcSkipTLS                                                                              bool
		enableSecurityScan, enableHostOps, enableInventoryAPI, hostOpsAllowWrites, enableMetrics bool
		enableIngestLocal, enableSyncWS, enableWorkspaceDrift, enableIntegrationRun              bool
		maxIngestBodyMB                                                                          int64
		showVersion                                                                              bool
	)
	fs.StringVar(&listen, "listen", "127.0.0.1:38671", "listen address (use 127.0.0.1 for local only)")
	fs.StringVar(&root, "root", "", "default filesystem root for relative {\"path\"} in API requests (empty = process cwd)")
	fs.StringVar(&webDist, "web-dist", "", "path to built web app directory (e.g. packages/web/dist) to serve at /")
	fs.BoolVar(&enableSecurityScan, "enable-security-scan", false, "register POST /api/v1/security/scan (local scans; requires --auth-token)")
	fs.BoolVar(&enableHostOps, "enable-host-ops", false, "register SSH host-ops endpoints (requires --auth-token)")
	fs.BoolVar(&enableInventoryAPI, "enable-inventory-api", false, "register GET /api/v1/inventory (requires --auth-token)")
	fs.BoolVar(&hostOpsAllowWrites, "host-ops-allow-writes", false, "allow systemd restart API (requires --enable-host-ops)")
	fs.StringVar(&authToken, "auth-token", "", "Bearer token for experimental APIs (or env OMNIGRAPH_SERVE_TOKEN)")
	fs.BoolVar(&enableMetrics, "enable-metrics", false, "enable Prometheus metrics endpoint at /metrics")
	fs.BoolVar(&enableIngestLocal, "enable-ingest-local-api", false, "register POST /api/v1/ingest/local (requires --auth-token)")
	fs.BoolVar(&enableSyncWS, "enable-sync-ws-api", false, "register GET /api/v1/sync/ws (requires --auth-token)")
	fs.BoolVar(&enableWorkspaceDrift, "enable-workspace-drift-api", false, "register POST /api/v1/workspace/drift (requires --auth-token)")
	fs.BoolVar(&enableIntegrationRun, "enable-integration-run-api", false, "register POST /api/v1/integrations/run (WASM integrations; requires --auth-token)")
	fs.Int64Var(&maxIngestBodyMB, "max-ingest-body-mb", 0, "max ingest JSON body in MiB (0 = default 64)")
	fs.StringVar(&oidcIssuer, "oidc-issuer", "", "OIDC issuer URL for JWT validation")
	fs.StringVar(&oidcClientID, "oidc-client-id", "", "expected OAuth2 client id (audience)")
	fs.StringVar(&oidcRequiredRoles, "oidc-required-roles", "", "comma-separated roles required in JWT")
	fs.BoolVar(&oidcSkipTLS, "oidc-skip-tls-verify", false, "skip TLS verify for OIDC discovery (dev only)")
	fs.BoolVar(&showVersion, "version", false, "print version and exit")

	_ = fs.Parse(os.Args[1:])
	if showVersion {
		fmt.Println(version.String())
		return
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n", fs.Args())
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tok := strings.TrimSpace(authToken)
	if tok == "" {
		tok = strings.TrimSpace(os.Getenv("OMNIGRAPH_SERVE_TOKEN"))
	}

	fmt.Fprintln(os.Stderr, "OmniGraph workspace server: starting (leave this running; Ctrl+C to stop)")
	var maxIngestBytes int64
	if maxIngestBodyMB > 0 {
		maxIngestBytes = maxIngestBodyMB << 20
	}
	if err := serve.Run(ctx, serve.Options{
		Listen:                  listen,
		Root:                    root,
		WebDist:                 webDist,
		EnableSecurityScanAPI:   enableSecurityScan,
		EnableHostOpsAPI:        enableHostOps,
		EnableInventoryAPI:      enableInventoryAPI,
		HostOpsAllowWrites:      hostOpsAllowWrites,
		EnableMetrics:           enableMetrics,
		EnableIngestLocalAPI:    enableIngestLocal,
		EnableSyncWSAPI:         enableSyncWS,
		EnableWorkspaceDriftAPI: enableWorkspaceDrift,
		EnableIntegrationRunAPI: enableIntegrationRun,
		MaxIngestBodyBytes:      maxIngestBytes,
		AuthToken:               tok,
		OIDCIssuerURL:           strings.TrimSpace(oidcIssuer),
		OIDCClientID:            strings.TrimSpace(oidcClientID),
		OIDCRequiredRoles:       strings.TrimSpace(oidcRequiredRoles),
		OIDCSkipTLSVerify:       oidcSkipTLS,
		OnBound: func(addrs []net.Addr) {
			for _, a := range addrs {
				fmt.Fprintf(os.Stderr, "  listening on %s\n", a.String())
			}
			if len(addrs) > 0 {
				host, port, err := net.SplitHostPort(addrs[0].String())
				if err == nil && host != "" && port != "" {
					v4 := net.JoinHostPort("127.0.0.1", port)
					fmt.Fprintf(os.Stderr, "  health check (other terminal): curl http://%s/api/v1/health\n", v4)
				}
			}
			fmt.Fprintln(os.Stderr, "  POST /api/v1/workspace/summary  body: {\"path\":\".\"}")
			if webDist != "" {
				p := mustPort(addrs)
				fmt.Fprintf(os.Stderr, "  UI: http://127.0.0.1:%s/ (and http://localhost:%s/ when IPv6 loopback is bound)\n", p, p)
			}
		},
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runIntegrationRun(args []string) error {
	fs := flag.NewFlagSet("integration-run", flag.ExitOnError)
	fs.SetOutput(os.Stderr)
	var wasm string
	fs.StringVar(&wasm, "wasm", "", "path to integration .wasm (wasip1 module)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: omnigraph integration-run --wasm=REL_PATH < run.json\n\n--wasm must be relative to the current working directory (no absolute paths). stdin must be omnigraph/integration-run/v1 JSON.\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(wasm) == "" {
		return fmt.Errorf("integration-run: --wasm is required")
	}
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	ctx := context.Background()
	out, err := serve.RunIntegrationCLI(ctx, wasm, stdin)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(out)
	return err
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
