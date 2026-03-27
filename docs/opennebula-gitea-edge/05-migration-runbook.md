# Migration Runbook: Windows Gitea → Linux (K3s on OpenNebula)

**Cutover model:** offline maintenance window (writes frozen during migration).

**Whole pipeline (not just the Gitea binary):** **IAM, messaging, tooling, ChatOps, and the Traefik gateway are Docker containers** — migrate their **volumes** and **re-run compose** (or Ansible) on OpenNebula Linux VMs. See **[CONTAINER-LIFT-TO-OPENNEBULA.md](CONTAINER-LIFT-TO-OPENNEBULA.md)**. Re-point integrations: [WHOLE-REPO-MIGRATION-SCOPE.md](WHOLE-REPO-MIGRATION-SCOPE.md).

## Participants and prerequisites

- Platform: OpenNebula + KVM + Ceph CSI + K3s operational; staging Gitea release tested.
- Access: admin on Windows source host; `kubectl`/Helm on target; DB credentials prepared.
- Downtime comms: announce window; freeze CI and webhooks if needed.

## Phase 0 — Inventory (T-7 days or earlier)

1. Record **Gitea version** on Windows (`gitea.exe --version` or UI footer).
2. Identify deployment mode:
   - Windows service binary vs Docker Desktop vs manual.
3. Locate data paths:
   - `CUSTOM_PATH`, repository root, `data/`, `custom/`, log path.
   - Typical Windows install: `C:\Gitea` or repo-adjacent `data`.
4. Identify **database** (PostgreSQL, MySQL, SQLite):
   - If SQLite, note path to `gitea.db`.
5. List **integrations:** OAuth/LDAP, SMTP, Actions runners, webhooks, packages, LFS.
6. Export **SSH host keys** directory if Git over SSH is used (reduces client warnings on cutover).

## Phase 1 — Staging dry run (T-3 days)

1. Clone target Helm values; deploy **staging** namespace with **empty** DB and PVCs.
2. On Windows, take a **copy** of production (not the live freeze yet):
   ```powershell
   # Service install — stop non-prod copy only if using a clone VM
   gitea.exe dump --type tar.gz --file C:\temp\gitea-dryrun.tar.gz
   ```
3. Transfer archive to Linux jump host:
   ```bash
   scp admin@windows:/c/temp/gitea-dryrun.tar.gz ./import/
   ```
4. Restore into **staging** per Option A or B below; run login, clone, LFS, webhook tests.
5. Reconcile gaps (LDAP mapping, `ROOT_URL`, SMTP).

### Restore options

**Option A — Native Gitea restore on a restore VM (simplest troubleshooting)**

1. Extract dump on a Linux VM with matching Gitea binary version.
2. Run `gitea restore` per upstream docs for the dump format.
3. Sync restored `repos/` and attachments into PVC before helm handoff **or** repackage for Helm init.

**Option B — Helm release + manual data seed**

1. `helm install` PostgreSQL + Gitea with valid `ROOT_URL` but **do not** expose to users.
2. Scale Gitea deployment to 0.
3. `kubectl exec` into postgres pod; drop empty DB if required; restore SQL dump from Windows export.
4. `rsync`/`tar` repository tree onto Gitea PVC mount (path per container layout, often `/data`).
5. Fix permissions (`git` user uid/gid in container).
6. Scale Gitea up; run `gitea migrate` if container entrypoint does not auto-run.

## Phase 2 — Production freeze (T-0)

1. Set banner / maintenance page (optional reverse-proxy).
2. **Stop** Gitea on Windows:
   - Service: `Stop-Service gitea`
   - Docker: `docker stop gitea`
3. Verify no open handles on repo directory (`handle.exe` or restart host if needed).
4. Final backup:
   ```powershell
   gitea.exe dump --type tar.gz --file C:\temp\gitea-CUTOVER.tar.gz
   ```
5. Optional integrity: hash archive
   ```powershell
   Get-FileHash C:\temp\gitea-CUTOVER.tar.gz -Algorithm SHA256 | Format-List
   ```

## Phase 3 — Transfer

Choose one:

```bash
# From Windows OpenSSH server
scp Administrator@windows:C:/temp/gitea-CUTOVER.tar.gz ./gitea-CUTOVER.tar.gz

# Or rsync if available
rsync -avP admin@windows:/c/temp/gitea-CUTOVER.tar.gz ./import/
```

Encrypt if crossing untrusted networks:

```bash
scp ./gitea-CUTOVER.tar.gz admin@bastion:/secure/import/
```

## Phase 4 — Restore production on K3s

1. Scale prod Gitea to 0 (if placeholder exists).
2. Restore DB and `/data` tree to prod PVCs (same as dry run).
3. Merge critical `app.ini` keys:
   - `SECRET_KEY`, `INTERNAL_TOKEN`, `JWT_SECRET` (must match old instance for sessions/tokens where applicable).
   - Database connection now points to K8s service DNS.
4. If preserving SSH host keys, mount `ssh` key directory as secret or subPath.
5. `helm upgrade` with final values; scale to 1 (or N with RWX storage only).
6. Watch logs:
   ```bash
   kubectl -n gitea-prod logs deploy/gitea -f --tail=200
   ```

## Phase 5 — User acceptance validation (still internal)

- [ ] Web UI login (local + SSO).
- [ ] HTTP clone/fetch push for sample repos.
- [ ] SSH clone if enabled (check host key stability).
- [ ] LFS push/pull sample.
- [ ] Webhooks fire (inspect receiver logs).
- [ ] Permissions/groups for orgs mirrored.
- [ ] Actions/workflows if used.

## Phase 5b — Repo stack: Docsify, gateway, and webhooks

Align with [docs/DOCSIFY_GITEA.md](../DOCSIFY_GITEA.md) and [docs/NETWORK_DESIGN.md](../NETWORK_DESIGN.md):

- [ ] **`DOCS_GIT_REPO`** (and any clone URL using `gitea` hostname) resolves to the **new** Gitea endpoint on **`100.64.1.0/24`** (`devsecops-gitea`) or the corporate DNS name you expose from that segment.
- [ ] **Single-pane / Traefik** on **`100.64.5.0/24`** (`devsecops-gateway`) still reaches Gitea on **3000** / Git HTTP as designed; update static routes or VR if IPs moved.
- [ ] **Docs-sync webhook** URL points to a reachable **`/webhook/docs-sync`** (gateway). **`GITEA_DOCS_WEBHOOK_SECRET`** still matches Gitea’s webhook **Secret**; **`X-Gitea-Signature`** verification must not break (avoid n8n relays that alter the body unless following [n8n-workflows/README.md](../../n8n-workflows/README.md)).
- [ ] Other Gitea webhooks (CI, chat, inventory) updated for **new Base URL** / IP allowlists on receivers.
- [ ] No **hard-coded Windows paths** (`C:\GiTeaRepos`, `D:\...`) in pipelines, runner configs, or repo hooks.

## Phase 6 — DNS / routing cutover

1. Update **internal/external** DNS `A`/`AAAA` for `git.<domain>` to the **ingress or MetalLB VIP** (typically on **`100.64.5.0/24`** per `devsecops-gateway`, or your published NAT target).
2. Update ISR **DNAT** / static routes per [deployments/opennebula-kvm/VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md) (VR next-hop, **`100.64.0.0/10`** reachability).
3. Re-enable ingress TLS if cert was issued against staging name only — re-issue if SAN mismatch.
4. Announce end of maintenance.

## Phase 7 — Post-cutover (24–48h)

1. Monitor error rates, latency, webhook retries (including **docs-sync** and **POST /webhook/docs-sync**).
2. Trigger a **test push** to the docs repo and confirm **Docsify** updates at **`/docs`** (or host-based route).
3. Keep Windows host **powered off or service disabled** but **retain disk** until retention policy expires.
4. Document actual timings and issues for retrospective.

## Rollback trigger

Execute [06-risks-mitigations-rollback.md](06-risks-mitigations-rollback.md) rollback section if blocking defects appear after DNS switch.

## Appendix — Common Windows paths

| Component | Typical path |
|-----------|----------------|
| Binary service | `C:\Gitea\gitea.exe` |
| Custom `app.ini` | `C:\Gitea\custom\conf\app.ini` |
| Data root | `C:\Gitea\data` |
| Repos (if separate) | `C:\GiTeaRepos` or `data\repositories` |

Align with `repository.ROOT` in `app.ini` before copying data.
