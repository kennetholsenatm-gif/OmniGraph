# Repo ownership split

## `C:\GiTeaRepos\devsecops-pipeline`

Primary ownership:

- OpenNebula runtime orchestration and operations
- Ansible control-plane playbooks/roles
- Local Incus control-plane and Gitea recovery runbooks
- Steady-state stack operation documentation
- **Site topology intent** (edge firewall mini PC, IAM mini PC, phased roadmap): [CANONICAL_DEPLOYMENT_VISION.md](CANONICAL_DEPLOYMENT_VISION.md), [ROADMAP.md](ROADMAP.md) — *orchestration of workloads on hosts that may live under OpenNebula; not turnkey OpenNebula installer*

## `C:\GiTeaRepos\Deploy`

Primary ownership:

- Greenfield one-shot bootstrap
- Initial registration and secret/bootstrap entrypoints
- Day-0/day-1 deployment wrappers

## Operator rule

- If it is greenfield bootstrap, run from `Deploy`.
- If it is runtime operations/control-plane, run from `devsecops-pipeline`.
