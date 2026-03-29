package inventory

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

var ansibleHostLine = regexp.MustCompile(`^([a-zA-Z0-9_.-]+)\s*(.*)$`)

// ParseAnsibleINIHosts extracts host keys from an Ansible INI-style inventory (bytes).
// Matches the Python omnigraph auto_discover behavior: INI sections (excluding group_* and DEFAULT),
// then line-based hosts if no keys found.
func ParseAnsibleINIHosts(text []byte) ([]string, error) {
	if len(text) == 0 {
		return nil, nil
	}
	cfg, err := ini.LoadSources(ini.LoadOptions{
		AllowBooleanKeys:   true,
		InsensitiveSections: true,
		InsensitiveKeys:    true,
		IgnoreInlineComment: false,
	}, text)
	if err != nil {
		return nil, fmt.Errorf("ini: %w", err)
	}
	var hosts []string
	for _, section := range cfg.Sections() {
		name := section.Name()
		if strings.EqualFold(name, ini.DefaultSection) {
			continue
		}
		if strings.HasPrefix(strings.ToLower(name), "group_") {
			continue
		}
		for _, k := range section.Keys() {
			key := strings.TrimSpace(k.Name())
			if key == "" || strings.HasPrefix(key, ";") {
				continue
			}
			if i := strings.IndexByte(key, ' '); i >= 0 {
				key = strings.TrimSpace(key[:i])
			}
			if key == "" || strings.HasPrefix(strings.ToLower(key), "ansible_") {
				continue
			}
			hosts = append(hosts, key)
		}
	}
	if len(hosts) == 0 {
		for _, line := range strings.Split(string(text), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
				continue
			}
			if m := ansibleHostLine.FindStringSubmatch(line); m != nil {
				hosts = append(hosts, m[1])
			}
		}
	}
	return dedupe(hosts), nil
}

// ParseAnsibleYAMLHosts extracts host names from a YAML inventory (best-effort).
func ParseAnsibleYAMLHosts(text []byte) ([]string, error) {
	var root any
	if err := yaml.Unmarshal(text, &root); err != nil {
		return nil, fmt.Errorf("yaml: %w", err)
	}
	var hosts []string
	switch v := root.(type) {
	case map[string]any:
		if h, ok := v["hosts"].([]any); ok {
			for _, x := range h {
				if s, ok := x.(string); ok && s != "" {
					hosts = append(hosts, s)
				}
			}
		}
		if all, ok := v["all"].(map[string]any); ok {
			for _, key := range []string{"children", "hosts"} {
				if ch, ok := all[key].([]any); ok {
					for _, x := range ch {
						if s, ok := x.(string); ok && s != "" {
							hosts = append(hosts, s)
						}
					}
				}
			}
		}
	case []any:
		for _, x := range v {
			if s, ok := x.(string); ok && s != "" {
				hosts = append(hosts, s)
			}
		}
	}
	return dedupe(hosts), nil
}

func dedupe(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
