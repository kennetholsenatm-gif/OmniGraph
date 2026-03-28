package serve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/kennetholsenatm-gif/omnigraph/internal/identity"
	"github.com/kennetholsenatm-gif/omnigraph/internal/repo"
	"github.com/kennetholsenatm-gif/omnigraph/internal/version"
)

// Options configures the HTTP server.
type Options struct {
	Listen  string
	Root    string
	WebDist string
	// OnBound is called once TCP listeners are open, before serving starts.
	// Use it to log actual bound addresses (they may differ from Listen when dual-stack loopback is used).
	OnBound func(addrs []net.Addr)

	// EnableSecurityScanAPI registers POST /api/v1/security/scan (local target only; runs on the serve host).
	EnableSecurityScanAPI bool
	// EnableHostOpsAPI registers SSH-backed read endpoints for systemd units and journal tails.
	EnableHostOpsAPI bool
	// EnableInventoryAPI registers GET /api/v1/inventory (aggregated state hosts; requires AuthToken).
	EnableInventoryAPI bool
	// HostOpsAllowWrites enables POST /api/v1/host-ops/systemd/restart (dangerous; off by default).
	HostOpsAllowWrites bool
	// AuthToken is an optional static Bearer secret (OMNIGRAPH_SERVE_TOKEN). When OIDC is configured, either a valid JWT or this token may be used.
	AuthToken string
	// OIDCIssuerURL enables Keycloak (or any OIDC provider) JWT validation, e.g. https://keycloak.example/realms/myrealm
	OIDCIssuerURL string
	// OIDCClientID is the expected OAuth2 client id (access token audience / azp handling per provider).
	OIDCClientID string
	// OIDCRequiredRoles is a comma-separated list; JWT must include at least one role in realm roles, client roles, or groups (empty = no role gate).
	OIDCRequiredRoles string
	// OIDCSkipTLSVerify disables TLS verification for OIDC discovery/JWKS (development only).
	OIDCSkipTLSVerify bool
	// Authorizer overrides the default ExperimentalAuthorizer when set (tests and advanced deployments).
	Authorizer identity.Authorizer
	// EnableMetrics registers GET /metrics (Prometheus format).
	EnableMetrics bool
}

type scanRequest struct {
	Path string `json:"path"`
}

type workspaceSummary struct {
	Root           string              `json:"root"`
	Discover       *repo.Result        `json:"discover"`
	StateInventory []repo.StateHostRow `json:"stateInventory"`
	StateErrors    []string            `json:"stateErrors,omitempty"`
	OmnigraphINI   string              `json:"omnigraphIni"`
}

// Run starts the HTTP server and blocks until the context is cancelled.
func Run(ctx context.Context, opts Options) error {
	privilegedAPI := opts.EnableSecurityScanAPI || opts.EnableHostOpsAPI || opts.EnableInventoryAPI
	hasStatic := strings.TrimSpace(opts.AuthToken) != ""
	hasOIDC := strings.TrimSpace(opts.OIDCIssuerURL) != "" && strings.TrimSpace(opts.OIDCClientID) != ""
	if privilegedAPI && !hasStatic && !hasOIDC {
		return fmt.Errorf("serve: privileged APIs require --auth-token (or OMNIGRAPH_SERVE_TOKEN) and/or OIDC (--oidc-issuer and --oidc-client-id)")
	}
	if opts.Listen == "" {
		opts.Listen = "127.0.0.1:38671"
	}
	listeners, err := openListenSockets(opts.Listen)
	if err != nil {
		return err
	}
	if opts.OnBound != nil {
		addrs := make([]net.Addr, len(listeners))
		for i, ln := range listeners {
			addrs[i] = ln.Addr()
		}
		opts.OnBound(addrs)
	}

	authInit, err := initServeAuth(ctx, opts)
	if err != nil {
		return err
	}

	expAPI := privilegedAPI
	var audit *AuditLog
	if expAPI {
		audit = NewAuditLog(200)
	}
	s := &server{
		root:               opts.Root,
		authToken:          strings.TrimSpace(opts.AuthToken),
		oidcVerifier:       authInit.verifier,
		oidcClientID:       strings.TrimSpace(opts.OIDCClientID),
		authz:              authInit.authz,
		audit:              audit,
		hostOpsAllowWrites: opts.HostOpsAllowWrites,
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", s.cors(s.getHealth))
	mux.HandleFunc("POST /api/v1/repo/scan", s.cors(s.postRepoScan))
	mux.HandleFunc("POST /api/v1/workspace/summary", s.cors(s.postWorkspaceSummary))
	mux.HandleFunc("GET /api/v1/workspace/stream", s.cors(s.getWorkspaceStream))
	if opts.EnableSecurityScanAPI {
		mux.HandleFunc("POST /api/v1/security/scan", s.cors(s.requirePermission(identity.PermServeSecurityScan, s.postSecurityScanAPI)))
	}
	if opts.EnableHostOpsAPI {
		mux.HandleFunc("POST /api/v1/host-ops/systemd/units", s.cors(s.requirePermission(identity.PermServeHostOpsRead, s.postHostOpsSystemdUnits)))
		mux.HandleFunc("POST /api/v1/host-ops/journal/tail", s.cors(s.requirePermission(identity.PermServeHostOpsRead, s.postHostOpsJournal)))
		mux.HandleFunc("POST /api/v1/host-ops/systemd/restart", s.cors(s.requirePermission(identity.PermServeHostOpsWrite, s.postHostOpsRestart)))
	}
	if opts.EnableInventoryAPI {
		mux.HandleFunc("GET /api/v1/inventory", s.cors(s.requirePermission(identity.PermServeInventoryRead, s.getInventory)))
	}
	if expAPI {
		mux.HandleFunc("GET /api/v1/audit", s.cors(s.requirePermission(identity.PermServeAuditRead, s.getAudit)))
	}
	if opts.EnableMetrics {
		mux.HandleFunc("GET /metrics", s.cors(GetMetricsCollector().Handler().ServeHTTP))
	}

	if opts.WebDist != "" {
		abs, err := filepath.Abs(opts.WebDist)
		if err != nil {
			return err
		}
		st, err := os.Stat(abs)
		if err != nil || !st.IsDir() {
			return fmt.Errorf("serve: --web-dist %q is not a directory", opts.WebDist)
		}
		indexPath := filepath.Join(abs, "index.html")
		if _, err := os.Stat(indexPath); err != nil {
			return fmt.Errorf("serve: %q has no index.html — run npm run build in packages/web and point --web-dist at packages/web/dist", abs)
		}
		mux.HandleFunc("GET /", s.cors(s.staticSPA(abs)))
	} else {
		mux.HandleFunc("GET /", s.cors(s.getRootLanding))
	}

	servers := make([]*http.Server, len(listeners))
	errCh := make(chan error, len(listeners))
	var wg sync.WaitGroup
	for i, ln := range listeners {
		srv := &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		}
		servers[i] = srv
		wg.Add(1)
		go func(ln net.Listener, srv *http.Server) {
			defer wg.Done()
			err := srv.Serve(ln)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- err
			}
		}(ln, srv)
	}

	select {
	case <-ctx.Done():
		shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, srv := range servers {
			_ = srv.Shutdown(shCtx)
		}
		wg.Wait()
		return nil
	case err := <-errCh:
		shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, srv := range servers {
			_ = srv.Shutdown(shCtx)
		}
		wg.Wait()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

// openListenSockets returns TCP listeners for optsListen. When the host is 127.0.0.1,
// a second listener on [::1] is added when the OS allows it so http://localhost works
// in browsers that prefer IPv6 loopback.
func openListenSockets(optsListen string) ([]net.Listener, error) {
	host, port, err := net.SplitHostPort(optsListen)
	if err != nil {
		return nil, fmt.Errorf("invalid listen address %q: %w", optsListen, err)
	}
	if host != "127.0.0.1" {
		ln, err := net.Listen("tcp", optsListen)
		if err != nil {
			return nil, err
		}
		return []net.Listener{ln}, nil
	}
	v4 := net.JoinHostPort("127.0.0.1", port)
	ln4, err := net.Listen("tcp", v4)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", v4, err)
	}
	out := []net.Listener{ln4}
	v6 := net.JoinHostPort("::1", port)
	if ln6, err := net.Listen("tcp", v6); err == nil {
		out = append(out, ln6)
	}
	return out, nil
}

type server struct {
	root               string
	authToken          string
	oidcVerifier       *oidc.IDTokenVerifier
	oidcClientID       string
	authz              identity.Authorizer
	audit              *AuditLog
	hostOpsAllowWrites bool
}

func (s *server) getRootLanding(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	const page = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>OmniGraph</title></head>
<body>
<h1>OmniGraph</h1>
<p>API only — no static UI was configured. Use the health link below or serve the built web app.</p>
<p><a href="/api/v1/health">GET /api/v1/health</a></p>
<p>To load the React UI from this same port, build it and restart with <code>--web-dist</code> pointing at <code>packages/web/dist</code>:</p>
<pre>cd packages/web &amp;&amp; npm run build
omnigraph serve --web-dist packages/web/dist</pre>
</body>
</html>`
	_, _ = w.Write([]byte(page))
}

func (s *server) staticSPA(root string) http.HandlerFunc {
	rootClean := filepath.Clean(root)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		rel := strings.TrimPrefix(r.URL.Path, "/")
		if rel == "" {
			http.ServeFile(w, r, filepath.Join(rootClean, "index.html"))
			return
		}
		candidate := filepath.Clean(filepath.Join(rootClean, filepath.FromSlash(rel)))
		relPath, err := filepath.Rel(rootClean, candidate)
		if err != nil || strings.HasPrefix(relPath, "..") {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		st, err := os.Stat(candidate)
		if err != nil || st.IsDir() {
			http.ServeFile(w, r, filepath.Join(rootClean, "index.html"))
			return
		}
		http.ServeFile(w, r, candidate)
	}
}

func (s *server) cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func (s *server) getHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"version": version.String(),
	})
}

func (s *server) postRepoScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	root, err := s.resolveBodyPath(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := repo.Discover(root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (s *server) postWorkspaceSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	root, err := s.resolveBodyPath(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sum, err := s.workspaceSummaryForRoot(root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sum)
}

func (s *server) workspaceSummaryForRoot(root string) (workspaceSummary, error) {
	disc, err := repo.Discover(root)
	if err != nil {
		return workspaceSummary{}, err
	}
	rows, stateErrs, err := repo.AggregateStateHosts(root, 32, 0)
	if err != nil {
		return workspaceSummary{}, err
	}
	return workspaceSummary{
		Root:           disc.Root,
		Discover:       disc,
		StateInventory: rows,
		StateErrors:    stateErrs,
		OmnigraphINI:   MergedOmnigraphINI(rows),
	}, nil
}

// getWorkspaceStream streams Server-Sent Events (SSE) with periodic workspace_summary payloads.
// Query: path — same resolution rules as POST /api/v1/workspace/summary (default ".").
func (s *server) getWorkspaceStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("path"))
	if q == "" {
		q = "."
	}
	root, err := resolveWorkspacePath(s.root, q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")

	sendSummary := func() bool {
		sum, err := s.workspaceSummaryForRoot(root)
		if err != nil {
			msg, _ := json.Marshal(err.Error())
			_, _ = fmt.Fprintf(w, "event: workspace_error\ndata: %s\n\n", msg)
			flusher.Flush()
			return false
		}
		b, err := json.Marshal(sum)
		if err != nil {
			msg, _ := json.Marshal(err.Error())
			_, _ = fmt.Fprintf(w, "event: workspace_error\ndata: %s\n\n", msg)
			flusher.Flush()
			return false
		}
		_, _ = fmt.Fprintf(w, "event: workspace_summary\ndata: %s\n\n", b)
		flusher.Flush()
		return true
	}

	if !sendSummary() {
		return
	}

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			_, _ = fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case <-ticker.C:
			if !sendSummary() {
				return
			}
		}
	}
}

func (s *server) resolveBodyPath(r *http.Request) (string, error) {
	var body scanRequest
	if r.Body != nil {
		defer r.Body.Close()
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil && !errors.Is(err, io.EOF) {
			return "", fmt.Errorf("invalid JSON body: %w", err)
		}
	}
	return resolveWorkspacePath(s.root, body.Path)
}

func resolveWorkspacePath(serverRoot, reqPath string) (string, error) {
	p := strings.TrimSpace(reqPath)
	if p == "" {
		p = "."
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p), nil
	}
	base := strings.TrimSpace(serverRoot)
	if base == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		base = wd
	}
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Join(absBase, p))
}
