package graph

import (
	"testing"

	"github.com/kennetholsenatm-gif/omnigraph/internal/security"
)

func testSecurityDoc(host, generatedAt string, vuln, highCrit, errs int) *security.Document {
	var results []security.ModuleResult
	for range highCrit {
		results = append(results, security.ModuleResult{
			Status:   security.StatusVulnerable,
			Severity: security.SeverityHigh,
		})
	}
	for range vuln - highCrit {
		results = append(results, security.ModuleResult{
			Status:   security.StatusVulnerable,
			Severity: security.SeverityMedium,
		})
	}
	for range errs {
		results = append(results, security.ModuleResult{Status: security.StatusError})
	}
	return &security.Document{
		APIVersion: "omnigraph/security/v1",
		Kind:       "SecurityScan",
		Metadata: security.Metadata{
			GeneratedAt: generatedAt,
			Target:      host,
			AnsibleHost: host,
			Profile:     "test-profile",
		},
		Spec: security.DocumentSpec{
			Summary: security.ScanSummary{
				ModulesRun: len(results),
				Vulnerable: vuln,
				Errors:     errs,
			},
			Results: results,
		},
	}
}

func TestPostureSummaryToMapKeys(t *testing.T) {
	s := security.HostPostureSummary{
		GeneratedAt:    "2024-01-01T00:00:00Z",
		Vulnerable:     2,
		NotVulnerable:  5,
		Errors:         1,
		HighOrCritical: 1,
	}
	m := postureSummaryToMap(s)
	for _, k := range []string{"generatedAt", "vulnerable", "notVulnerable", "errors", "highOrCritical"} {
		if _, ok := m[k]; !ok {
			t.Fatalf("missing key %q in %v", k, m)
		}
	}
}

func TestMergeSecurity_singleDocument(t *testing.T) {
	d := &Document{Spec: GraphSpec{Nodes: []Node{
		{ID: "h1", Kind: "host", Label: "10.0.0.1", Attributes: map[string]any{"ansible_host": "10.0.0.1"}},
	}}}
	sec := testSecurityDoc("10.0.0.1", "2024-01-01T00:00:00Z", 2, 1, 0)
	MergeSecurity(d, sec)
	attr := d.Spec.Nodes[0].Attributes
	if attr == nil || attr[attrSecurityPosture] == nil {
		t.Fatal("missing securityPosture")
	}
	if d.Spec.Nodes[0].State != "attention" {
		t.Fatalf("state %q", d.Spec.Nodes[0].State)
	}
}

func TestMergeSecurityDocuments_orderIndependentStrongerWins(t *testing.T) {
	weak := testSecurityDoc("10.0.0.1", "2024-01-01T00:00:00Z", 1, 0, 0)
	strong := testSecurityDoc("10.0.0.1", "2024-01-02T00:00:00Z", 3, 2, 0)

	d1 := &Document{Spec: GraphSpec{Nodes: []Node{
		{ID: "h1", Kind: "host", Label: "10.0.0.1"},
	}}}
	MergeSecurityDocuments(d1, []*security.Document{weak, strong}, MergeSecurityOptions{})

	d2 := &Document{Spec: GraphSpec{Nodes: []Node{
		{ID: "h1", Kind: "host", Label: "10.0.0.1"},
	}}}
	MergeSecurityDocuments(d2, []*security.Document{strong, weak}, MergeSecurityOptions{})

	m1 := d1.Spec.Nodes[0].Attributes[attrSecurityPosture].(map[string]any)
	m2 := d2.Spec.Nodes[0].Attributes[attrSecurityPosture].(map[string]any)
	if intField(m1, "vulnerable") != intField(m2, "vulnerable") {
		t.Fatalf("order-dependent: %v vs %v", m1, m2)
	}
	if intField(m1, "vulnerable") != 3 {
		t.Fatalf("want stronger scan vulnerable=3 got %v", m1)
	}
}

func TestMergeSecurityDocuments_callbackDecisions(t *testing.T) {
	weak := testSecurityDoc("10.0.0.1", "2024-01-01T00:00:00Z", 1, 0, 0)
	strong := testSecurityDoc("10.0.0.1", "2024-01-02T00:00:00Z", 2, 0, 0)

	var reasons []string
	opts := MergeSecurityOptions{
		OnDecision: func(hostKey, nodeID string, kept, discarded security.HostPostureSummary, reason string) {
			reasons = append(reasons, reason)
		},
	}
	d := &Document{Spec: GraphSpec{Nodes: []Node{
		{ID: "h1", Kind: "host", Label: "10.0.0.1"},
	}}}
	MergeSecurityDocuments(d, []*security.Document{weak, strong}, opts)
	if len(reasons) != 2 {
		t.Fatalf("reasons %v", reasons)
	}
	if reasons[0] != "initial" || reasons[1] != "skipped_incoming_lower_precedence" {
		t.Fatalf("got %v", reasons)
	}
}

func TestMergeSecurity_stableNodeOrder(t *testing.T) {
	sec := testSecurityDoc("10.0.0.1", "2024-01-01T00:00:00Z", 1, 0, 0)
	d := &Document{Spec: GraphSpec{Nodes: []Node{
		{ID: "z", Kind: "host", Label: "10.0.0.1"},
		{ID: "a", Kind: "host", Label: "10.0.0.1"},
	}}}
	var ids []string
	opts := MergeSecurityOptions{
		OnDecision: func(hostKey, nodeID string, kept, discarded security.HostPostureSummary, reason string) {
			ids = append(ids, nodeID)
		},
	}
	MergeSecurityWithOptions(d, sec, opts)
	if len(ids) != 2 || ids[0] != "a" || ids[1] != "z" {
		t.Fatalf("visit order %v want a,z", ids)
	}
}

func TestComparePostureDominance_generatedAt(t *testing.T) {
	a := security.HostPostureSummary{Vulnerable: 1, GeneratedAt: "2024-06-01T00:00:00Z"}
	b := security.HostPostureSummary{Vulnerable: 1, GeneratedAt: "2024-01-01T00:00:00Z"}
	if comparePostureDominance(a, b, security.Metadata{}, security.Metadata{}) != 1 {
		t.Fatal("newer should dominate")
	}
}
