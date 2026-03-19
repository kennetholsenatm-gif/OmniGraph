# Wazuh SIEM config directory

TLS material and OpenSearch security files are **not** committed. Generate them once:

```powershell
# From repo root (Linux host recommended; Docker required)
.\scripts\bootstrap-wazuh-siem-config.ps1
```

This clones [wazuh-docker](https://github.com/wazuh/wazuh-docker) `v4.9.2`, runs `generate-indexer-certs.yml`, copies `single-node/config/*` into `wazuh-config/`, and appends `server.basePath: "/wazuh"` for Traefik.

Then start the optional stack (process env from Vault / `secrets-bootstrap.ps1`):

```powershell
$env:DEVSECOPS_INCLUDE_SIEM = "1"
cd docker-compose
.\launch-stack.ps1
# Or: docker compose -f docker-compose.siem.yml up -d
```

See [docs/WAZUH_SIEM.md](../../docs/WAZUH_SIEM.md) for passwords, `vm.max_map_count`, Keycloak OIDC, and agent registration.
