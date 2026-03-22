# Lifecycle and handoff

End-to-end deployment flow (target architecture):

1. **Trigger** — PR opened; control plane runs in CI or locally.
2. **Phase 1 — Validation** — Parse `.omnigraph.schema`; fail on type errors.
3. **Phase 2 — Plan** — `tofu plan` (or Terraform); parse plan JSON for projected resources; optional `ansible-playbook --check` against projected inventory.
4. **Phase 3 — Visualization** — Merge results into **omnigraph/graph/v1** JSON for the UI or PR artifact.
5. **Phase 4 — Apply and handoff** — After approval, `tofu apply`; **intercept** `.tfstate`; map outputs (e.g. instance IPs) into Ansible; run playbooks against live hosts.
6. **Phase 5 — Sync** — Webhooks to NetBox (and similar) for CMDB alignment.

Today’s implementation covers **validation**, **coercion**, **state/plan parsing**, **inventory text**, **graph emission**, **exec runner**, and a **NetBox webhook client**. Orchestrating full Phase 2–4 in one command and posting graph JSON back to GitHub is the next integration step.

## Related documentation

- [Architecture overview](https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/docs/architecture.md)
- [Integrations](https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/docs/integrations.md)
- [Execution matrix](https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/docs/execution-matrix.md)
