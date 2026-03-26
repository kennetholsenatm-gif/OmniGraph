package safepath

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UnderRoot returns an absolute path under root joined with rel, rejecting escapes.
// rel may use slash or native separators; absolute rel is rejected.
func UnderRoot(root, rel string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("empty root")
	}
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", err
	}
	rp := filepath.FromSlash(strings.TrimSpace(rel))
	if rp == "" || rp == "." {
		return "", fmt.Errorf("empty relative path")
	}
	if filepath.IsAbs(rp) {
		return "", fmt.Errorf("absolute subpath not allowed")
	}
	full := filepath.Join(absRoot, rp)
	full = filepath.Clean(full)
	out, err := filepath.Rel(absRoot, full)
	if err != nil {
		return "", err
	}
	if out == ".." || strings.HasPrefix(out, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root directory")
	}
	return full, nil
}

// UnderDir is like UnderRoot but root may be relative; it is cleaned with Abs using cwd.
func UnderDir(dir, rel string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("empty directory")
	}
	absDir, err := filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return "", err
	}
	return UnderRoot(absDir, rel)
}

// SSHKeyUnderHome resolves keyPath to an absolute path under the user's home directory.
// Empty keyPath defaults to ~/.ssh/id_rsa (resolved absolutely).
func SSHKeyUnderHome(keyPath string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home: %w", err)
	}
	absHome, err := filepath.Abs(filepath.Clean(home))
	if err != nil {
		return "", err
	}
	k := strings.TrimSpace(keyPath)
	if k == "" {
		return UnderRoot(absHome, ".ssh/id_rsa")
	}
	k = filepath.Clean(k)
	if !filepath.IsAbs(k) {
		k = filepath.Join(absHome, k)
	}
	k = filepath.Clean(k)
	rel, err := filepath.Rel(absHome, k)
	if err != nil {
		return "", fmt.Errorf("ssh key path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("ssh private key must be under home directory %q", absHome)
	}
	return k, nil
}
