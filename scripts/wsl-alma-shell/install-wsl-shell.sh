#!/usr/bin/env bash
# Install readline + bash snippets into the current WSL user's home.
# Idempotent: safe to re-run.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="${HOME}/.config/omnigraph-wsl"
INPUTRC_TARGET="${CONFIG_DIR}/inputrc"
BASHRC_TARGET="${CONFIG_DIR}/bashrc"
MARKER_BEGIN="# >>> omnigraph-wsl-shell"
MARKER_END="# <<< omnigraph-wsl-shell"

mkdir -p "${CONFIG_DIR}"
cp -f "${SCRIPT_DIR}/inputrc.snippet" "${INPUTRC_TARGET}"
cp -f "${SCRIPT_DIR}/bashrc.snippet" "${BASHRC_TARGET}"

touch "${HOME}/.inputrc"
if ! grep -qF 'omnigraph-wsl/inputrc' "${HOME}/.inputrc" 2>/dev/null; then
	{
		echo "${MARKER_BEGIN}"
		printf '$include %s\n' "${INPUTRC_TARGET}"
		echo "${MARKER_END}"
	} >>"${HOME}/.inputrc"
fi

touch "${HOME}/.bashrc"
if ! grep -qF 'omnigraph-wsl/bashrc' "${HOME}/.bashrc" 2>/dev/null; then
	{
		echo "${MARKER_BEGIN}"
		printf '[[ -f %q ]] && source %q\n' "${BASHRC_TARGET}" "${BASHRC_TARGET}"
		echo "${MARKER_END}"
	} >>"${HOME}/.bashrc"
fi

echo "Installed:"
echo "  ${INPUTRC_TARGET}"
echo "  ${BASHRC_TARGET}"
echo "Updated: ~/.inputrc (include), ~/.bashrc (source)"
echo "Open a new shell or run: bind -f ~/.inputrc   (bash)"
