#!/usr/bin/env bash
# Install Playwright Chromium for the invoking user (no apt-get — works on Alma).
# Playwright's 'install-deps' targets Debian/Ubuntu; on Alma use this or dnf install equivalent libs manually.
set -euo pipefail
require_cmd() { command -v "$1" >/dev/null 2>&1 || { echo "missing: $1" >&2; exit 1; }; }
require_cmd python3
python3 -m pip install -q --user playwright 2>/dev/null || true
python3 -m playwright install chromium
