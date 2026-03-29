package serve

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kennetholsenatm-gif/omnigraph/internal/repo"
)

const workspaceWatchDebounce = 500 * time.Millisecond

// dirWatchIndex maps absolute directory paths to basenames of files to watch in that directory.
type dirWatchIndex map[string]map[string]struct{}

func buildWorkspaceWatchIndex(root string) (dirWatchIndex, error) {
	disc, err := repo.Discover(root)
	if err != nil {
		return nil, err
	}
	idx := make(dirWatchIndex)
	for _, f := range disc.Files {
		switch f.Kind {
		case repo.KindTerraformState, repo.KindAnsibleInventory:
			full := filepath.Join(disc.Root, filepath.FromSlash(f.Path))
			full, err := filepath.Abs(full)
			if err != nil {
				continue
			}
			d := filepath.Clean(filepath.Dir(full))
			b := filepath.Base(full)
			if idx[d] == nil {
				idx[d] = make(map[string]struct{})
			}
			idx[d][b] = struct{}{}
		}
	}
	return idx, nil
}

func (idx dirWatchIndex) matchesEvent(name string) bool {
	name = filepath.Clean(name)
	dir := filepath.Clean(filepath.Dir(name))
	base := filepath.Base(name)
	for d, basenames := range idx {
		if !strings.EqualFold(dir, d) {
			continue
		}
		_, ok := basenames[base]
		return ok
	}
	return false
}

// runWorkspaceWatch invokes onStable after debounced quiet period following relevant fs changes.
// cleanup must be called to release the watcher; it blocks until the goroutine exits.
func runWorkspaceWatch(ctx context.Context, idx dirWatchIndex, onStable func()) (cleanup func(), err error) {
	if len(idx) == 0 {
		return func() {}, nil
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	for d := range idx {
		if err := w.Add(d); err != nil {
			_ = w.Close()
			return nil, err
		}
	}

	var mu sync.Mutex
	var debounceTimer *time.Timer

	schedule := func() {
		mu.Lock()
		defer mu.Unlock()
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
		debounceTimer = time.AfterFunc(workspaceWatchDebounce, func() {
			onStable()
		})
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-w.Events:
				if !ok {
					return
				}
				if ev.Name == "" {
					continue
				}
				if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove) == 0 {
					continue
				}
				if idx.matchesEvent(ev.Name) {
					schedule()
				}
			case _, ok := <-w.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	return func() {
		_ = w.Close()
		<-done
		mu.Lock()
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
		mu.Unlock()
	}, nil
}
