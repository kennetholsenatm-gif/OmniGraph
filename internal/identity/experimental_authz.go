package identity

import "strings"

// ExperimentalAuthorizer gates privileged serve routes. Static bootstrap tokens map to
// BootstrapSubjectID and receive all experimental permissions. OIDC subjects must pass
// an optional role gate (Keycloak realm/client roles or groups) before the same permission set applies.
type ExperimentalAuthorizer struct {
	StaticTokenConfigured bool
	RequiredOIDCRoles     []string
}

// Can implements Authorizer for experimental APIs only.
func (e *ExperimentalAuthorizer) Can(s Subject, perm string, _ string) bool {
	if strings.TrimSpace(s.ID) == "" {
		return false
	}
	if e.StaticTokenConfigured && s.ID == BootstrapSubjectID {
		return experimentalPermitted(perm)
	}
	if s.ID == BootstrapSubjectID {
		return false
	}
	if len(e.RequiredOIDCRoles) > 0 && !SubjectHasAnyRole(s, e.RequiredOIDCRoles) {
		return false
	}
	return experimentalPermitted(perm)
}

func experimentalPermitted(perm string) bool {
	switch perm {
	case PermServeSecurityScan, PermServeHostOpsRead, PermServeHostOpsWrite,
		PermServeInventoryRead, PermServeAuditRead, PermServeIngestLocal, PermServeSyncWS:
		return true
	default:
		return false
	}
}

// FlatRoleStrings merges groups, realm roles, and client roles for matching.
func FlatRoleStrings(s Subject) []string {
	var out []string
	out = append(out, s.Groups...)
	out = append(out, s.RealmRoles...)
	out = append(out, s.ClientRoles...)
	return out
}

// SubjectHasAnyRole reports whether the subject has at least one of required roles.
// An empty required list always matches (no role gate).
func SubjectHasAnyRole(s Subject, required []string) bool {
	if len(required) == 0 {
		return true
	}
	fr := FlatRoleStrings(s)
	for _, r := range required {
		for _, x := range fr {
			if x == r {
				return true
			}
		}
	}
	return false
}
