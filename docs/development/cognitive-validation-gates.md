# Cognitive validation gates

Use this checklist before claiming completion on architecture/UI/docs changes.

## Language gate (WUI-first)

- Product/operator docs lead with browser workflow and graph value.
- No operator-facing terminal-first narrative.

## Cognitive gate

Every changed area maps to at least one principle:

- Cognitive load reduction
- Progressive disclosure
- Gestalt/signal clarity
- SA support (perceive -> understand -> act)

## Structural gate (code)

- Targeted components/functions reduce branching or concern mixing.
- Pure derivation helpers exist where parse/map/render were previously fused.
- Side effects are explicitly documented at mutation boundaries.

## Error ergonomics gate

- New/edited backend errors include subsystem context.
- Error text is actionable and scoped to the failing surface.

## Documentation gate

- Wayfinding present on key pages.
- Action-oriented guide language remains intact.
- Mermaid dual-coding appears for complex flows.

## Verification gate

- `go vet ./...`
- `go build -o bin/omnigraph ./cmd/omnigraph`
- `cd packages/web && npm run lint`

## Acceptance criteria

- All six plan to-dos are complete.
- No regressions in lint/build checks.
- Resulting docs and code clearly improve operator decision speed and developer comprehension.
