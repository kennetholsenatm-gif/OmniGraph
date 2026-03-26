package identity

import (
	"strings"
)

// Authorizer answers permission checks for a subject and optional resource scope
// (e.g. stack key, repository id). Scope matching is prefix or equality per policy.
type Authorizer interface {
	Can(subj Subject, permission string, resourceScope string) bool
}

// StaticRBAC maps subject id -> permission set. Group and role expansion is done
// by upstream IdP adapters before calling into this authorizer.
type StaticRBAC struct {
	// UserPermissions is subject ID -> list of permissions.
	UserPermissions map[string][]string
	// AdminSubjects receive all permissions when non-nil check passes.
	AdminSubjects map[string]struct{}
}

// Can implements Authorizer.
func (s *StaticRBAC) Can(subj Subject, permission string, _ string) bool {
	if s == nil {
		return false
	}
	if len(s.AdminSubjects) > 0 {
		if _, ok := s.AdminSubjects[subj.ID]; ok {
			return true
		}
	}
	perms := s.UserPermissions[subj.ID]
	return stringSliceContains(perms, permission)
}

// ClaimMapper expands OIDC/LDAP-derived claims into effective permissions.
type ClaimMapper struct {
	// GroupPermissions: group name/CN -> permissions
	GroupPermissions map[string][]string
	// RolePermissions: realm or client role -> permissions
	RolePermissions map[string][]string
}

// PermissionsForSubject merges group and role grants (union, deduplicated).
func (m *ClaimMapper) PermissionsForSubject(subj Subject) []string {
	if m == nil {
		return nil
	}
	seen := make(map[string]struct{})
	var out []string
	add := func(p []string) {
		for _, x := range p {
			x = strings.TrimSpace(x)
			if x == "" {
				continue
			}
			if _, ok := seen[x]; ok {
				continue
			}
			seen[x] = struct{}{}
			out = append(out, x)
		}
	}
	for _, g := range subj.Groups {
		add(m.GroupPermissions[g])
	}
	for _, r := range subj.RealmRoles {
		add(m.RolePermissions[r])
	}
	for _, r := range subj.ClientRoles {
		add(m.RolePermissions[r])
	}
	return out
}

func stringSliceContains(hay []string, needle string) bool {
	for _, h := range hay {
		if h == needle {
			return true
		}
	}
	return false
}
