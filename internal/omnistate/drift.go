package omnistate

import (
	"context"
	"encoding/json"
	"fmt"
)

// CompareIntendedVsRuntime diffs two unified states and reports degraded nodes and fractured edges.
func CompareIntendedVsRuntime(ctx context.Context, intended, runtime *OmniGraphState) (*DriftReport, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if intended == nil || runtime == nil {
		return nil, fmt.Errorf("intended and runtime states must be non-nil")
	}
	rt := make(map[string]StateNode, len(runtime.Nodes))
	for _, n := range runtime.Nodes {
		rt[n.ID] = n
	}
	it := make(map[string]StateNode, len(intended.Nodes))
	for _, n := range intended.Nodes {
		it[n.ID] = n
	}
	rep := &DriftReport{
		AnalyzedNodes: len(intended.Nodes),
		AnalyzedEdges: len(intended.Edges),
	}
	for _, n := range intended.Nodes {
		rn, ok := rt[n.ID]
		if !ok {
			rep.DegradedNodes = append(rep.DegradedNodes, DegradedNode{
				NodeID:  n.ID,
				Reasons: []string{"unresolved_reference"},
			})
			continue
		}
		if diff := attrDiffSummary(n.Attributes, rn.Attributes); len(diff) > 0 {
			rep.DegradedNodes = append(rep.DegradedNodes, DegradedNode{
				NodeID:    n.ID,
				Reasons:   []string{"drift"},
				AttrDiffs: diff,
			})
		}
	}
	for _, e := range intended.Edges {
		_, fromOK := rt[e.From]
		_, toOK := rt[e.To]
		if fromOK && toOK {
			continue
		}
		fe := FracturedEdge{
			From:   e.From,
			To:     e.To,
			Kind:   e.Kind,
			Reason: "missing_endpoint",
		}
		if !toOK {
			fe.TranslucentTarget = e.To
		} else if !fromOK {
			fe.TranslucentTarget = e.From
			fe.Reason = "stale_dependency"
		}
		rep.FracturedEdges = append(rep.FracturedEdges, fe)
	}
	return rep, nil
}

func attrDiffSummary(a, b map[string]any) map[string]DiffPair {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	out := make(map[string]DiffPair)
	keys := make(map[string]struct{})
	for k := range a {
		keys[k] = struct{}{}
	}
	for k := range b {
		keys[k] = struct{}{}
	}
	for k := range keys {
		if k == "depends_on" {
			continue
		}
		sa := jsonScalarString(a[k])
		sb := jsonScalarString(b[k])
		if sa != sb {
			out[k] = DiffPair{Intended: sa, Runtime: sb}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func jsonScalarString(v any) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}
