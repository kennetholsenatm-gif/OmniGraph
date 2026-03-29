package omnistate

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
)

// AnsibleININormalizer parses Ansible INI-style inventory text.
type AnsibleININormalizer struct{}

func (AnsibleININormalizer) Kind() SourceKind { return SourceAnsibleINI }

// Normalize parses [group] sections and host lines into ansible_group / ansible_host nodes.
func (AnsibleININormalizer) Normalize(ctx context.Context, in NormalizerInput) (OmniGraphStateFragment, error) {
	_ = ctx
	var fr OmniGraphStateFragment
	ref := in.Ref
	if ref.Type == "" {
		ref.Type = SourceAnsibleINI
	}
	if ref.Name == "" {
		ref.Name = in.Name
	}
	if len(bytes.TrimSpace(in.Data)) == 0 {
		fr.PartialErrors = append(fr.PartialErrors, NormalizeError{
			Path:    in.Name,
			Code:    "E_ANSIBLE_INI_EMPTY",
			Message: "empty inventory",
		})
		return fr, nil
	}
	sc := bufio.NewScanner(bytes.NewReader(in.Data))
	var group string
	lineNo := 0
	for sc.Scan() {
		lineNo++
		if err := sc.Err(); err != nil {
			fr.PartialErrors = append(fr.PartialErrors, NormalizeError{
				Path:    in.Name,
				Code:    "E_ANSIBLE_INI_READ",
				Message: err.Error(),
			})
			break
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			group = strings.TrimSpace(line[1 : len(line)-1])
			// skip :children / :vars markers for structure — still record group node
			baseGroup := group
			if i := strings.Index(group, ":"); i > 0 {
				suffix := group[i+1:]
				if suffix == "children" || suffix == "vars" {
					baseGroup = group[:i]
				}
			}
			if baseGroup != "" {
				gid := ansibleGroupID(ref, baseGroup)
				fr.Nodes = append(fr.Nodes, StateNode{
					ID:         gid,
					Kind:       "ansible_group",
					Label:      baseGroup,
					Attributes: map[string]any{"group": baseGroup},
					Provenance: ref,
				})
			}
			continue
		}
		hostPart, vars := splitIniHostLine(line)
		if hostPart == "" {
			continue
		}
		if group == "" {
			group = "ungrouped"
			gid := ansibleGroupID(ref, group)
			fr.Nodes = append(fr.Nodes, StateNode{
				ID:         gid,
				Kind:       "ansible_group",
				Label:      group,
				Attributes: map[string]any{"group": group},
				Provenance: ref,
			})
		}
		baseGroup := group
		if i := strings.Index(group, ":"); i > 0 {
			if suf := group[i+1:]; suf == "children" || suf == "vars" {
				baseGroup = group[:i]
			}
		}
		hid := ansibleHostID(ref, baseGroup, hostPart)
		attrs := map[string]any{"hostname": hostPart, "group": baseGroup}
		for k, v := range vars {
			attrs[k] = v
		}
		fr.Nodes = append(fr.Nodes, StateNode{
			ID:         hid,
			Kind:       "ansible_host",
			Label:      hostPart,
			Attributes: attrs,
			Provenance: ref,
		})
		gid := ansibleGroupID(ref, baseGroup)
		fr.Edges = append(fr.Edges, StateEdge{
			From:       hid,
			To:         gid,
			Kind:       "member_of",
			Provenance: ref,
		})
	}
	return fr, nil
}

func ansibleGroupID(ref SourceRef, group string) string {
	return fmt.Sprintf("ansible:group:%s:%s", sanitizeKey(ref.Name), sanitizeKey(group))
}

func ansibleHostID(ref SourceRef, group, host string) string {
	return fmt.Sprintf("ansible:host:%s:%s:%s", sanitizeKey(ref.Name), sanitizeKey(group), sanitizeKey(host))
}

func sanitizeKey(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-', r == '.':
			return r
		default:
			return '_'
		}
	}, s)
	if s == "" {
		return "x"
	}
	return s
}

func splitIniHostLine(line string) (host string, vars map[string]string) {
	vars = make(map[string]string)
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", vars
	}
	host = fields[0]
	for _, f := range fields[1:] {
		if idx := strings.IndexByte(f, '='); idx > 0 {
			k := strings.TrimSpace(f[:idx])
			v := strings.TrimSpace(f[idx+1:])
			vars[k] = v
		}
	}
	return host, vars
}
