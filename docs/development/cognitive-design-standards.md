# Cognitive design standards (WUI-first)

These standards guide architecture, frontend, and documentation decisions in OmniGraph.

## 1) Working-memory budget (code)

- Keep complex routines close to **4-7 active concepts**.
- Prefer pure transformation helpers over mixed parse+map+render logic.
- Split orchestration from derivation. Name helpers by outcome (e.g., `buildGraphCanvasViewModel`).

## 2) Progressive disclosure contract (UI/docs)

- Present information in this order: **overview -> focused drilldown -> raw schema/details**.
- Keep global chrome low during node-focused triage.
- In docs, place mental model first; contracts and wire format second.

## 3) Visual triage grammar (Gestalt + signal detection)

- Use **proximity** to group related infrastructure.
- Use **similarity** for resource-kind consistency.
- Reserve high-contrast rings/alerts for incident-relevant signals.
- Avoid using animation as default status; motion indicates state transition or urgency only.

## 4) State language contract (NOC/SOC alignment)

Prefer terms that align with incident and security operations:

- **drift**, **reconciliation**, **freshness**, **staleness**, **blast radius**, **handoff**.
- For each status/error, include: **where**, **what likely happened**, **next action**.

## 5) Error ergonomics and side effects (backend)

- Prefix errors by subsystem (`serve:*`, `integration host:*`, `state:*`).
- Keep errors local and actionable (path/field/context).
- Annotate mutation boundaries (e.g., replacing hub state, patch revision increments).

## 6) Documentation wayfinding

Each major page should answer quickly:

1. Where am I in the documentation hierarchy?
2. Which decision is this page helping me make?
3. What should I read next?

## 7) Dual-coding requirement

For non-trivial flows (handoff, sync, integration execution), include a compact mermaid diagram alongside prose.

## 8) Scope discipline

- Favor targeted changes over broad refactors.
- Keep product/operator narrative WUI-first.
- Keep contributor shell procedures in `docs/development/contributor-commands.md` only.
