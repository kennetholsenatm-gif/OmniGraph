package identity

// FreeIPA is typically fronted as LDAP (389 Directory Server) and Kerberos. ADR 005 treats
// FreeIPA as a directory source for groups and POSIX attributes.
//
// Implementation outline for a future package (e.g. internal/identity/ldapdir):
//   - LDAPS or StartTLS connection to FreeIPA servers.
//   - Service bind DN with minimal ACL: read group membership for users authenticating via OIDC
//     (lookup by uid) or bind-as-user flows where policy allows.
//   - Map memberOf / ipausergroup values into Subject.Groups for ClaimMapper.
//
// Kerberos (SPNEGO) is optional at the edge (reverse proxy) rather than inside omnigraph core.

// LDAPDirectoryConfig holds connection parameters for FreeIPA LDAP.
type LDAPDirectoryConfig struct {
	URL          string // ldaps://ipa.example.com:636
	BindDN       string
	BindPassword string // prefer env or secret file reference in real deployments
	BaseDN       string // e.g. dc=example,dc=com
	UserFilter   string // e.g. (uid=%s) — parameterized at runtime
}
