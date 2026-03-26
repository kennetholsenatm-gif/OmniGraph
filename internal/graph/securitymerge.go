package graph

import (
	"encoding/json"

	"github.com/kennetholsenatm-gif/omnigraph/internal/security"
)

// MergeSecurity attaches securityPosture summaries to matching host nodes (attributes.ansible_host or label vs security metadata.ansibleHost or target).
func MergeSecurity(d *Document, sec *security.Document) {
	if d == nil || sec == nil {
		return
	}
	key := sec.Metadata.AnsibleHost
	if key == "" {
		key = sec.Metadata.Target
	}
	if key == "" {
		return
	}
	summary := security.SummarizeForGraph(sec)
	raw, err := json.Marshal(summary)
	if err != nil {
		return
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return
	}
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
		if d.Spec.Nodes[i].Attributes == nil {
			d.Spec.Nodes[i].Attributes = map[string]any{}
		}
		d.Spec.Nodes[i].Attributes["securityPosture"] = payload
		if summary.Vulnerable > 0 && summary.HighOrCritical > 0 {
			d.Spec.Nodes[i].State = "attention"
		} else if summary.Vulnerable > 0 {
			d.Spec.Nodes[i].State = "review"
		}
	}
}
