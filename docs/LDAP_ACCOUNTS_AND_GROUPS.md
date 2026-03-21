# LDAP accounts and groups template

This repo separates **who you are** (uid, human vs automation) from **what you may do** (Keycloak `privilege_level` → realm role, LDAP `memberOf` / POSIX `memberUid`).

Canonical machine-readable templates:

| Artifact | Purpose |
|----------|---------|
| [ldap-directory.template.yaml](../ldap-directory.template.yaml) | One-file view: `ldap_groups` + `identities` (copy `identities` to `identities.yaml` for [sync-identities-to-keycloak.ps1](../scripts/sync-identities-to-keycloak.ps1)) |
| [ldap-groups.template.ldif](../ldap-groups.template.ldif) | `ou=groups` and POSIX groups with `memberUid` |
| [identities.example.yaml](../identities.example.yaml) | Minimal copy-paste list aligned with the template |
| [privilege_levels.json](../privilege_levels.json) | Maps `privilege_level` → Keycloak realm role + default LDAP OU (override with per-user `ou:`) |
| [devsecops.identities.schema](../devsecops.identities.schema) | VARLOCK-style field definitions |

## Account kinds (`account_kind`)

| Kind | Intent | Typical use |
|------|--------|-------------|
| `human_interactive` | Human privileged **interactive** | SSO, **chatbots**, web apps, daily operator work |
| `human_privileged_infra` | Human **infrastructure** | Break-glass style, IaC, cluster/platform admin — not routine app/chat |
| `human_privileged_server` | Human **server / OS** | SSH, sudo on nodes, session access |
| `service_agent` | **Non-interactive** bind | AI agents, batch jobs, LDAP or app passwords from Vault — prefer **Keycloak client credentials** when the app supports OIDC |

`account_kind` is documentation and LDIF `description` for operators; **Keycloak** still uses `privilege_level` from [privilege_levels.json](../privilege_levels.json).

## Groups (POSIX template)

- **`human-interactive`** — `kbolsen`, `gale`, `katelyn`
- **`privileged-infra`** — `kbolsen-infra`, `gale-infra`
- **`privileged-server`** — `kbolsen-server`, `gale-server`
- **`agent-service-accounts`** — add `svc-*` (and `memberUid` in LDIF) as you create them

Adjust **gidNumber** values for your site; avoid collisions with local system groups.

## Keycloak sync vs LDAP

- [scripts/sync-identities-to-keycloak.ps1](../scripts/sync-identities-to-keycloak.ps1) creates **password-capable realm users** from `identities.yaml`. For **pure OIDC client** automation, you usually **do not** mirror those as human users — use the automation client ([IAM_LDAP_AND_AUTOMATION.md](IAM_LDAP_AND_AUTOMATION.md)).
- **Service agent** UIDs in the template are optional: keep them in FreeIPA/LDAP for binds; omit them from `identities.yaml` if you do not want Keycloak user entries.

## LDIF export

```powershell
.\scripts\export-identities-to-ldif.ps1 -IdentityFile .\identities.yaml -OutputPath .\identities.ldif `
  -AppendGroupsLdifPath .\ldap-groups.template.ldif
```

Or append manually: `Get-Content .\ldap-groups.template.ldif | Add-Content .\identities.ldif`

FreeIPA users often use `ipa user-add` / `ipa group-add-member` instead of raw LDIF; this template still defines the **intended** group membership for firewall and sudo matchers (`memberOf`, Unix groups).
