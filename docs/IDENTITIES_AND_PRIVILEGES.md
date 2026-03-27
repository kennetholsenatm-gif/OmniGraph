# Identity and Privilege Definition (Greenfield VARLOCK)

## Goal

The **greenfielder** defines **who** has access and **at what level** in a single, VARLOCK-tagged file. No secrets (passwords, keys) live in this file—only usernames, privilege levels, and optional LDAP-style attributes. Secrets are generated or stored in Vault and linked by username.

## Ideas

### 1. **VARLOCK identity schema (recommended)**

A schema file (e.g. `devsecops.identities.schema`) in the same style as `devsecops.env.schema`:

- **@identity-spec** / **@role-spec** — Sections for “identity” (user) or “role” (group of users).
- **Privilege level** — Single field, e.g. `privilege_level: admin | operator | developer | viewer | auditor`, with **@validation** and **@description**.
- **LDAP-friendly attributes** — Optional `ou` (organizational unit), `cn`, `uid`, `memberOf` so the same file can be:
  - Consumed by Keycloak (users + realm roles),
  - Exported to LDIF for FreeIPA / 389-ds,
  - Used by Ansible or a bootstrap script to create users and assign roles.

Example (one user per block, no passwords):

```
@identity-spec: ADMIN
uid: kbolsen
@description: "Admin identity; password and SSH key in Vault at secret/devsecops"
privilege_level: admin
@validation: "admin | operator | developer | viewer | auditor"
ou: admins
cn: "Kenneth Olsen"
mail: kenneth.olsen.atm@gmail.com
```

### 2. **Privilege levels (canonical list)**

Define a small set of levels so every consumer (Keycloak, LDAP, Teleport, n8n) maps to the same semantics:

| Level     | Typical use | Keycloak role (example) | LDAP group (example) |
|----------|-------------|--------------------------|----------------------|
| **admin**   | Full pipeline and IAM | `master-realm` admin, `realm-admin` | `ou=admins` |
| **operator**| Deploy, scale, view secrets | `operator`, `view-secrets` | `ou=operators` |
| **developer** | Push code, run workflows, read logs | `developer`, `n8n-user` | `ou=developers` |
| **viewer**   | Read-only dashboards and logs | `view-only` | `ou=viewers` |
| **auditor**  | Read-only + audit logs | `auditor` | `ou=auditors` |
| **service_agent** | Non-interactive LDAP / agent binds | `service-account` (example) | `ou=service-accounts` |

Schema can enforce this with **@validation: "admin | operator | developer | viewer | auditor | service_agent"** and **@privilege-level** in the doc.

### 3. **LDAP-style layout with VARLOCK tags**

Keep attributes LDAP-like so the file can be transformed to LDIF or used to drive an LDAP provider:

- **dn** — Optional; e.g. `dn: uid=kbolsen,ou=admins,dc=devsecops,dc=local`. VARLOCK tag **@dn** or a single **dn:** line.
- **ou** — Organizational unit = logical group (admins, operators, developers).
- **memberOf** — List of group DNs or role names the user belongs to.
- **objectClass** — Optional; e.g. `inetOrgPerson`, `posixAccount` for LDIF export.

All of these stay as VARLOCK-tagged keys so the file is still one source of truth and tooling can generate LDIF, Keycloak JSON, or Terraform/Ansible from it.

### 4. **Single file: identities + role definitions**

- **Part 1: Role definitions** — Map privilege levels to Keycloak roles and/or LDAP groups.
- **Part 2: Identities** — List of users with `uid`, `privilege_level`, optional `ou`, `mail`, `cn`.

Bootstrap (or Ansible) reads this file, creates users in Keycloak/LDAP, assigns roles from the role definitions, and passes passwords/keys from Vault (by uid) so no secrets are in the identity file.

### 5. **Where secrets live**

- **Identity file** — Only identifiers and privilege: `uid`, `privilege_level`, `ou`, `mail`, `cn`, etc.
- **Vault** — Per-user secrets keyed by username (e.g. `secret/devsecops/users/kbolsen` or flat keys `KEYCLOAK_ADMIN`, `KEYCLOAK_ADMIN_PASSWORD` for the bootstrap admin). Bootstrap or `secrets-bootstrap.ps1` generates passwords and writes them to Vault; identity file is only “who and what level.”

---

## Recommended: `devsecops.identities.schema`

A single schema file that:

1. Defines **allowed privilege levels** (with @validation).
2. Lists **identities** (one block per user) with VARLOCK tags and LDAP-friendly attributes.
3. Optionally defines **role-to-level** mapping (e.g. Keycloak role name per level).
4. Is consumed by greenfield tooling to create Keycloak users, assign roles, and optionally export LDIF.

- **`devsecops.identities.schema`** — VARLOCK-tagged definition of privilege levels (with **@keycloak-role** and **@ldap-ou** per level) and example identities (no passwords, no keys).
- **`privilege_levels.json`** — Machine-readable mapping: for each privilege level, `keycloak_role` and `ldap_ou`. Consumed by sync and LDIF scripts so one mapping drives both Keycloak and LDAP.
- **`identities.example.yaml`** — Optional YAML list of identities; copy to `identities.yaml` (gitignore it) and point tooling at it. Same fields: `uid`, `privilege_level`, `ou`, `cn`, `mail`, plus optional `account_kind`, `member_of` (see [LDAP_ACCOUNTS_AND_GROUPS.md](LDAP_ACCOUNTS_AND_GROUPS.md)).
- **`ldap-directory.template.yaml`**, **`ldap-groups.template.ldif`** — Templated LDAP groups (human interactive vs infra vs server vs agent service accounts); narrative in [LDAP_ACCOUNTS_AND_GROUPS.md](LDAP_ACCOUNTS_AND_GROUPS.md).
- **`scripts/sync-identities-to-keycloak.ps1`** — Reads `identities.yaml` (or `.json`) and `privilege_levels.json`, creates Keycloak users in the master realm, ensures realm roles exist, assigns the role for each user’s `privilege_level`. Passwords from Vault (`secret/devsecops` or `secret/users/<uid>`) or generated and stored in Vault. Requires Keycloak admin credentials (env or Vault).
- **`scripts/export-identities-to-ldif.ps1`** — Reads the same identity list and privilege levels, writes **identities.ldif** with organizationalUnit and inetOrgPerson entries for import into LDAP/FreeIPA/389-ds. No passwords in LDIF; use `-BaseDn` (default `dc=devsecops,dc=local`) and `-OutputPath`.
