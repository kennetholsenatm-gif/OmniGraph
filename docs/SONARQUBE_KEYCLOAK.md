# SonarQube and Keycloak (SSO)

SonarQube in this stack is served at **`/sonarqube`** behind Traefik ([routes.yml](../single-pane-of-glass/traefik/dynamic/routes.yml)) with **`SONAR_WEB_CONTEXT=/sonarqube`**.

## Database

- JDBC uses the **messaging** Postgres on `msg_backbone_net` (`postgres:5432`).
- **`sonar-db-init`** (in [docker-compose.messaging.yml](../docker-compose/docker-compose.messaging.yml)) creates the `sonar` role and database using **`SONAR_JDBC_PASSWORD`** from Vault / `secrets-bootstrap.ps1`.
- Start **messaging before** (or merged with) tooling so the DB exists before SonarQube connects.

## Community Edition vs OIDC

- **Community Edition** (default image `sonarqube:*-community`) does **not** ship first-class OpenID Connect to Keycloak like Grafana. Supported options include:
  - **Developer Edition+**: configure **SAML** or vendor **OAuth/OIDC** per SonarQube documentation; use **`SONARQUBE_OIDC_CLIENT_SECRET`** (or SAML keys) from Vault with a Keycloak client whose redirect URIs include your gateway, e.g. `https://<gateway>/sonarqube/oauth2/callback` (exact path per SonarQube version).
  - **Community**: **HTTP reverse proxy authentication** (request headers) or **Traefik forward-auth** to Keycloak, aligned with SonarQube’s reverse-proxy auth docs for your LTS version (properties change between releases—verify before production).
  - **OAuth providers** SonarQube CE supports natively (GitHub, GitLab, Microsoft): optionally front with Keycloak identity brokering.

## Keycloak client checklist (Developer+ / SAML-OIDC)

1. Create a **confidential** (or appropriate) client in Keycloak.
2. **Valid redirect URIs**: include `http(s)://<GATEWAY_HOST>/sonarqube/...` per SonarQube’s callback path.
3. Store the client secret in Vault as **`SONARQUBE_OIDC_CLIENT_SECRET`** (exported by `secrets-bootstrap.ps1` into the process environment).
4. Map Keycloak groups to SonarQube permissions as required.

## Secrets (zero-disk)

- **`SONAR_JDBC_PASSWORD`**: generated in [scripts/secrets-bootstrap.ps1](../scripts/secrets-bootstrap.ps1) and written to Vault.
- **`SONARQUBE_OIDC_CLIENT_SECRET`**: generated for Keycloak client provisioning; inject into SonarQube configuration when using licensed/OIDC flows.
