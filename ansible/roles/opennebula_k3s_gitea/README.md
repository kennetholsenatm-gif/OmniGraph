# opennebula_k3s_gitea

Runs `kubernetes.core.helm` to install/upgrade the **Gitea** chart on an existing **K3s** (or any kubeconfig-backed) cluster. Aligns with [deployments/opennebula-gitea](../../../deployments/opennebula-gitea/README.md).

**Prereqs**

- `kubernetes.core` collection (`ansible-galaxy collection install -r collections/requirements.yml`).
- Valid kubeconfig on the target host (`opennebula_kubeconfig_path`, default `/etc/rancher/k3s/k3s.yaml`).
- Values file: copy `helm/gitea-values.example.yaml` to a writable path and set secrets.

**Enable from play**

Set `opennebula_helm_gitea_enabled: true` on hosts in `opennebula_k3s_helm` (see `playbooks/opennebula-hybrid-site.yml`).
