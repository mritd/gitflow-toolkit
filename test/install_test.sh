#!/bin/bash
# Integration tests for gitflow-toolkit installation
set -e

BINARY="/tmp/gitflow-toolkit"
INSTALL_DIR="/usr/local/bin"
HOME_DIR="$HOME/.gitflow-toolkit"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; exit 1; }
info() { echo -e "${YELLOW}→ $1${NC}"; }

# Test 1: Install without hook
test_install_no_hook() {
    info "Test: Install without hook"
    
    sudo "$BINARY" install -d "$INSTALL_DIR" 2>/dev/null || true
    
    # Check binary exists
    [ -f "$INSTALL_DIR/gitflow-toolkit" ] || fail "Binary not installed"
    
    # Check symlinks
    [ -L "$INSTALL_DIR/git-ci" ] || fail "git-ci symlink missing"
    [ -L "$INSTALL_DIR/git-ps" ] || fail "git-ps symlink missing"
    [ -L "$INSTALL_DIR/git-feat" ] || fail "git-feat symlink missing"
    [ -L "$INSTALL_DIR/git-fix" ] || fail "git-fix symlink missing"
    
    # Check home directory ownership (should be owned by testuser, not root)
    if [ -d "$HOME_DIR" ]; then
        OWNER=$(stat -c '%U' "$HOME_DIR" 2>/dev/null || stat -f '%Su' "$HOME_DIR")
        [ "$OWNER" = "testuser" ] || fail "Home directory owned by $OWNER, not testuser"
    fi
    
    # Hook should NOT be installed
    [ ! -f "$HOME_DIR/hooks/commit-msg" ] || fail "Hook should not be installed"
    
    pass "Install without hook"
}

# Test 2: Install with hook
test_install_with_hook() {
    info "Test: Install with hook"
    
    # Uninstall first
    sudo "$BINARY" uninstall -d "$INSTALL_DIR" 2>/dev/null || true
    
    sudo "$BINARY" install -d "$INSTALL_DIR" --hook 2>/dev/null || true
    
    # Check hook symlink exists
    [ -L "$HOME_DIR/hooks/commit-msg" ] || fail "Hook symlink missing"
    
    # Check git config
    HOOKS_PATH=$(git config --global core.hooksPath 2>/dev/null || echo "")
    [ "$HOOKS_PATH" = "$HOME_DIR/hooks" ] || fail "Git hooksPath not set correctly: $HOOKS_PATH"
    
    # Check ownership
    OWNER=$(stat -c '%U' "$HOME_DIR/hooks" 2>/dev/null || stat -f '%Su' "$HOME_DIR/hooks")
    [ "$OWNER" = "testuser" ] || fail "Hooks directory owned by $OWNER, not testuser"
    
    pass "Install with hook"
}

# Test 3: Uninstall
test_uninstall() {
    info "Test: Uninstall"
    
    sudo "$BINARY" uninstall -d "$INSTALL_DIR" 2>/dev/null || true
    
    # Check binary removed
    [ ! -f "$INSTALL_DIR/gitflow-toolkit" ] || fail "Binary not removed"
    
    # Check symlinks removed
    [ ! -L "$INSTALL_DIR/git-ci" ] || fail "git-ci symlink not removed"
    [ ! -L "$INSTALL_DIR/git-ps" ] || fail "git-ps symlink not removed"
    
    # Check home directory removed
    [ ! -d "$HOME_DIR" ] || fail "Home directory not removed"
    
    # Check git config unset
    HOOKS_PATH=$(git config --global core.hooksPath 2>/dev/null || echo "")
    [ -z "$HOOKS_PATH" ] || fail "Git hooksPath not unset: $HOOKS_PATH"
    
    pass "Uninstall"
}

# Test 4: Git subcommand invocation
test_git_subcommands() {
    info "Test: Git subcommand invocation"
    
    # Install first
    sudo "$BINARY" install -d "$INSTALL_DIR" 2>/dev/null || true
    
    # Test help output
    "$INSTALL_DIR/git-ci" --help | grep -q "commit" || fail "git-ci --help failed"
    "$INSTALL_DIR/git-ps" --help | grep -q "push" || fail "git-ps --help failed"
    "$INSTALL_DIR/git-feat" --help | grep -q "feat" || fail "git-feat --help failed"
    
    pass "Git subcommand invocation"
}

# Test 5: Commit message hook validation
test_commit_hook() {
    info "Test: Commit message hook validation"
    
    # Create temp file with valid message
    VALID_MSG=$(mktemp)
    echo "feat(api): add new endpoint" > "$VALID_MSG"
    "$BINARY" hook commit-msg "$VALID_MSG" || fail "Valid message rejected"
    rm "$VALID_MSG"
    
    # Create temp file with invalid message
    INVALID_MSG=$(mktemp)
    echo "random commit message" > "$INVALID_MSG"
    if "$BINARY" hook commit-msg "$INVALID_MSG" 2>/dev/null; then
        rm "$INVALID_MSG"
        fail "Invalid message accepted"
    fi
    rm "$INVALID_MSG"
    
    pass "Commit message hook validation"
}

# Test 6: Version command
test_version() {
    info "Test: Version command"
    
    "$BINARY" --version | grep -q "gitflow-toolkit" || fail "Version output incorrect"
    
    pass "Version command"
}

# Run all tests
echo "=========================================="
echo "  gitflow-toolkit Integration Tests"
echo "=========================================="
echo ""

test_version
test_install_no_hook
test_install_with_hook
test_git_subcommands
test_commit_hook
test_uninstall

echo ""
echo "=========================================="
echo -e "  ${GREEN}All tests passed!${NC}"
echo "=========================================="
