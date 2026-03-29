package serve

import (
	"encoding/json"
	"net/http"

	"github.com/kennetholsenatm-gif/omnigraph/internal/omnistate"
)

// postWorkspaceDrift compares intended vs runtime OmniGraphState and returns a DriftReport.
// Request JSON: { "intended": OmniGraphState, "runtime": OmniGraphState optional }.
// When "runtime" is omitted and a sync hub is configured, the hub snapshot is used as runtime.
func (s *server) postWorkspaceDrift(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	const maxBody = 8 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	defer r.Body.Close()

	var req struct {
		Intended *omnistate.OmniGraphState `json:"intended"`
		Runtime  *omnistate.OmniGraphState `json:"runtime"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Intended == nil {
		http.Error(w, "intended state required", http.StatusBadRequest)
		return
	}
	runtime := req.Runtime
	if runtime == nil {
		if s.syncHub == nil {
			http.Error(w, "runtime state required when sync hub is disabled", http.StatusBadRequest)
			return
		}
		snap := s.syncHub.snapshot()
		runtime = &snap
	}

	rep, err := omnistate.CompareIntendedVsRuntime(r.Context(), req.Intended, runtime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rep)
}
