package identity

// BootstrapSubjectID is the synthetic subject id for requests authenticated with the
// static OMNIGRAPH_SERVE_TOKEN / --auth-token (migration and CI bootstrap).
const BootstrapSubjectID = "bootstrap-token"

// Subject is an authenticated principal (human or service). Values are opaque outside AuthZ.
type Subject struct {
	ID          string   // stable id (sub, uid, service account name)
	DisplayName string   // optional
	Groups      []string // directory groups or OIDC "groups"
	RealmRoles  []string // Keycloak realm roles, if applicable
	ClientRoles []string // Keycloak client roles, if applicable
}
