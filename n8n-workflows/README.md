# n8n workflow exports

Import JSON files via **n8n → Workflows → Import from File**. Create credentials in n8n as referenced by each workflow (e.g. **HTTP Basic Auth** named `Zulip API` for Zulip ChatOps flows).

## ChatOps / webhooks

| Workflow | Webhook path (POST) | Notes |
|----------|---------------------|--------|
| **sFlow anomaly → Zulip** | `/webhook/sflow-anomaly` | Requires Zulip API credential; optional env `ZULIP_SFLOW_STREAM` (default `network`). |
| **Dependency-Track → Zulip** | `/webhook/dependency-track-alert` | Optional env `ZULIP_DEPENDENCY_TRACK_STREAM` (default `security`). |
| **Gitea docs sync → gateway** | `/webhook/gitea-docs-sync-relay` | Optional relay: forwards **raw body** + `X-Gitea-Signature` to `GATEWAY_DOC_SYNC_URL` (default `http://host.docker.internal/webhook/docs-sync`). **Prefer** Gitea → gateway directly; see [docs/DOCSIFY_GITEA.md](../docs/DOCSIFY_GITEA.md). |

### sFlow anomaly URLs (copy-paste)

| From | URL |
|------|-----|
| Docker (`n8n_net` / `sdn_lab_net`) | `http://n8n:5678/webhook/sflow-anomaly` |
| Host (published port) | `http://127.0.0.1:5678/webhook/sflow-anomaly` |
| Via Traefik | `http://<gateway>/n8n/webhook/sflow-anomaly` |

Point sFlow-RT REST hooks or your detector at one of the above. Details: [docs/SDN_TELEMETRY.md](../docs/SDN_TELEMETRY.md).
