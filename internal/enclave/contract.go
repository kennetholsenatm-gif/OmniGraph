package enclave

import (
	"fmt"
	"regexp"
	"strings"
)

var contractNameRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.-]*$`)

// ValidateContract enforces explicit requires/provides and peer/env boundary rules.
func ValidateContract(e *Enclave) error {
	if e == nil {
		return fmt.Errorf("nil enclave")
	}
	if err := validateNameList("spec.requires", e.Spec.Requires, true); err != nil {
		return err
	}
	if err := validateNameList("spec.provides", e.Spec.Provides, true); err != nil {
		return err
	}
	allowed := make(map[string]struct{})
	for _, s := range e.Spec.Requires {
		allowed[s] = struct{}{}
	}
	for _, s := range e.Spec.Provides {
		allowed[s] = struct{}{}
	}
	for _, peer := range e.Spec.TrustBoundary.AllowedPeers {
		peer = strings.TrimSpace(peer)
		if peer == "" {
			continue
		}
		if _, ok := allowed[peer]; !ok {
			return fmt.Errorf("spec.trustBoundary.allowedPeers: peer %q must appear in spec.requires or spec.provides", peer)
		}
	}
	for k, v := range e.Spec.Environment {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if ok, name := crossEnclaveRefInValue(v); ok {
			if name == "" {
				return fmt.Errorf("spec.environment[%q]: cross-enclave URI %q must name a peer", k, v)
			}
			if _, allowedName := allowed[name]; !allowedName {
				return fmt.Errorf("spec.environment[%q]: cross-enclave reference %q (peer %q) is not listed in requires/provides", k, v, name)
			}
		}
	}
	return nil
}

// crossEnclaveRefInValue reports whether value uses a peer:// or enclave:// URI and returns the declared peer name.
func crossEnclaveRefInValue(value string) (ok bool, name string) {
	lower := strings.ToLower(value)
	for _, prefix := range []string{"peer://", "enclave://"} {
		idx := strings.Index(lower, prefix)
		if idx < 0 {
			continue
		}
		rest := value[idx+len(prefix):]
		rest = strings.TrimPrefix(rest, "/")
		tok := rest
		if i := strings.IndexAny(tok, "/?#"); i >= 0 {
			tok = tok[:i]
		}
		tok = strings.TrimSpace(tok)
		if tok == "" {
			return true, ""
		}
		return true, tok
	}
	return false, ""
}

func validateNameList(field string, names []string, requireNonEmpty bool) error {
	if requireNonEmpty && len(names) == 0 {
		return fmt.Errorf("%s: at least one name is required", field)
	}
	seen := make(map[string]struct{}, len(names))
	for _, raw := range names {
		s := strings.TrimSpace(raw)
		if s == "" {
			return fmt.Errorf("%s: empty entry is not allowed", field)
		}
		if !contractNameRe.MatchString(s) {
			return fmt.Errorf("%s: invalid identifier %q (use [a-zA-Z0-9_.-], start with a letter)", field, s)
		}
		if _, dup := seen[s]; dup {
			return fmt.Errorf("%s: duplicate %q", field, s)
		}
		seen[s] = struct{}{}
	}
	return nil
}
