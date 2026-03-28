# End-to-end (E2E) testing

OmniGraph’s **unit and integration tests** prove that individual packages behave correctly. They are necessary but not sufficient: **real deployments** fail in ways that isolated tests rarely reproduce—Ansible exits non-zero, HTTP APIs hang, inventory looks right on paper but a playbook still aborts. The **`e2e/`** directory hosts a **dedicated harness** that exercises the **full pipeline** against **simulated infrastructure** so we see the system **under duress**, not only on the happy path.

## What lives in `e2e/`

- **Simulated Ansible-facing endpoints** — HTTP mocks or containerized stubs that speak the shapes our orchestration and inventory paths expect, without requiring a real data center.
- **Mock IR, graph, and state fixtures** — versioned JSON and workspace layouts checked into `e2e/fixtures/` (or similar), so scenarios are **reproducible** in CI.
- **Drivers** — Go tests that `go build` the CLI (`e2e/cli_test.go`), run `omnigraph ir` against fixtures, and assert exit codes and stdout.
- **HTTP simulation** — `e2e/ansible_simulation_test.go` uses `httptest` to model failure responses from Ansible-adjacent endpoints (extend with real orchestration wiring over time).

The exact layout follows the harness implementation; the **contract** for contributors is: **if it spans CLI + Emitter Engine + external tool boundaries and you care about failure modes, it belongs in E2E**.

## Static integration tests versus E2E

| Use | When |
|-----|------|
| **`go test ./pkg/...` and `./internal/...`** | Pure logic, parsers, emitters, small fixtures—fast feedback. |
| **Web unit tests** (`packages/web`) | React components, hooks, TypeScript utilities. |
| **`e2e/`** | **Full pipeline**: binary on disk, optional Wasm build, mocked Ansible/API behavior, **injected failures** (timeouts, 500s, malformed bodies, non-zero playbook exits). |

If a regression could only appear when **two processes and the filesystem** disagree, lean toward E2E.

## Failure injection philosophy

E2E scenarios **must** include **unhappy paths**: slow responses, truncated JSON, exit code 2 from a fake playbook runner, missing inventory keys. The goal is not sadism—it is **confidence** that operators see **actionable errors** and that the **browser** and **CLI** remain usable when one leg of the pipeline misbehaves.

## Running locally and in CI

From the repository root (after `go work sync` and building the CLI):

```bash
go test ./e2e/...
```

Some suites may require tags or environment variables (for example skipping container-based cases); see comments in `e2e/` once the harness is present. **CI** runs the same target so merges cannot drift from documented behavior. For automation context, see [CLI and CI](../cli-and-ci.md#end-to-end-e2e-suite).

## Related reading

- [Platform architecture](platform-architecture.md) — why E2E exists in the overall design
- [Execution matrix](../core-concepts/execution-matrix.md) — runners and real orchestration
- [CONTRIBUTING.md](../../CONTRIBUTING.md) — full verification checklist before a PR
