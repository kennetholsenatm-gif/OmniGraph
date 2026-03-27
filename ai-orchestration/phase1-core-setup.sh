#!/usr/bin/env bash
# Phase 1: Core Infrastructure & CLI Tooling (AlmaLinux 10 container)
# Run inside AlmaLinux 10: bash phase1-core-setup.sh
# Prereqs: dnf, curl, Node 20+ (install from NodeSource or dnf module nodejs:20)

set -e
CODE_SERVER_PORT="${CODE_SERVER_PORT:-8080}"
CODE_SERVER_PASSWORD="${CODE_SERVER_PASSWORD:-}"

# --- 1.1 code-server (VSCode Web UI) ---
install_code_server() {
  if command -v code-server &>/dev/null; then
    echo "code-server already installed."
    return 0
  fi
  curl -fsSL https://code-server.dev/install.sh | sh
  # Generate password if not set
  if [[ -z "$CODE_SERVER_PASSWORD" ]]; then
    CODE_SERVER_PASSWORD=$(openssl rand -base64 16)
    echo "Generated code-server password (save it): $CODE_SERVER_PASSWORD"
  fi
  export PASSWORD="$CODE_SERVER_PASSWORD"
  echo "Run code-server with: PASSWORD=\$PASSWORD code-server --bind-addr 0.0.0.0:${CODE_SERVER_PORT} --auth password"
}

# --- 1.2 Google Cloud CLI ---
install_gcloud() {
  if command -v gcloud &>/dev/null; then
    echo "gcloud already installed."
    return 0
  fi
  if [[ ! -f /etc/yum.repos.d/google-cloud-sdk.repo ]]; then
    sudo tee /etc/yum.repos.d/google-cloud-sdk.repo <<EOF
[google-cloud-cli]
name=Google Cloud CLI
baseurl=https://packages.cloud.google.com/yum/repos/cloud-sdk-el10-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=0
gpgkey=https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF
  fi
  sudo dnf install -y google-cloud-cli
  echo "Run 'gcloud init' for project/region. Skip 'gcloud auth login' and complete auth manually later."
}

# --- 1.3 OpenCode CLI ---
install_opencode() {
  if command -v opencode &>/dev/null; then
    echo "opencode already installed: $(opencode --version 2>/dev/null || true)"
    return 0
  fi
  if curl -fsSL https://opencode.ai/install 2>/dev/null | bash; then
    echo "OpenCode installed via curl script."
  else
    npm install -g opencode-ai
  fi
  opencode --version
}

# --- 1.4 Cline CLI ---
install_cline() {
  if command -v cline &>/dev/null; then
    echo "cline already installed: $(cline version 2>/dev/null || true)"
    return 0
  fi
  npm install -g cline
  cline version
  # Optional: @yaegaki/cline-cli and init
  # npm install -g @yaegaki/cline-cli && cline-cli init
}

# --- Main ---
install_code_server
install_gcloud
install_opencode
install_cline

echo "Phase 1 complete. code-server password (if generated): $CODE_SERVER_PASSWORD"
echo "Start code-server: PASSWORD=<password> code-server --bind-addr 0.0.0.0:${CODE_SERVER_PORT} --auth password"
