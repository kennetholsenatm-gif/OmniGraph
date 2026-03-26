package serve

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/kennetholsenatm-gif/omnigraph/internal/identity"
)

type ctxKey int

const subjectCtxKey ctxKey = 1

// SubjectFromRequest returns the authenticated subject when requirePermission ran successfully.
func SubjectFromRequest(r *http.Request) (identity.Subject, bool) {
	s, ok := r.Context().Value(subjectCtxKey).(identity.Subject)
	return s, ok
}

// keycloakClaims captures common Keycloak JWT claims for RBAC mapping.
type keycloakClaims struct {
	PreferredUsername string              `json:"preferred_username"`
	Groups            []string            `json:"groups"`
	RealmAccess       realmAccess         `json:"realm_access"`
	ResourceAccess    map[string]roleList `json:"resource_access"`
}

type realmAccess struct {
	Roles []string `json:"roles"`
}

type roleList struct {
	Roles []string `json:"roles"`
}

func subjectFromIDToken(tok *oidc.IDToken, clientID string) (identity.Subject, error) {
	var c keycloakClaims
	if err := tok.Claims(&c); err != nil {
		return identity.Subject{}, err
	}
	out := identity.Subject{
		ID:          tok.Subject,
		DisplayName: c.PreferredUsername,
		Groups:      append([]string{}, c.Groups...),
		RealmRoles:  append([]string{}, c.RealmAccess.Roles...),
	}
	if clientID != "" {
		if ra, ok := c.ResourceAccess[clientID]; ok {
			out.ClientRoles = append(out.ClientRoles, ra.Roles...)
		}
	}
	return out, nil
}

func bearerRaw(r *http.Request) string {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if h == "" {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer"))
}

func (s *server) authenticate(r *http.Request) (identity.Subject, error) {
	raw := bearerRaw(r)
	if raw == "" {
		return identity.Subject{}, errors.New("missing bearer token")
	}
	if s.oidcVerifier != nil && strings.Count(raw, ".") == 2 {
		idTok, err := s.oidcVerifier.Verify(r.Context(), raw)
		if err == nil {
			return subjectFromIDToken(idTok, s.oidcClientID)
		}
		// Invalid JWT: fall through to static token match.
	}
	if s.authToken != "" && raw == s.authToken {
		return identity.Subject{ID: identity.BootstrapSubjectID}, nil
	}
	return identity.Subject{}, errors.New("unauthorized")
}

func (s *server) requirePermission(perm string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subj, err := s.authenticate(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if s.authz == nil || !s.authz.Can(subj, perm, "") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), subjectCtxKey, subj)
		next(w, r.WithContext(ctx))
	}
}

func auditSubjectDetail(r *http.Request) string {
	if subj, ok := SubjectFromRequest(r); ok && strings.TrimSpace(subj.ID) != "" {
		return "subj=" + subj.ID
	}
	return ""
}

type serveAuthInit struct {
	verifier *oidc.IDTokenVerifier
	authz    identity.Authorizer
}

func initServeAuth(parent context.Context, opts Options) (*serveAuthInit, error) {
	staticTok := strings.TrimSpace(opts.AuthToken)
	issuer := strings.TrimSpace(opts.OIDCIssuerURL)
	clientID := strings.TrimSpace(opts.OIDCClientID)
	hasOIDC := issuer != "" && clientID != ""

	var verifier *oidc.IDTokenVerifier
	if hasOIDC {
		ctx, cancel := context.WithTimeout(parent, 30*time.Second)
		defer cancel()
		hc := &http.Client{Timeout: 30 * time.Second}
		if opts.OIDCSkipTLSVerify {
			hc.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // dev-only flag; documented
			}
		}
		pctx := oidc.ClientContext(ctx, hc)
		provider, err := oidc.NewProvider(pctx, issuer)
		if err != nil {
			return nil, fmt.Errorf("serve: OIDC provider %q: %w", issuer, err)
		}
		verifier = provider.Verifier(&oidc.Config{ClientID: clientID})
	}

	var required []string
	for _, p := range strings.Split(opts.OIDCRequiredRoles, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			required = append(required, p)
		}
	}
	var az identity.Authorizer
	if opts.Authorizer != nil {
		az = opts.Authorizer
	} else {
		az = &identity.ExperimentalAuthorizer{
			StaticTokenConfigured: staticTok != "",
			RequiredOIDCRoles:     required,
		}
	}
	return &serveAuthInit{verifier: verifier, authz: az}, nil
}
