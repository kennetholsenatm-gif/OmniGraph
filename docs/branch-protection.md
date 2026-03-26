# Branch protection for `main`

GitHub does not store branch protection rules in this repository. Configure them in the repo **Settings** on GitHub, or apply the ruleset in [`scripts/github-ruleset-main.json`](../scripts/github-ruleset-main.json) using the scripts below (requires admin rights and a token with `repo` scope).

## What to enforce

These settings match OmniGraph’s CI layout: every PR to `main` runs **CI**, **CodeQL Advanced**, and **Security** on all paths. Workflows that only run when certain paths change (**Policy Check**, **Workflow Style**) are **not** listed as required checks, because GitHub would otherwise leave PRs waiting for checks that never start.

| Goal | Setting |
|------|---------|
| No direct pushes of unreviewed code | Require a pull request before merging |
| Review | At least **1** approving review |
| Threads | Require conversation resolution before merging |
| Up-to-date | Require branches to be up to date before merging (strict status checks) |
| CI gates | Required status checks (see table below) |
| History | Block force-pushes and branch deletion |

### Required status check names

Use **exact** names as they appear on a pull request’s “Checks” tab (usually `Workflow name / job name`):

| Check name | Workflow | Notes |
|------------|----------|--------|
| `CI / go` | CI | Go build, lint, tests, Wasm |
| `CI / web` | CI | Vite lint + build |
| `CodeQL Advanced / Analyze (go)` | CodeQL Advanced | Go analysis |
| `CodeQL Advanced / Analyze (actions)` | CodeQL Advanced | Actions workflow analysis |
| `Security / trivy-fs` | Security | Trivy filesystem scan |

If GitHub shows slightly different labels (for example punctuation), pick the names from the dropdown when editing the rule, or copy from a recent PR.

### Optional checks (path-filtered)

- **Policy Check** and **Workflow Style** only run when relevant files change. Do **not** add them as required status checks unless you change those workflows to run on every PR (for example by widening `paths` or removing path filters).

### Dependabot and automation

If **required reviews** block Dependabot from merging its own PRs, either:

- Grant Dependabot merge via **Ruleset bypass actors** (GitHub Settings → Rules → Rulesets → edit ruleset → Bypass list), or  
- Merge dependency PRs manually after checks pass.

## Option A: GitHub UI (rulesets, recommended)

1. Open **Settings → Rules → Rulesets** for [OmniGraph](https://github.com/kennetholsenatm-gif/OmniGraph/settings/rules).
2. **New ruleset** → **New branch ruleset**.
3. **Name:** e.g. `Protect main`.
4. **Enforcement status:** Active.
5. **Target branches:** Add pattern `main` (or include `refs/heads/main`).
6. Add rules:
   - Restrict deletions  
   - Require linear history (optional)  
   - Require a pull request before merging (set review count, dismiss stale reviews, require conversation resolution)  
   - Require status checks to pass (add the five checks above; enable **Require branches to be up to date before merging**)
7. Save.

Classic **Branch protection** (older UI under Settings → Branches) can enforce the same ideas; rulesets are easier to audit and extend.

## Option B: Apply ruleset via API (`gh` CLI)

Prerequisites: [GitHub CLI](https://cli.github.com/) (`gh`), logged in as a user with **admin** on the repository.

From the repository root:

**Linux / macOS / Git Bash:**

```bash
./scripts/apply-main-ruleset.sh
```

**Windows PowerShell:**

```powershell
./scripts/apply-main-ruleset.ps1
```

The scripts POST [`scripts/github-ruleset-main.json`](../scripts/github-ruleset-main.json) to `POST /repos/{owner}/{repo}/rulesets`.

If a ruleset with the same name already exists, delete or rename it in the UI first, or update it with `PATCH /repos/{owner}/{repo}/rulesets/{ruleset_id}` (use `GET /repos/{owner}/{repo}/rulesets` to list IDs).

## Verify

- Open a test PR against `main` and confirm all five required checks appear and gate the merge button.
- Confirm you cannot push directly to `main` without bypass (if direct pushes are disabled by the ruleset).

## References

- [About rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets)
- [REST API: Create a repository ruleset](https://docs.github.com/en/rest/repos/rules#create-a-repository-ruleset)
