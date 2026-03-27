#!/usr/bin/env bash
# Phase 5: Validation and CLI tests. Run after Phase 1 (and optionally 2) inside AlmaLinux 10.

set -e
echo "=== Phase 5 validation ==="

# OpenCode
if command -v opencode &>/dev/null; then
  echo "OpenCode: $(opencode --version 2>/dev/null || true)"
  opencode run "echo 'Environment Ready'" 2>/dev/null || echo "opencode run skipped (no API key or network)"
else
  echo "OpenCode not installed; run phase1-core-setup.sh"
fi

# Cline
if command -v cline &>/dev/null; then
  echo "Cline: $(cline version 2>/dev/null || true)"
  cline version
else
  echo "Cline not installed; run phase1-core-setup.sh"
fi

echo "=== See SERVICE_SUMMARY.md and .env.example ==="
