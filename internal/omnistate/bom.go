package omnistate

import "time"

const (
	BOMAPIVersion                    = "omnigraph/bom/v1"
	ReconciliationSnapshotAPIVersion = "omnigraph/reconciliation-snapshot/v1"
)

type BOMEntityClass string

const (
	BOMEntitySoftware BOMEntityClass = "software_component"
	BOMEntityHardware BOMEntityClass = "hardware_asset"
	BOMEntityService  BOMEntityClass = "service_endpoint"
)

type BOMConfidence string

const (
	BOMConfidenceAuthoritative BOMConfidence = "authoritative"
	BOMConfidenceHigh          BOMConfidence = "high"
	BOMConfidenceMedium        BOMConfidence = "medium"
	BOMConfidenceLow           BOMConfidence = "low"
	BOMConfidenceUnknown       BOMConfidence = "unknown"
)

type BOMRelationType string

const (
	BOMRelationDependsOn BOMRelationType = "depends_on"
	BOMRelationRunsOn    BOMRelationType = "runs_on"
	BOMRelationHosts     BOMRelationType = "hosts"
	BOMRelationConnects  BOMRelationType = "connects_to"
)

type BOMRelationDriftCategory string

const (
	BOMRelationDriftMissingDependency BOMRelationDriftCategory = "missing_dependency"
	BOMRelationDriftStaleDependency   BOMRelationDriftCategory = "stale_dependency"
	BOMRelationDriftConfidenceDrop    BOMRelationDriftCategory = "confidence_drop"
)

type BOMDocument struct {
	APIVersion string  `json:"apiVersion"`
	Kind       string  `json:"kind"`
	Metadata   BOMMeta `json:"metadata"`
	Spec       BOMSpec `json:"spec"`
}

type BOMMeta struct {
	GeneratedAt   string `json:"generatedAt"`
	Source        string `json:"source"`
	CorrelationID string `json:"correlationId,omitempty"`
}

type BOMSpec struct {
	Entities  []BOMEntity      `json:"entities"`
	Relations []BOMRelation    `json:"relations"`
	Errors    []NormalizeError `json:"errors,omitempty"`
}

type BOMEntity struct {
	ID         string         `json:"id"`
	Class      BOMEntityClass `json:"class"`
	Name       string         `json:"name"`
	Version    string         `json:"version,omitempty"`
	Provenance string         `json:"provenance,omitempty"`
	Confidence BOMConfidence  `json:"confidence,omitempty"`
	ObservedAt string         `json:"observedAt,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

type BOMRelation struct {
	From       string          `json:"from"`
	To         string          `json:"to"`
	Type       BOMRelationType `json:"type"`
	Confidence BOMConfidence   `json:"confidence,omitempty"`
	ObservedAt string          `json:"observedAt,omitempty"`
	Attributes map[string]any  `json:"attributes,omitempty"`
}

type BOMRelationDrift struct {
	From         string                   `json:"from"`
	To           string                   `json:"to"`
	RelationType BOMRelationType          `json:"relationType"`
	Category     BOMRelationDriftCategory `json:"category"`
	Message      string                   `json:"message"`
}

type ReconciliationSnapshot struct {
	APIVersion string                     `json:"apiVersion"`
	Kind       string                     `json:"kind"`
	Metadata   ReconciliationSnapshotMeta `json:"metadata"`
	Spec       ReconciliationSnapshotSpec `json:"spec"`
}

type ReconciliationSnapshotMeta struct {
	GeneratedAt string `json:"generatedAt"`
	Source      string `json:"source"`
	Revision    int64  `json:"revision,omitempty"`
}

type ReconciliationSnapshotSpec struct {
	BOM            BOMDocument        `json:"bom"`
	DegradedNodes  []DegradedNode     `json:"degradedNodes"`
	FracturedEdges []FracturedEdge    `json:"fracturedEdges"`
	RelationDrifts []BOMRelationDrift `json:"relationDrifts"`
	NextActions    []string           `json:"nextActions"`
	Errors         []NormalizeError   `json:"errors,omitempty"`
}

func BuildBOMFromState(st *OmniGraphState, source string) BOMDocument {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if st == nil {
		return BOMDocument{
			APIVersion: BOMAPIVersion,
			Kind:       "BOM",
			Metadata:   BOMMeta{GeneratedAt: now, Source: source},
			Spec:       BOMSpec{Entities: []BOMEntity{}, Relations: []BOMRelation{}},
		}
	}
	entities := make([]BOMEntity, 0, len(st.Nodes))
	for _, n := range st.Nodes {
		entities = append(entities, BOMEntity{
			ID:         n.ID,
			Class:      classifyNodeKind(n.Kind),
			Name:       n.Label,
			Provenance: n.Provenance.PathHint,
			Confidence: BOMConfidenceMedium,
			ObservedAt: st.GeneratedAt,
			Attributes: n.Attributes,
		})
	}
	relations := make([]BOMRelation, 0, len(st.Edges))
	for _, e := range st.Edges {
		relations = append(relations, BOMRelation{
			From:       e.From,
			To:         e.To,
			Type:       classifyEdgeKind(e.Kind),
			Confidence: BOMConfidenceMedium,
			ObservedAt: st.GeneratedAt,
			Attributes: e.Attributes,
		})
	}
	return BOMDocument{
		APIVersion: BOMAPIVersion,
		Kind:       "BOM",
		Metadata: BOMMeta{
			GeneratedAt:   now,
			Source:        source,
			CorrelationID: st.CorrelationID,
		},
		Spec: BOMSpec{
			Entities:  entities,
			Relations: relations,
			Errors:    st.PartialErrors,
		},
	}
}

func BuildRelationDrifts(intended, runtime *OmniGraphState) []BOMRelationDrift {
	if intended == nil || runtime == nil {
		return nil
	}
	rtNode := make(map[string]struct{}, len(runtime.Nodes))
	for _, n := range runtime.Nodes {
		rtNode[n.ID] = struct{}{}
	}
	rtEdge := make(map[string]struct{}, len(runtime.Edges))
	for _, e := range runtime.Edges {
		rtEdge[e.From+"\x00"+e.To+"\x00"+e.Kind] = struct{}{}
	}
	var out []BOMRelationDrift
	for _, e := range intended.Edges {
		kind := classifyEdgeKind(e.Kind)
		if _, ok := rtNode[e.To]; !ok {
			out = append(out, BOMRelationDrift{
				From:         e.From,
				To:           e.To,
				RelationType: kind,
				Category:     BOMRelationDriftMissingDependency,
				Message:      "Dependency target is missing from runtime evidence.",
			})
			continue
		}
		if _, ok := rtNode[e.From]; !ok {
			out = append(out, BOMRelationDrift{
				From:         e.From,
				To:           e.To,
				RelationType: kind,
				Category:     BOMRelationDriftStaleDependency,
				Message:      "Dependency source is missing from runtime evidence.",
			})
			continue
		}
		key := e.From + "\x00" + e.To + "\x00" + e.Kind
		if _, ok := rtEdge[key]; !ok {
			out = append(out, BOMRelationDrift{
				From:         e.From,
				To:           e.To,
				RelationType: kind,
				Category:     BOMRelationDriftConfidenceDrop,
				Message:      "Dependency edge was not observed at runtime; confidence reduced.",
			})
		}
	}
	return out
}

func classifyNodeKind(k string) BOMEntityClass {
	switch k {
	case "host", "device", "baremetal", "switch", "router":
		return BOMEntityHardware
	case "service", "endpoint", "load_balancer", "database", "api":
		return BOMEntityService
	default:
		return BOMEntitySoftware
	}
}

func classifyEdgeKind(k string) BOMRelationType {
	switch k {
	case "runs_on":
		return BOMRelationRunsOn
	case "hosts":
		return BOMRelationHosts
	case "connects_to":
		return BOMRelationConnects
	default:
		return BOMRelationDependsOn
	}
}
