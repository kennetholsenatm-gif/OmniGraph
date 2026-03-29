package serve

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kennetholsenatm-gif/omnigraph/internal/identity"
	"github.com/kennetholsenatm-gif/omnigraph/internal/omnistate"
)

func TestGetWorkspaceStreamSSE(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "mod"), 0o755)
	stateJSON := []byte(`{"version":4,"values":{"outputs":{"x":{"value":"10.0.0.1"}},"root_module":{"resources":[]}}}`)
	_ = os.WriteFile(filepath.Join(dir, "mod", "terraform.tfstate"), stateJSON, 0o600)

	s := &server{root: dir}
	ts := httptest.NewServer(http.HandlerFunc(s.cors(s.getWorkspaceStream)))
	t.Cleanup(ts.Close)

	req, err := http.NewRequest(http.MethodGet, ts.URL+"?path=.", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content-type %q", ct)
	}
	// Read one chunk only — the handler keeps the connection open for periodic pushes.
	buf := make([]byte, 16384)
	n, err := res.Body.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	body := string(buf[:n])
	if !bytes.Contains([]byte(body), []byte("event: workspace_summary")) {
		t.Fatalf("missing workspace_summary event in %q", body)
	}
	if !bytes.Contains([]byte(body), []byte(`"root"`)) {
		t.Fatalf("missing json payload in %q", body)
	}
}

func TestPostWorkspaceSummary(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "mod"), 0o755)
	stateJSON := []byte(`{"version":4,"values":{"outputs":{"x":{"value":"10.0.0.1"}},"root_module":{"resources":[]}}}`)
	_ = os.WriteFile(filepath.Join(dir, "mod", "terraform.tfstate"), stateJSON, 0o600)

	s := &server{root: dir}
	ts := httptest.NewServer(http.HandlerFunc(s.cors(s.postWorkspaceSummary)))
	t.Cleanup(ts.Close)

	body := []byte(`{"path":"."}`)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var sum workspaceSummary
	if err := json.NewDecoder(res.Body).Decode(&sum); err != nil {
		t.Fatal(err)
	}
	if len(sum.StateInventory) != 1 {
		t.Fatalf("inventory %+v", sum.StateInventory)
	}
}

func TestRequireToken(t *testing.T) {
	s := &server{
		authToken: "s3cret",
		authz:     &identity.ExperimentalAuthorizer{StaticTokenConfigured: true},
	}
	called := false
	ts := httptest.NewServer(http.HandlerFunc(s.requirePermission(identity.PermServeInventoryRead, func(http.ResponseWriter, *http.Request) {
		called = true
	})))
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", res.StatusCode)
	}
	if called {
		t.Fatal("handler should not run")
	}

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer s3cret")
	res2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("want 200 got %d", res2.StatusCode)
	}
	if !called {
		t.Fatal("handler should run")
	}
}

func TestGetInventoryRequiresAuth(t *testing.T) {
	dir := t.TempDir()
	s := &server{
		root:      dir,
		authToken: "tok",
		audit:     NewAuditLog(10),
		authz:     &identity.ExperimentalAuthorizer{StaticTokenConfigured: true},
	}
	h := s.cors(s.requirePermission(identity.PermServeInventoryRead, s.getInventory))
	ts := httptest.NewServer(http.HandlerFunc(h))
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "?path=.")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", res.StatusCode)
	}
}

func TestGetInventoryJSONAndINI(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "mod"), 0o755)
	stateJSON := []byte(`{"version":4,"values":{"outputs":{"x":{"value":"10.0.0.1"}},"root_module":{"resources":[]}}}`)
	_ = os.WriteFile(filepath.Join(dir, "mod", "terraform.tfstate"), stateJSON, 0o600)

	s := &server{
		root:      dir,
		authToken: "tok",
		audit:     NewAuditLog(10),
		authz:     &identity.ExperimentalAuthorizer{StaticTokenConfigured: true},
	}
	h := s.cors(s.requirePermission(identity.PermServeInventoryRead, s.getInventory))
	ts := httptest.NewServer(http.HandlerFunc(h))
	t.Cleanup(ts.Close)

	req, err := http.NewRequest(http.MethodGet, ts.URL+"?path=.&format=json", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer tok")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var env workspaceInventoryResponse
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.APIVersion != inventoryAPIVersion {
		t.Fatalf("apiVersion %q", env.APIVersion)
	}
	if len(env.Spec.Hosts) != 1 {
		t.Fatalf("hosts %+v", env.Spec.Hosts)
	}

	req2, err := http.NewRequest(http.MethodGet, ts.URL+"?path=.&format=ini", nil)
	if err != nil {
		t.Fatal(err)
	}
	req2.Header.Set("Authorization", "Bearer tok")
	res2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("ini status %d", res2.StatusCode)
	}
	body, _ := io.ReadAll(res2.Body)
	if !bytes.Contains(body, []byte("[omnigraph]")) || !bytes.Contains(body, []byte("ansible_host=10.0.0.1")) {
		t.Fatalf("ini body %s", body)
	}
}

func TestPostIngestLocalMergeAndPartialFailure(t *testing.T) {
	hub := newSyncHub()
	s := &server{
		authToken:          "tok",
		authz:              &identity.ExperimentalAuthorizer{StaticTokenConfigured: true},
		syncHub:            hub,
		maxIngestBodyBytes: 4 << 20,
	}
	h := s.cors(s.requirePermission(identity.PermServeIngestLocal, s.postIngestLocal))
	ts := httptest.NewServer(http.HandlerFunc(h))
	t.Cleanup(ts.Close)

	tfMinimal := `{"version":4,"terraform_version":"1.0.0","values":{"root_module":{"resources":[{"address":"aws_instance.x","type":"aws_instance","name":"x","mode":"managed","provider_name":"registry.terraform.io/hashicorp/aws","values":{}}]}}}`
	body := map[string]any{
		"clientSessionId": "sess-1",
		"files": []map[string]string{
			{"name": "terraform.tfstate", "contentType": "application/json", "encoding": "utf8", "data": tfMinimal},
			{"name": "bad..name", "contentType": "text/plain", "encoding": "utf8", "data": "x"},
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, ts.URL, bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var out ingestLocalResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.State.Nodes) < 1 {
		t.Fatalf("expected terraform nodes, got %+v", out.State.Nodes)
	}
	if len(out.Errors) != 1 || out.Errors[0].Code != "E_NAME" {
		t.Fatalf("errors %+v", out.Errors)
	}
	snap := hub.snapshot()
	if len(snap.Nodes) != len(out.State.Nodes) {
		t.Fatalf("hub snapshot out of sync: hub %d resp %d", len(snap.Nodes), len(out.State.Nodes))
	}
}

func TestSyncWebSocketPingAndStateDelta(t *testing.T) {
	hub := newSyncHub()
	s := &server{
		authToken: "tok",
		authz:     &identity.ExperimentalAuthorizer{StaticTokenConfigured: true},
		syncHub:   hub,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/sync/ws", s.cors(s.requirePermission(identity.PermServeSyncWS, s.getSyncWebSocket)))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	httpURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	wsScheme := "ws"
	if httpURL.Scheme == "https" {
		wsScheme = "wss"
	}
	wsAddr := wsScheme + "://" + httpURL.Host + "/api/v1/sync/ws"
	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer tok")
	d := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	conn, _, err := d.Dial(wsAddr, hdr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(map[string]string{"type": "ping"}); err != nil {
		t.Fatal(err)
	}
	var pong map[string]string
	if err := conn.ReadJSON(&pong); err != nil {
		t.Fatal(err)
	}
	if pong["type"] != "pong" {
		t.Fatalf("pong %+v", pong)
	}

	delta := map[string]any{
		"type": "state_delta",
		"patch": omnistate.StatePatch{
			UpsertNodes: []omnistate.StateNode{
				{ID: "n-ws-test", Kind: "synthetic", Label: "ws", Provenance: omnistate.SourceRef{Type: omnistate.SourceAgentLocal, Name: "test"}},
			},
		},
	}
	if err := conn.WriteJSON(delta); err != nil {
		t.Fatal(err)
	}
	var ack wsStateAck
	if err := conn.ReadJSON(&ack); err != nil {
		t.Fatal(err)
	}
	if ack.Type != "state_ack" || ack.Revision < 1 {
		t.Fatalf("ack %+v", ack)
	}
	snap := hub.snapshot()
	var found bool
	for _, n := range snap.Nodes {
		if n.ID == "n-ws-test" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("hub missing patched node: %+v", snap.Nodes)
	}

	if err := conn.WriteJSON(map[string]any{
		"type":       "mutation_result",
		"mutationId": "m1",
		"ok":         true,
	}); err != nil {
		t.Fatal(err)
	}
	var mack wsGeneric
	if err := conn.ReadJSON(&mack); err != nil {
		t.Fatal(err)
	}
	if mack.Type != "mutation_ack" || !mack.OK || mack.MutationID != "m1" {
		t.Fatalf("mutation_ack %+v", mack)
	}
}

func TestPostIngestLocalEmptyFiles(t *testing.T) {
	s := &server{
		authToken: "tok",
		authz:     &identity.ExperimentalAuthorizer{StaticTokenConfigured: true},
	}
	h := s.cors(s.requirePermission(identity.PermServeIngestLocal, s.postIngestLocal))
	ts := httptest.NewServer(http.HandlerFunc(h))
	t.Cleanup(ts.Close)
	req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(`{"files":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d", res.StatusCode)
	}
}
