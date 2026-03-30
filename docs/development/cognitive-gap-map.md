# Cognitive gap map (baseline)

This baseline audits current OmniGraph code and docs against expanded cognitive-science criteria before further refactors.

## Scope reviewed

- `README.md`
- `docs/README.md`
- `docs/core-concepts/ux-architecture.md`
- `docs/guides/workflows-noc-sre.md`
- `docs/guides/workflows-soc-secops.md`
- `packages/web/src/graph/GraphCanvas.tsx`
- `packages/web/src/mvp/OmniGraphMVP.tsx`
- `internal/serve/*` + `internal/runner/integration_host.go` + `internal/omnistate/types.go`

## Gap matrix

| Area | Current strengths | Gaps to close | Priority |
|---|---|---|---|
| WUI-first product narrative | Root docs are graph-first and operator-oriented | Some contributor terminology can still pull attention toward backend internals too early | Medium |
| Cognitive load in frontend | Clear tab separation by operational mode; typed state | Large component surfaces (`OmniGraphMVP`, `GraphCanvas`) still require tracking many concerns | High |
| Gestalt / preattentive scanning | Enclave clusters, kind-specific nodes, drill highlight rings | Visual grammar is implied in code but not formalized as a reusable standard | High |
| Situation awareness / triage | Triage mode and node-scoped context are present | Confidence/freshness cues and “what next” hints are not consistently explicit across panels | Medium |
| Backend error ergonomics | Many subsystems already prefix errors | A few paths still return unscoped errors or mixed wording | Medium |
| Side-effect clarity | Several comments describe behavior | Side-effect boundaries are not uniformly documented on mutating paths | Medium |
| Documentation wayfinding | `docs/README.md` improved; workflow guides are action-oriented | “You are here / next decision” framing is still uneven across deep-dive pages | Medium |
| Dual-coding (text+diagram) | Core architecture and handoff pages include mermaid flows | Need stronger decision-flow diagrams in workflow and validation docs | Low |

## Applied cognitive lenses

- **Cognitive Load Theory:** reduce variables and branch points per component/function.
- **Gestalt + preattentive cues:** encode relatedness and risk at a glance.
- **Situation Awareness (Endsley):** support perception, comprehension, and projection in incident flow.
- **Signal detection / decision hygiene:** reduce false-positive visual urgency; reserve high-contrast cues for real risk.
- **Distributed cognition:** make handoff artifacts legible across NOC/SOC/automation roles.

## Immediate implementation targets

1. Extract pure view-model derivation from graph rendering path (`GraphCanvas`).
2. Codify cognitive/UI/error standards for contributors in one doc.
3. Tighten subsystem-local error and side-effect language in `serve`/`runner`/`omnistate`.
4. Add explicit validation gates with measurable acceptance criteria.
