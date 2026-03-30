package e2e

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSimulatedAnsibleEndpoint_ServiceUnavailable documents the E2E pattern for
// standing in for Ansible-facing HTTP surfaces: inject failures (503, timeouts,
// malformed bodies) before wiring full orchestration tests.
func TestSimulatedAnsibleEndpoint_ServiceUnavailable(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.Error(w, "playbook runner unavailable", http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/mock-playbook")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status %d", resp.StatusCode)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
}
