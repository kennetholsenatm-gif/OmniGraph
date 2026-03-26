package inventory

import (
	"sort"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/state"
)

// BuildINI renders an Ansible inventory from extracted host mappings (name -> ansible_host).
func BuildINI(hosts map[string]string) string {
	if len(hosts) == 0 {
		return "[omnigraph]\n"
	}
	keys := make([]string, 0, len(hosts))
	for k := range hosts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("[omnigraph]\n")
	for _, k := range keys {
		safe := sanitizeName(k)
		b.WriteString(safe)
		b.WriteString(" ansible_host=")
		b.WriteString(hosts[k])
		b.WriteByte('\n')
	}
	return b.String()
}

// FromStateFile loads tfstate JSON and builds an inventory using state.ExtractHosts.
func FromStateFile(path string) (string, error) {
	st, err := state.Load(path)
	if err != nil {
		return "", err
	}
	return BuildINI(state.ExtractHosts(st)), nil
}

// SanitizeHostKey normalizes a logical host name for Ansible INI keys (same rules as BuildINI).
func SanitizeHostKey(name string) string {
	return sanitizeName(name)
}

func sanitizeName(name string) string {
	// Ansible host names should be simple identifiers; mangle addresses.
	s := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
			return r
		default:
			return '_'
		}
	}, name)
	if s == "" {
		return "host"
	}
	return s
}
