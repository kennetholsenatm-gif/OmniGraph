package security

import (
	"encoding/json"
	"fmt"
	"time"
)

// NewDocument builds a security/v1 document from results.
// ansibleHost is optional; when set, graph merge matches host node attributes.ansible_host.
func NewDocument(target, ansibleHost, transport, profile string, results []ModuleResult) *Document {
	now := time.Now().UTC().Format(time.RFC3339)
	sum := ScanSummary{}
	for _, r := range results {
		sum.ModulesRun++
		switch r.Status {
		case StatusVulnerable:
			sum.Vulnerable++
		case StatusNotVulnerable:
			sum.NotVulnerable++
		case StatusError:
			sum.Errors++
		case StatusNotApplicable:
			sum.NotApplicable++
		default:
			sum.Errors++
		}
	}
	return &Document{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata: Metadata{
			GeneratedAt: now,
			Target:      target,
			AnsibleHost: ansibleHost,
			Transport:   transport,
			Profile:     profile,
			Disclaimer:  defaultDisclaimer,
		},
		Spec: DocumentSpec{
			Summary: sum,
			Results: results,
		},
	}
}

// EncodeIndent returns indented JSON.
func EncodeIndent(d *Document) ([]byte, error) {
	if d == nil {
		return nil, fmt.Errorf("nil document")
	}
	return json.MarshalIndent(d, "", "  ")
}

// HostPostureSummary is merged into graph host attributes (see graph.MergeSecurity).
type HostPostureSummary struct {
	GeneratedAt    string `json:"generatedAt"`
	Vulnerable     int    `json:"vulnerable"`
	NotVulnerable  int    `json:"notVulnerable"`
	Errors         int    `json:"errors"`
	HighOrCritical int    `json:"highOrCritical,omitempty"`
}

// SummarizeForGraph extracts a compact summary for graph node attributes.
func SummarizeForGraph(d *Document) HostPostureSummary {
	if d == nil {
		return HostPostureSummary{}
	}
	hc := 0
	for _, r := range d.Spec.Results {
		if r.Status == StatusVulnerable && (r.Severity == SeverityHigh || r.Severity == SeverityCritical) {
			hc++
		}
	}
	return HostPostureSummary{
		GeneratedAt:    d.Metadata.GeneratedAt,
		Vulnerable:     d.Spec.Summary.Vulnerable,
		NotVulnerable:  d.Spec.Summary.NotVulnerable,
		Errors:         d.Spec.Summary.Errors,
		HighOrCritical: hc,
	}
}
