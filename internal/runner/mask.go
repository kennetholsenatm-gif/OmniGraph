package runner

import (
	"bytes"
	"slices"
)

const maxRedactSecretLen = 4096

// Redact replaces non-overlapping occurrences of each secret substring in data with repl.
// Secrets are applied longest-first so longer values mask before shorter prefixes.
// Empty strings are skipped; values longer than maxRedactSecretLen are truncated for matching only.
func Redact(data []byte, secrets []string, repl []byte) []byte {
	if len(data) == 0 || len(secrets) == 0 {
		return data
	}
	if len(repl) == 0 {
		repl = []byte("***")
	}
	uniq := make(map[string]struct{})
	var sorted []string
	for _, s := range secrets {
		if s == "" {
			continue
		}
		if len(s) > maxRedactSecretLen {
			s = s[:maxRedactSecretLen]
		}
		if _, ok := uniq[s]; ok {
			continue
		}
		uniq[s] = struct{}{}
		sorted = append(sorted, s)
	}
	slices.SortFunc(sorted, func(a, b string) int {
		if len(a) != len(b) {
			return len(b) - len(a)
		}
		return bytes.Compare([]byte(a), []byte(b))
	})
	out := append([]byte(nil), data...)
	for _, sec := range sorted {
		pat := []byte(sec)
		if len(pat) == 0 {
			continue
		}
		out = bytes.ReplaceAll(out, pat, repl)
	}
	return out
}

// secretValuesForStep returns env values plus RedactExtra for log redaction.
func secretValuesForStep(s Step) []string {
	n := len(s.Env) + len(s.RedactExtra)
	if n == 0 {
		return nil
	}
	out := make([]string, 0, n)
	for _, v := range s.Env {
		out = append(out, v)
	}
	out = append(out, s.RedactExtra...)
	return out
}
