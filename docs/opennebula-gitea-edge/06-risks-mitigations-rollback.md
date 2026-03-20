# Risks, Mitigations, and Rollback

## Risk register

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Ceph IOPS saturation on UCS-E SSD | Slow Git ops, UI timeouts | Medium | Dedicated storage VLAN; SSD OSDs; limit recovery traffic; size pool; consider `size=2` cautiously |
| Small Ceph cluster data loss (`size=2`) | Repo corruption | Low–Med | Document RPO; add OSD host; frequent `gitea dump` off cluster |
| STP loop / misconfigured trunk | Outage | Low | BPDU guard; explicit allowed VLANs; lab validation; change window |
| K3s single control plane | Platform rebuild needed if CP lost | Med | Backups of etcd; second Mini-PC DR scripts; add 3rd server node when possible |
| Gitea RWO PVC + multi-replica | Second pod cannot mount volume | High if ignored | Single replica or move to RWX (CephFS/NFS) |
| Version skew (Windows Gitea vs Linux) | Failed migration | Med | Match versions dry-run; upgrade source first if required |
| Missed `SECRET_KEY` / JWT | Forced re-login / broken tokens | Med | Copy secrets from old `app.ini`; test OAuth |
| LFS not copied | Missing objects | Med | Verify LFS path in dump; checksum sample objects |
| Webhook IP allowlists | Broken CI | Med | Document egress IP changes; update receivers |

## Bottlenecks

- **Edge uplink bandwidth:** Large monorepos prolong transfer; run final dump over LAN; compress selectively.
- **Disk latency:** Keep PostgreSQL on SSD-backed pool; tune `shared_buffers` and connections.
- **Antivirus on Windows:** Can slow dump; exclude Gitea data path temporarily per policy.

## Rollback procedure (post-DNS cutover)

**When:** Critical auth loss, data corruption detected, or prolonged outage beyond SLA.

1. **Freeze target:** scale `gitea-prod` deployment to 0; block ingress at ISR ACL if needed.
2. **Revert DNS** to previous record (TTL permitting) or publish maintenance page pointing to status.
3. **Start Windows source** Gitea from **pre-cutover** snapshot/dump (never mix old binary with new partial data).
4. **Verify** internal admin login + sample repo before announcing.
5. **Post-incident:** clone failing data from target PVC offline for forensics; do not delete until root-caused.

## Rollback procedure (pre-DNS — safest)

If validation fails before DNS switch:

1. Keep production Windows host frozen until root cause fixed **or** abort freeze.
2. Delete staging/preview data on K3s; fix restore procedure; reschedule window.
3. If unfrozen early, run incremental activity on Windows until next attempt.

## Contingency: partial migration

If only Git data moved but DB restore failed:

- Do **not** expose service; restore DB from dump again or revert entirely.
- Never let users write to two masters.

## Security notes

- Treat dumps as **credential-bearing**; encrypt at rest; limit ACLs on transfer shares.
- Rotate inbound deploy keys if compromise suspected during transfer.
