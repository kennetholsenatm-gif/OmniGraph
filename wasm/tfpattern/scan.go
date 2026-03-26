package tfpattern

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"
)

// Finding is a lightweight policy hit for HCL-ish text (checkov-style subset, browser Wasm spike).
type Finding struct {
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	Detail   string `json:"detail,omitempty"`
	Line     int    `json:"line,omitempty"`
}

var (
	reQuotedSensitive = regexp.MustCompile(`(?i)\b(password|secret|api_key|access_key|secret_key|private_key)\s*=\s*"([^"]*)"`)
	reAKIA            = regexp.MustCompile(`AKIA[0-9A-Z]{16}`)
)

// Scan returns findings for lines that look like hardcoded secrets in Terraform-style HCL.
// Interpolation (${ … }) and short placeholders are ignored for the quoted-value rule.
func Scan(src []byte) []Finding {
	var out []Finding
	sc := bufio.NewScanner(bytes.NewReader(src))
	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := sc.Text()
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}
		if strings.Contains(line, "${") {
			continue
		}
		for _, m := range reQuotedSensitive.FindAllStringSubmatch(line, -1) {
			val := m[2]
			if len(val) < 6 {
				continue
			}
			out = append(out, Finding{
				Severity: "warning",
				Summary:  "possible hardcoded sensitive value",
				Detail:   "quoted literal for " + m[1] + "; prefer variables or a secret backend",
				Line:     lineNum,
			})
		}
		for range reAKIA.FindAllString(line, -1) {
			out = append(out, Finding{
				Severity: "error",
				Summary:  "AWS access key id pattern (AKIA…)",
				Detail:   "rotate if real; use IAM roles or vault-backed config",
				Line:     lineNum,
			})
		}
	}
	return out
}
