# Identity and LDAP templates

Human vs automation accounts and POSIX-style groups are **templated in git** (no passwords in repo).

## Where to look

| Artifact | Path |
|----------|------|
| LDAP + group template (YAML) | `ldap-directory.template.yaml` |
| Group LDIF | `ldap-groups.template.ldif` |
| Keycloak / LDIF user list example | `identities.example.yaml` → copy to gitignored `identities.yaml` |
| Privilege → role / OU map | `privilege_levels.json` |
| Narrative | `docs/LDAP_ACCOUNTS_AND_GROUPS.md` |
| Keycloak automation (no static admin password) | `docs/IAM_LDAP_AND_AUTOMATION.md` |

## Account kinds (summary)

| Kind | Use |
|------|-----|
| `human_interactive` | SSO, chatbots, applications |
| `human_privileged_infra` | IaC / break-glass style |
| `human_privileged_server` | SSH / sudo on nodes |
| `service_agent` | Non-interactive binds; prefer OIDC client credentials when possible |

## Scripts

- Sync users to Keycloak: `scripts/sync-identities-to-keycloak.ps1`
- Export LDIF: `scripts/export-identities-to-ldif.ps1` (optional `-AppendGroupsLdifPath ldap-groups.template.ldif`)
