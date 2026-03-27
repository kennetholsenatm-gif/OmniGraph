#!/usr/bin/env bash
# Run on the target mini PC (e.g. Mini-PC-IAM) after mounting the USB bundle.
# Does not push secrets; loads optional docker images, creates Docker bridges; optional Ansible LXC IAM bootstrap.
#
# Usage:
#   sudo BUNDLE_ROOT=/mnt/usb ./bootstrap-on-target.sh
#
# Env:
#   BUNDLE_ROOT      - Required. Directory containing repo/, collections/, optional images/
#   REPO_SUBDIR      - Default: repo
#   SKIP_DOCKER_LOAD - Set to 1 to skip docker load images/*.tar
#   SKIP_CREATE_NETWORKS - Set to 1 to skip scripts/create-networks.sh
#   RUN_ANSIBLE_BOOTSTRAP - Set to 1 to run playbooks/bootstrap-mini-pc-iam.yml (needs ansible-playbook, inventory)
#   BOOTSTRAP_INVENTORY - Required if RUN_ANSIBLE_BOOTSTRAP=1; path to inventory file (e.g. /root/mini-pc-iam.yml)
#   ANSIBLE_PLAYBOOK_EXTRA_ARGS - Extra args for ansible-playbook (e.g. -K or --vault-password-file /path)
set -euo pipefail

BUNDLE_ROOT="${BUNDLE_ROOT:-}"
REPO_SUBDIR="${REPO_SUBDIR:-repo}"
SKIP_DOCKER_LOAD="${SKIP_DOCKER_LOAD:-0}"
SKIP_CREATE_NETWORKS="${SKIP_CREATE_NETWORKS:-0}"
RUN_ANSIBLE_BOOTSTRAP="${RUN_ANSIBLE_BOOTSTRAP:-0}"
BOOTSTRAP_INVENTORY="${BOOTSTRAP_INVENTORY:-}"
ANSIBLE_PLAYBOOK_EXTRA_ARGS="${ANSIBLE_PLAYBOOK_EXTRA_ARGS:-}"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root (sudo) for docker and network create." >&2
  exit 1
fi

if [[ -z "${BUNDLE_ROOT}" ]] || [[ ! -d "${BUNDLE_ROOT}/${REPO_SUBDIR}" ]]; then
  echo "Set BUNDLE_ROOT to the bundle directory that contains ${REPO_SUBDIR}/ (devsecops-pipeline copy)." >&2
  exit 1
fi

REPO_ABS="${BUNDLE_ROOT}/${REPO_SUBDIR}"
COLLECTIONS_ABS="${BUNDLE_ROOT}/collections"
IMAGES_DIR="${BUNDLE_ROOT}/images"

export ANSIBLE_COLLECTIONS_PATH="${COLLECTIONS_ABS}"
export ANSIBLE_NOCOWS=1

echo "==> Bundle: ${BUNDLE_ROOT}"
echo "==> Repo:  ${REPO_ABS}"
echo "==> ANSIBLE_COLLECTIONS_PATH=${ANSIBLE_COLLECTIONS_PATH}"

if ! command -v docker >/dev/null 2>&1; then
  echo "WARNING: docker not in PATH; skip image load and create-networks may fail." >&2
else
  if [[ "${SKIP_DOCKER_LOAD}" != "1" ]] && [[ -d "${IMAGES_DIR}" ]]; then
    shopt -s nullglob
    for tar in "${IMAGES_DIR}"/*.tar "${IMAGES_DIR}"/*.tar.gz; do
      echo "==> docker load < ${tar}"
      docker load -i "${tar}"
    done
    shopt -u nullglob
  fi

  if [[ "${SKIP_CREATE_NETWORKS}" != "1" ]] && [[ -x "${REPO_ABS}/scripts/create-networks.sh" ]]; then
    echo "==> Creating Docker bridge networks (100.64.*)"
    bash "${REPO_ABS}/scripts/create-networks.sh"
  fi
fi

if [[ "${RUN_ANSIBLE_BOOTSTRAP}" == "1" ]]; then
  if [[ -z "${BOOTSTRAP_INVENTORY}" ]] || [[ ! -f "${BOOTSTRAP_INVENTORY}" ]]; then
    echo "ERROR: RUN_ANSIBLE_BOOTSTRAP=1 requires BOOTSTRAP_INVENTORY pointing to an existing file." >&2
    exit 1
  fi
  if ! command -v ansible-playbook >/dev/null 2>&1; then
    echo "ERROR: ansible-playbook not in PATH. Install ansible-core on the target or use venv." >&2
    exit 1
  fi
  echo "==> ansible-playbook playbooks/bootstrap-mini-pc-iam.yml (LXC devsecops-iam only)"
  cd "${REPO_ABS}/ansible"
  # shellcheck disable=SC2086
  ansible-playbook \
    -i "${BOOTSTRAP_INVENTORY}" \
    playbooks/bootstrap-mini-pc-iam.yml \
    ${ANSIBLE_PLAYBOOK_EXTRA_ARGS}
  echo "==> Ansible bootstrap finished."
fi

echo ""
echo "==> Next steps (collections under BUNDLE_ROOT/collections if build-bundle.sh was used):"
echo "    export ANSIBLE_COLLECTIONS_PATH=\"${COLLECTIONS_ABS}\""
echo "    cd ${REPO_ABS}/ansible"
echo ""
echo "    # One-shot LXC IAM (same as RUN_ANSIBLE_BOOTSTRAP=1):"
echo "    # ansible-playbook -i /path/to/inventory.yml playbooks/bootstrap-mini-pc-iam.yml -K"
echo ""
echo "    # Or full mesh / bare-metal IAM — see printed commands in prior bundle revisions or:"
echo "    # docs/BOOTSTRAP_USB_BUNDLE.md"
echo ""
echo "See docs/BOOTSTRAP_USB_BUNDLE.md and docs/CANONICAL_DEPLOYMENT_VISION.md."
