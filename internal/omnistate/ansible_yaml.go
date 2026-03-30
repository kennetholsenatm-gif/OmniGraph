package omnistate

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
)

// AnsibleYAMLNormalizer parses Ansible YAML inventory (static structure).
type AnsibleYAMLNormalizer struct{}

func (AnsibleYAMLNormalizer) Kind() SourceKind { return SourceAnsibleYAML }

// Normalize decodes YAML and walks all/children/groups/hosts patterns.
func (AnsibleYAMLNormalizer) Normalize(ctx context.Context, in NormalizerInput) (OmniGraphStateFragment, error) {
	_ = ctx
	var fr OmniGraphStateFragment
	ref := in.Ref
	if ref.Type == "" {
		ref.Type = SourceAnsibleYAML
	}
	if ref.Name == "" {
		ref.Name = in.Name
	}
	var root any
	if err := yaml.Unmarshal(in.Data, &root); err != nil {
		fr.PartialErrors = append(fr.PartialErrors, NormalizeError{
			Path:    in.Name,
			Code:    "E_ANSIBLE_YAML_SYNTAX",
			Message: err.Error(),
		})
		return fr, nil
	}
	m, ok := root.(map[string]any)
	if !ok {
		fr.PartialErrors = append(fr.PartialErrors, NormalizeError{
			Path:    in.Name,
			Code:    "E_ANSIBLE_YAML_SHAPE",
			Message: "root must be a mapping",
		})
		return fr, nil
	}
	if all, ok := m["all"].(map[string]any); ok {
		yamlWalkGroup(ref, "all", all, &fr)
		return fr, nil
	}
	for gname, gv := range m {
		if gname == "" {
			continue
		}
		gm, ok := gv.(map[string]any)
		if !ok {
			continue
		}
		yamlWalkGroup(ref, gname, gm, &fr)
	}
	return fr, nil
}

func yamlWalkGroup(ref SourceRef, groupName string, gm map[string]any, fr *OmniGraphStateFragment) {
	gid := ansibleGroupID(ref, groupName)
	fr.Nodes = append(fr.Nodes, StateNode{
		ID:         gid,
		Kind:       "ansible_group",
		Label:      groupName,
		Attributes: map[string]any{"group": groupName},
		Provenance: ref,
	})
	if hosts, ok := gm["hosts"].(map[string]any); ok {
		for hname, hv := range hosts {
			attrs := map[string]any{"hostname": hname, "group": groupName}
			mergeYAMLHostVars(attrs, hv)
			hid := ansibleHostID(ref, groupName, hname)
			fr.Nodes = append(fr.Nodes, StateNode{
				ID:         hid,
				Kind:       "ansible_host",
				Label:      hname,
				Attributes: attrs,
				Provenance: ref,
			})
			fr.Edges = append(fr.Edges, StateEdge{
				From:       hid,
				To:         gid,
				Kind:       "member_of",
				Provenance: ref,
			})
		}
	}
	if children, ok := gm["children"].(map[string]any); ok {
		for cname, cv := range children {
			childMap, ok := cv.(map[string]any)
			if !ok {
				childMap = map[string]any{}
			}
			yamlWalkGroup(ref, cname, childMap, fr)
			cgid := ansibleGroupID(ref, cname)
			fr.Edges = append(fr.Edges, StateEdge{
				From:       cgid,
				To:         gid,
				Kind:       "child_group",
				Provenance: ref,
			})
		}
	}
}

func mergeYAMLHostVars(attrs map[string]any, hv any) {
	switch t := hv.(type) {
	case map[string]any:
		for k, v := range t {
			attrs[k] = v
		}
	case nil:
	default:
		attrs["value"] = fmt.Sprint(t)
	}
}
