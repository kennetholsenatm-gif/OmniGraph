package security

import (
	"strings"
	"unicode"
)

// InventoryHost is a minimal Ansible INI host line.
type InventoryHost struct {
	Name string
	Host string
	User string
	Port string
}

// ParseAnsibleInventoryINI extracts hosts from Ansible inventory text (INI style).
// Lines like `name ansible_host=1.2.3.4 ansible_user=root ansible_port=22` are supported.
func ParseAnsibleInventoryINI(text string) []InventoryHost {
	var out []InventoryHost
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		first := fields[0]
		if strings.Contains(first, "=") {
			continue
		}
		h := InventoryHost{Name: first, Host: first}
		for _, f := range fields[1:] {
			k, v, ok := strings.Cut(f, "=")
			if !ok {
				continue
			}
			switch strings.TrimSpace(k) {
			case "ansible_host":
				h.Host = strings.TrimSpace(v)
			case "ansible_user":
				h.User = strings.TrimSpace(v)
			case "ansible_port":
				h.Port = strings.TrimSpace(v)
			}
		}
		out = append(out, h)
	}
	return out
}

// SanitizeFilename maps a host label to a safe file name fragment.
func SanitizeFilename(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "host"
	}
	return out
}
