package identity

import "testing"

func TestExperimentalAuthorizerBootstrap(t *testing.T) {
	a := &ExperimentalAuthorizer{StaticTokenConfigured: true}
	if !a.Can(Subject{ID: BootstrapSubjectID}, PermServeInventoryRead, "") {
		t.Fatal("bootstrap inventory")
	}
	if !a.Can(Subject{ID: BootstrapSubjectID}, PermServeWorkspaceDrift, "") {
		t.Fatal("bootstrap workspace drift")
	}
	if !a.Can(Subject{ID: BootstrapSubjectID}, PermServeIntegrationRun, "") {
		t.Fatal("bootstrap integration run")
	}
	if a.Can(Subject{ID: BootstrapSubjectID}, PermServeHealth, "") {
		t.Fatal("health not experimental")
	}
}

func TestExperimentalAuthorizerOIDCRoleGate(t *testing.T) {
	a := &ExperimentalAuthorizer{
		StaticTokenConfigured: false,
		RequiredOIDCRoles:     []string{"omnigraph-api"},
	}
	if a.Can(Subject{ID: "user-1"}, PermServeInventoryRead, "") {
		t.Fatal("missing role should deny")
	}
	if !a.Can(Subject{ID: "user-1", RealmRoles: []string{"omnigraph-api"}}, PermServeInventoryRead, "") {
		t.Fatal("with realm role allow")
	}
	if !a.Can(Subject{ID: "user-1", Groups: []string{"omnigraph-api"}}, PermServeAuditRead, "") {
		t.Fatal("with group allow")
	}
}
