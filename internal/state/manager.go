package state

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Manager provides enhanced state management with drift detection and locking
type Manager struct {
	mu         sync.RWMutex
	rootDir    string
	locks      map[string]*StateLock
	watchers   map[string]*StateWatcher
	history    map[string][]*StateVersion
	driftCache map[string]*DriftResult
}

// StateLock represents a state lock
type StateLock struct {
	ID        string                 `json:"id"`
	Path      string                 `json:"path"`
	Owner     string                 `json:"owner"`
	Operation string                 `json:"operation"`
	LockedAt  time.Time              `json:"lockedAt"`
	ExpiresAt time.Time              `json:"expiresAt"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// StateWatcher monitors state files for changes
type StateWatcher struct {
	Path     string
	Interval time.Duration
	Callback func(path string, oldHash, newHash string)
	stopCh   chan struct{}
}

// StateVersion represents a versioned state snapshot
type StateVersion struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Hash      string    `json:"hash"`
	Size      int64     `json:"size"`
	Timestamp time.Time `json:"timestamp"`
	Author    string    `json:"author,omitempty"`
	Message   string    `json:"message,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
}

// DriftResult represents the result of a drift detection
type DriftResult struct {
	Path         string                 `json:"path"`
	HasDrift     bool                   `json:"hasDrift"`
	ExpectedHash string                 `json:"expectedHash"`
	ActualHash   string                 `json:"actualHash"`
	CheckedAt    time.Time              `json:"checkedAt"`
	Changes      []DriftChange          `json:"changes,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// DriftChange represents a single drift change
type DriftChange struct {
	Type     string      `json:"type"` // added, removed, modified
	Path     string      `json:"path"`
	OldValue interface{} `json:"oldValue,omitempty"`
	NewValue interface{} `json:"newValue,omitempty"`
}

// NewManager creates a new state manager
func NewManager(rootDir string) (*Manager, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(absRoot, 0755); err != nil {
		return nil, err
	}

	return &Manager{
		rootDir:    absRoot,
		locks:      make(map[string]*StateLock),
		watchers:   make(map[string]*StateWatcher),
		history:    make(map[string][]*StateVersion),
		driftCache: make(map[string]*DriftResult),
	}, nil
}

// AcquireLock acquires a lock on a state file
func (m *Manager) AcquireLock(ctx context.Context, path, owner, operation string, ttl time.Duration) (*StateLock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, err := m.resolvePath(path)
	if err != nil {
		return nil, err
	}

	// Check if already locked
	if existing, ok := m.locks[absPath]; ok {
		if time.Now().Before(existing.ExpiresAt) {
			return nil, fmt.Errorf("state already locked by %s (expires at %s)", existing.Owner, existing.ExpiresAt)
		}
		// Lock expired, remove it
		delete(m.locks, absPath)
	}

	lock := &StateLock{
		ID:        generateLockID(),
		Path:      absPath,
		Owner:     owner,
		Operation: operation,
		LockedAt:  time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	m.locks[absPath] = lock

	// Write lock file
	lockFile := absPath + ".lock"
	lockData, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(lockFile, lockData, 0644); err != nil {
		delete(m.locks, absPath)
		return nil, err
	}

	return lock, nil
}

// ReleaseLock releases a lock on a state file
func (m *Manager) ReleaseLock(path, lockID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, err := m.resolvePath(path)
	if err != nil {
		return err
	}

	lock, ok := m.locks[absPath]
	if !ok {
		return fmt.Errorf("no lock found for %s", path)
	}

	if lock.ID != lockID {
		return fmt.Errorf("lock ID mismatch: expected %s, got %s", lock.ID, lockID)
	}

	delete(m.locks, absPath)

	// Remove lock file
	lockFile := absPath + ".lock"
	if err := os.Remove(lockFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// ListLocks lists all active locks
func (m *Manager) ListLocks() []*StateLock {
	m.mu.RLock()
	defer m.mu.RUnlock()

	locks := make([]*StateLock, 0, len(m.locks))
	for _, lock := range m.locks {
		if time.Now().Before(lock.ExpiresAt) {
			locks = append(locks, lock)
		}
	}
	return locks
}

// DetectDrift detects drift in a state file
func (m *Manager) DetectDrift(ctx context.Context, path string, expectedHash string) (*DriftResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, err := m.resolvePath(path)
	if err != nil {
		return nil, err
	}

	// Calculate actual hash
	actualHash, err := m.calculateFileHash(absPath)
	if err != nil {
		return nil, err
	}

	result := &DriftResult{
		Path:         absPath,
		ExpectedHash: expectedHash,
		ActualHash:   actualHash,
		HasDrift:     expectedHash != actualHash,
		CheckedAt:    time.Now(),
	}

	// If drift detected, try to analyze changes
	if result.HasDrift {
		changes, err := m.analyzeChanges(absPath, expectedHash, actualHash)
		if err == nil {
			result.Changes = changes
		}
	}

	m.driftCache[absPath] = result

	return result, nil
}

// GetDriftResult returns cached drift result
func (m *Manager) GetDriftResult(path string) (*DriftResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	absPath, err := m.resolvePath(path)
	if err != nil {
		return nil, err
	}

	result, ok := m.driftCache[absPath]
	if !ok {
		return nil, fmt.Errorf("no drift result found for %s", path)
	}

	return result, nil
}

// CreateVersion creates a new version of a state file
func (m *Manager) CreateVersion(ctx context.Context, path, author, message string, tags []string) (*StateVersion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, err := m.resolvePath(path)
	if err != nil {
		return nil, err
	}

	hash, err := m.calculateFileHash(absPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	version := &StateVersion{
		ID:        generateVersionID(),
		Path:      absPath,
		Hash:      hash,
		Size:      info.Size(),
		Timestamp: time.Now(),
		Author:    author,
		Message:   message,
		Tags:      tags,
	}

	m.history[absPath] = append(m.history[absPath], version)

	// Save version metadata
	versionDir := filepath.Join(m.rootDir, ".versions")
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return nil, err
	}

	versionFile := filepath.Join(versionDir, fmt.Sprintf("%s-%s.json", filepath.Base(absPath), version.ID))
	versionData, err := json.MarshalIndent(version, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(versionFile, versionData, 0644); err != nil {
		return nil, err
	}

	return version, nil
}

// GetVersionHistory returns version history for a state file
func (m *Manager) GetVersionHistory(path string, limit int) ([]*StateVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	absPath, err := m.resolvePath(path)
	if err != nil {
		return nil, err
	}

	versions, ok := m.history[absPath]
	if !ok {
		return nil, nil
	}

	if limit <= 0 || limit > len(versions) {
		limit = len(versions)
	}

	return versions[len(versions)-limit:], nil
}

// Watch starts watching a state file for changes
func (m *Manager) Watch(path string, interval time.Duration, callback func(path string, oldHash, newHash string)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, err := m.resolvePath(path)
	if err != nil {
		return err
	}

	if _, ok := m.watchers[absPath]; ok {
		return fmt.Errorf("already watching %s", path)
	}

	watcher := &StateWatcher{
		Path:     absPath,
		Interval: interval,
		Callback: callback,
		stopCh:   make(chan struct{}),
	}

	m.watchers[absPath] = watcher

	go m.runWatcher(watcher)

	return nil
}

// Unwatch stops watching a state file
func (m *Manager) Unwatch(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, err := m.resolvePath(path)
	if err != nil {
		return err
	}

	watcher, ok := m.watchers[absPath]
	if !ok {
		return fmt.Errorf("not watching %s", path)
	}

	close(watcher.stopCh)
	delete(m.watchers, absPath)

	return nil
}

// runWatcher runs the state watcher
func (m *Manager) runWatcher(watcher *StateWatcher) {
	var lastHash string

	// Get initial hash
	hash, err := m.calculateFileHash(watcher.Path)
	if err == nil {
		lastHash = hash
	}

	ticker := time.NewTicker(watcher.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-watcher.stopCh:
			return
		case <-ticker.C:
			hash, err := m.calculateFileHash(watcher.Path)
			if err != nil {
				continue
			}

			if hash != lastHash {
				oldHash := lastHash
				lastHash = hash
				watcher.Callback(watcher.Path, oldHash, hash)
			}
		}
	}
}

// calculateFileHash calculates SHA256 hash of a file
func (m *Manager) calculateFileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

// analyzeChanges analyzes changes between two state versions
func (m *Manager) analyzeChanges(path, oldHash, newHash string) ([]DriftChange, error) {
	// Read old version from history
	versions, ok := m.history[path]
	if !ok || len(versions) == 0 {
		return nil, fmt.Errorf("no version history found")
	}

	// Find old version
	var oldVersion *StateVersion
	for _, v := range versions {
		if v.Hash == oldHash {
			oldVersion = v
			break
		}
	}

	if oldVersion == nil {
		return nil, fmt.Errorf("old version not found in history")
	}

	// For now, return a simple change indicating modification
	// In a real implementation, this would parse JSON and compare
	return []DriftChange{
		{
			Type: "modified",
			Path: path,
		},
	}, nil
}

// resolvePath resolves a path relative to the root directory
func (m *Manager) resolvePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	return filepath.Abs(filepath.Join(m.rootDir, path))
}

// generateLockID generates a unique lock ID
func generateLockID() string {
	return fmt.Sprintf("lock-%d", time.Now().UnixNano())
}

// generateVersionID generates a unique version ID
func generateVersionID() string {
	return fmt.Sprintf("v-%d", time.Now().UnixNano())
}

// CleanupExpiredLocks removes expired locks
func (m *Manager) CleanupExpiredLocks() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for path, lock := range m.locks {
		if now.After(lock.ExpiresAt) {
			delete(m.locks, path)
			// Remove lock file
			lockFile := path + ".lock"
			os.Remove(lockFile)
		}
	}
}

// GetStats returns statistics about the state manager
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"activeLocks":    len(m.locks),
		"activeWatchers": len(m.watchers),
		"cachedDrifts":   len(m.driftCache),
		"versionHistory": len(m.history),
	}
}
