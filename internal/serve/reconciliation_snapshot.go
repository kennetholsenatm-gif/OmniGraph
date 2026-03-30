package serve

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/omnistate"
)

func (s *server) postReconciliationSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	root, err := s.resolveBodyPath(r)
	if err != nil {
		writeAPIErrorJSON(w, "RECONCILIATION_PATH_INVALID", "reconciliation/snapshot: "+err.Error(), http.StatusBadRequest)
		return
	}
	sum, err := s.workspaceSummaryForRoot(root)
	if err != nil {
		writeAPIErrorJSON(w, "RECONCILIATION_SUMMARY_FAILED", "reconciliation/snapshot: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var runtime omnistate.OmniGraphState
	if s.syncHub != nil {
		snap := s.syncHub.snapshot()
		runtime = snap
	} else {
		runtime = omnistate.OmniGraphState{
			APIVersion:  omnistate.APIVersion,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
			Nodes:       []omnistate.StateNode{},
			Edges:       []omnistate.StateEdge{},
		}
	}

	// Intended is currently reconstructed from summary-derived inventory rows.
	intended := omnistate.OmniGraphState{
		APIVersion: omnistate.APIVersion,
		Nodes:      make([]omnistate.StateNode, 0, len(sum.StateInventory)),
		Edges:      []omnistate.StateEdge{},
	}
	for _, row := range sum.StateInventory {
		intended.Nodes = append(intended.Nodes, omnistate.StateNode{
			ID:    "inv:" + row.Name,
			Kind:  "service",
			Label: row.Name,
			Attributes: map[string]any{
				"ansible_host": row.AnsibleHost,
				"origin":       row.Origin,
			},
			Provenance: omnistate.SourceRef{
				Type:     omnistate.SourceTerraformState,
				Name:     row.Origin,
				PathHint: row.Origin,
			},
		})
	}

	rep, err := omnistate.CompareIntendedVsRuntime(r.Context(), &intended, &runtime)
	if err != nil {
		writeAPIErrorJSON(w, "RECONCILIATION_COMPARE_FAILED", "reconciliation/snapshot: "+err.Error(), http.StatusBadRequest)
		return
	}
	relationDrifts := omnistate.BuildRelationDrifts(&intended, &runtime)
	bom := omnistate.BuildBOMFromState(&runtime, "serve:reconciliation")
	if len(bom.Spec.Entities) == 0 {
		// Fallback to intended evidence so Inventory/Topology can still render context.
		bom = omnistate.BuildBOMFromState(&intended, "serve:reconciliation")
	}
	snapshot := omnistate.ReconciliationSnapshot{
		APIVersion: omnistate.ReconciliationSnapshotAPIVersion,
		Kind:       "ReconciliationSnapshot",
		Metadata: omnistate.ReconciliationSnapshotMeta{
			GeneratedAt: bom.Metadata.GeneratedAt,
			Source:      "serve:reconciliation",
			Revision:    runtime.Revision,
		},
		Spec: omnistate.ReconciliationSnapshotSpec{
			BOM:            bom,
			DegradedNodes:  rep.DegradedNodes,
			FracturedEdges: rep.FracturedEdges,
			RelationDrifts: relationDrifts,
			NextActions: []string{
				"Refresh Inventory evidence from server summary or ingest/local to confirm missing dependencies.",
				"Inspect selected topology node and compare intended attributes against runtime values.",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snapshot)
}
