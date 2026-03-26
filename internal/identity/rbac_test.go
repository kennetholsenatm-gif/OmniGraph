package identity

import "testing"

func TestStaticRBAC(t *testing.T) {
	a := &StaticRBAC{
		UserPermissions: map[string][]string{
			"alice": {PermServeInventoryRead, PermIRValidate},
		},
		AdminSubjects: map[string]struct{}{"root": {}},
	}
	if !a.Can(Subject{ID: "root"}, PermIREmit, "") {
		t.Fatal("admin should allow")
	}
	if !a.Can(Subject{ID: "alice"}, PermServeInventoryRead, "") {
		t.Fatal("alice inventory")
	}
	if a.Can(Subject{ID: "alice"}, PermIREmit, "") {
		t.Fatal("alice should not emit")
	}
}

func TestClaimMapperUnion(t *testing.T) {
	m := &ClaimMapper{
		GroupPermissions: map[string][]string{
			"ipausers":            {PermServeHealth},
			"omnigraph-operators": {PermServeInventoryRead, PermIRValidate},
		},
		RolePermissions: map[string][]string{
			"omnigraph-admin": {PermIREmit},
		},
	}
	subj := Subject{
		ID:         "bob",
		Groups:     []string{"ipausers", "omnigraph-operators"},
		RealmRoles: []string{"omnigraph-admin"},
	}
	got := m.PermissionsForSubject(subj)
	if len(got) != 4 {
		t.Fatalf("got %v", got)
	}
}
