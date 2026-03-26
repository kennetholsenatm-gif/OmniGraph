package identity

// Keycloak integration (OIDC / OAuth2) is specified in ADR 005 and docs/integrations.md.
//
// Implementation outline for a future package (e.g. internal/identity/keycloakjwt):
//   - Discover OIDC metadata (issuer URL).
//   - Validate JWT signature and standard claims (iss, aud, exp).
//   - Map realm_access.roles and resource_access[client].roles into Subject.RealmRoles / ClientRoles.
//   - Optional: map "groups" claim or group mapper into Subject.Groups.
//
// The serve binary should accept --oidc-issuer and --oidc-audience (or config file) and
// construct Subject on each request before invoking Authorizer.

// OIDCConfig holds deployment-time settings for Keycloak (or any OIDC provider).
type OIDCConfig struct {
	IssuerURL     string
	ClientID      string // expected "aud" when validating access tokens
	SkipTLSVerify bool   // dev only; must be false in production
}
