package graph

import (
	"fmt"
	"sort"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/security"
)

const attrSecurityPosture = "securityPosture"
const attrSecurityPostureMergeMeta = "securityPostureMergeMeta"

// MergeSecurityOptions configures merge behavior and optional decision callbacks.
type MergeSecurityOptions struct {
	// OnDecision is optional. When non-nil, invoked after each merge decision per node.
	// kept is the summary retained on the node; discarded is the other candidate (incoming when skipped, previous when replaced).
	// reason is one of: "initial", "replaced", "skipped_incoming_lower_precedence".
	OnDecision func(hostKey, nodeID string, kept, discarded security.HostPostureSummary, reason string)
}

// MergeSecurity attaches securityPosture summaries to matching host nodes (attributes.ansible_host or label vs security metadata.ansibleHost or target).
// It delegates to MergeSecurityWithOptions with empty options.
func MergeSecurity(d *Document, sec *security.Document) {
	MergeSecurityWithOptions(d, sec, MergeSecurityOptions{})
}

// MergeSecurityWithOptions merges one security document using deterministic host matching and optional callbacks.
func MergeSecurityWithOptions(d *Document, sec *security.Document, opts MergeSecurityOptions) {
	if sec == nil {
		return
	}
	MergeSecurityDocuments(d, []*security.Document{sec}, opts)
}

// MergeSecurityDocuments merges multiple security documents in an order independent of the input slice:
// documents are sorted by descending posture precedence (see comparePostureDominance), then each is applied per host node;
// a newer scan replaces an existing one only if it strictly dominates per the same hierarchy.
func MergeSecurityDocuments(d *Document, docs []*security.Document, opts MergeSecurityOptions) {
	if d == nil {
		return
	}
	list := make([]*security.Document, 0, len(docs))
	for _, doc := range docs {
		if doc != nil {
			list = append(list, doc)
		}
	}
	if len(list) == 0 {
		return
	}
	sort.SliceStable(list, func(i, j int) bool {
		return documentSortPrecedence(list[i], list[j])
	})
	for _, doc := range list {
		mergeSecurityOneDocument(d, doc, opts)
	}
}

// documentSortPrecedence returns true if a should be processed before b (a is strictly stronger or tie-break smaller Target/Profile).
func documentSortPrecedence(a, b *security.Document) bool {
	if a == nil {
		return false
	}
	if b == nil {
		return true
	}
	sa := security.SummarizeForGraph(a)
	sb := security.SummarizeForGraph(b)
	c := comparePostureDominance(sa, sb, a.Metadata, b.Metadata)
	if c > 0 {
		return true
	}
	if c < 0 {
		return false
	}
	if a.Metadata.Target != b.Metadata.Target {
		return a.Metadata.Target < b.Metadata.Target
	}
	return a.Metadata.Profile < b.Metadata.Profile
}

// comparePostureDominance returns +1 if a should win over b, -1 if b should win over a, 0 if tied on all keys (caller may tie-break).
//
// Resolution hierarchy (deterministic):
//  1. Higher HighOrCritical
//  2. Higher Vulnerable
//  3. Higher Errors
//  4. Newer GeneratedAt (time.RFC3339); if either parse fails, lexicographic string compare
//  5. Lexicographic Profile
//  6. Lexicographic Target
func comparePostureDominance(a, b security.HostPostureSummary, metaA, metaB security.Metadata) int {
	switch {
	case a.HighOrCritical > b.HighOrCritical:
		return 1
	case a.HighOrCritical < b.HighOrCritical:
		return -1
	}
	switch {
	case a.Vulnerable > b.Vulnerable:
		return 1
	case a.Vulnerable < b.Vulnerable:
		return -1
	}
	switch {
	case a.Errors > b.Errors:
		return 1
	case a.Errors < b.Errors:
		return -1
	}
	tcmp := compareGeneratedAt(a.GeneratedAt, b.GeneratedAt)
	if tcmp != 0 {
		return tcmp
	}
	switch {
	case metaA.Profile > metaB.Profile:
		return 1
	case metaA.Profile < metaB.Profile:
		return -1
	}
	switch {
	case metaA.Target > metaB.Target:
		return 1
	case metaA.Target < metaB.Target:
		return -1
	}
	return 0
}

func compareGeneratedAt(a, b string) int {
	ta, errA := time.Parse(time.RFC3339, a)
	tb, errB := time.Parse(time.RFC3339, b)
	if errA == nil && errB == nil {
		switch {
		case ta.After(tb):
			return 1
		case ta.Before(tb):
			return -1
		default:
			return 0
		}
	}
	switch {
	case a > b:
		return 1
	case a < b:
		return -1
	default:
		return 0
	}
}

func hostMatchKey(sec *security.Document) string {
	if sec == nil {
		return ""
	}
	key := sec.Metadata.AnsibleHost
	if key == "" {
		key = sec.Metadata.Target
	}
	return key
}

func mergeSecurityOneDocument(d *Document, sec *security.Document, opts MergeSecurityOptions) {
	key := hostMatchKey(sec)
	if key == "" {
		return
	}
	incoming := security.SummarizeForGraph(sec)
	metaIn := sec.Metadata

	indices := matchingHostNodeIndices(d, key)
	if len(indices) == 0 {
		return
	}
	sort.Slice(indices, func(i, j int) bool {
		return d.Spec.Nodes[indices[i]].ID < d.Spec.Nodes[indices[j]].ID
	})

	for _, i := range indices {
		node := &d.Spec.Nodes[i]
		hostKey := key
		nodeID := node.ID
		prevSum, prevMeta, had := readStoredPosture(node.Attributes)
		if !had {
			applyPostureToNode(node, incoming, metaIn)
			if opts.OnDecision != nil {
				opts.OnDecision(hostKey, nodeID, incoming, security.HostPostureSummary{}, "initial")
			}
			continue
		}
		if comparePostureDominance(incoming, prevSum, metaIn, prevMeta) > 0 {
			applyPostureToNode(node, incoming, metaIn)
			if opts.OnDecision != nil {
				opts.OnDecision(hostKey, nodeID, incoming, prevSum, "replaced")
			}
			continue
		}
		if opts.OnDecision != nil {
			opts.OnDecision(hostKey, nodeID, prevSum, incoming, "skipped_incoming_lower_precedence")
		}
	}
}

func matchingHostNodeIndices(d *Document, key string) []int {
	var out []int
	for i := range d.Spec.Nodes {
		if d.Spec.Nodes[i].Kind != "host" {
			continue
		}
		attr := d.Spec.Nodes[i].Attributes
		ah := ""
		if attr != nil {
			if v, ok := attr["ansible_host"].(string); ok {
				ah = v
			}
		}
		lb := d.Spec.Nodes[i].Label
		if ah != key && lb != key {
			continue
		}
		out = append(out, i)
	}
	return out
}

func readStoredPosture(attr map[string]any) (security.HostPostureSummary, security.Metadata, bool) {
	if attr == nil {
		return security.HostPostureSummary{}, security.Metadata{}, false
	}
	raw, ok := attr[attrSecurityPosture]
	if !ok || raw == nil {
		return security.HostPostureSummary{}, security.Metadata{}, false
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return security.HostPostureSummary{}, security.Metadata{}, false
	}
	sum := security.HostPostureSummary{
		GeneratedAt:    stringField(m, "generatedAt"),
		Vulnerable:     intField(m, "vulnerable"),
		NotVulnerable:  intField(m, "notVulnerable"),
		Errors:         intField(m, "errors"),
		HighOrCritical: intField(m, "highOrCritical"),
	}
	var meta security.Metadata
	if mm, ok := attr[attrSecurityPostureMergeMeta].(map[string]any); ok {
		meta.Profile = stringField(mm, "profile")
		meta.Target = stringField(mm, "target")
		meta.AnsibleHost = stringField(mm, "ansibleHost")
	}
	return sum, meta, true
}

func stringField(m map[string]any, k string) string {
	v, ok := m[k]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

func intField(m map[string]any, k string) int {
	v, ok := m[k]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	default:
		return 0
	}
}

func postureSummaryToMap(s security.HostPostureSummary) map[string]any {
	m := map[string]any{
		"generatedAt":   s.GeneratedAt,
		"vulnerable":    s.Vulnerable,
		"notVulnerable": s.NotVulnerable,
		"errors":        s.Errors,
	}
	if s.HighOrCritical != 0 {
		m["highOrCritical"] = s.HighOrCritical
	}
	return m
}

func mergeMetaToMap(meta security.Metadata) map[string]any {
	return map[string]any{
		"profile":     meta.Profile,
		"target":      meta.Target,
		"ansibleHost": meta.AnsibleHost,
	}
}

func applyPostureToNode(node *Node, summary security.HostPostureSummary, meta security.Metadata) {
	if node.Attributes == nil {
		node.Attributes = map[string]any{}
	}
	node.Attributes[attrSecurityPosture] = postureSummaryToMap(summary)
	node.Attributes[attrSecurityPostureMergeMeta] = mergeMetaToMap(meta)
	if summary.Vulnerable > 0 && summary.HighOrCritical > 0 {
		node.State = "attention"
	} else if summary.Vulnerable > 0 {
		node.State = "review"
	}
}
