#!/usr/bin/env bash
# Install OmniGraph Git hooks
# This script installs the pre-commit hook for policy validation

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOOKS_DIR="$SCRIPT_DIR/hooks"
GIT_HOOKS_DIR=".git/hooks"

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo -e "${RED}Error: Not in a git repository${NC}"
    echo "Please run this script from the root of a git repository"
    exit 1
fi

# Create hooks directory if it doesn't exist
mkdir -p "$GIT_HOOKS_DIR"

# Install pre-commit hook
echo -e "${YELLOW}Installing pre-commit hook...${NC}"

if [ -f "$HOOKS_DIR/pre-commit" ]; then
    cp "$HOOKS_DIR/pre-commit" "$GIT_HOOKS_DIR/pre-commit"
    chmod +x "$GIT_HOOKS_DIR/pre-commit"
    echo -e "${GREEN}Pre-commit hook installed successfully${NC}"
else
    echo -e "${RED}Error: pre-commit hook not found at $HOOKS_DIR/pre-commit${NC}"
    exit 1
fi

# Check if omnigraph is available
if command -v omnigraph &> /dev/null; then
    echo -e "${GREEN}OmniGraph found in PATH${NC}"
    
    # Test the hook
    echo -e "${YELLOW}Testing pre-commit hook...${NC}"
    if "$GIT_HOOKS_DIR/pre-commit" --help &> /dev/null 2>&1; then
        echo -e "${GREEN}Pre-commit hook test passed${NC}"
    else
        echo -e "${YELLOW}Warning: Pre-commit hook test failed (this is normal if no policy files exist)${NC}"
    fi
else
    echo -e "${YELLOW}Warning: OmniGraph not found in PATH${NC}"
    echo "Please ensure 'omnigraph' is installed and available in your PATH"
    echo "Or set the OMNIGRAPH environment variable to point to the omnigraph binary"
fi

echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo ""
echo "The pre-commit hook will now validate:"
echo "  - Rego policy files (*.rego)"
echo "  - Policy YAML/JSON files (containing omnigraph/policy/v1)"
echo "  - Schema files (*.omnigraph.schema)"
echo ""
echo "To bypass the hook for a specific commit, use:"
echo "  git commit --no-verify"
echo ""
echo "To uninstall the hook, delete:"
echo "  $GIT_HOOKS_DIR/pre-commit"