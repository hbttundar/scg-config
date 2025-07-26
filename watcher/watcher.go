package watcher

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/hbttundar/scg-config/contract"
)

// Watcher provides file watching capabilities for configuration files.
type Watcher struct {
	config   contract.Config
	watcher  *fsnotify.Watcher
	done     chan struct{}
	mu       sync.Mutex
	eventMux sync.Mutex
	wg       sync.WaitGroup
	files    map[string]func()
	started  bool
}

// NewWatcher creates a new Watcher instance.
func NewWatcher(config contract.Config) *Watcher {
	return &Watcher{
		config:   config,
		done:     make(chan struct{}),
		files:    make(map[string]func()),
		watcher:  nil,
		started:  false,
		mu:       sync.Mutex{},
		eventMux: sync.Mutex{},
		wg:       sync.WaitGroup{},
	}
}

// AddFile adds a file to the watcher and registers its callback.
func (w *Watcher) AddFile(path string, callback func()) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.watcher == nil {
		newWatcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("failed to create file watcher: %w", err)
		}

		w.watcher = newWatcher
	}

	if err := w.watcher.Add(path); err != nil {
		return fmt.Errorf("failed to add file to watcher: %w", err)
	}

	w.files[path] = callback
	w.startLocked()

	return nil
}

// Watch starts the watcher loop if not already running.
func (w *Watcher) Watch(callback func()) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for path := range w.files {
		w.files[path] = callback
	}

	w.startLocked()
}

// startLocked starts the watcher goroutine if not already started.
// Assumes the caller holds w.mu.
func (w *Watcher) startLocked() {
	if w.started {
		return
	}

	w.started = true
	w.wg.Add(1)

	go w.run()
}

// run is the goroutine that dispatches file system events.
func (w *Watcher) run() {
	defer w.wg.Done()

	for {
		select {
		case <-w.done:
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			w.handleEvent(event)
		case err, ok := <-w.watcher.Errors:
			// Check if the error channel has been closed
			if !ok {
				return
			}
			// Optionally log the error or handle it here. We assign to the blank identifier
			// to avoid unused variable warnings without leaving a dangling comment as the
			// last statement in the block.
			_ = err
		}
	}
}

// handleEvent is called for every fsnotify event.
func (w *Watcher) handleEvent(event fsnotify.Event) {
	w.eventMux.Lock()
	defer w.eventMux.Unlock()

	if event.Op&fsnotify.Write == fsnotify.Write {
		if reloadable, ok := w.config.(interface{ ReloadConfig() }); ok {
			reloadable.ReloadConfig()
		}

		w.mu.Lock()
		cb := w.files[event.Name]
		w.mu.Unlock()

		if cb != nil {
			cb()
		}
	}
}

// Close stops the watcher.
func (w *Watcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.watcher != nil {
		close(w.done)
		w.wg.Wait()
		err := w.watcher.Close()
		w.watcher = nil
		w.files = make(map[string]func())
		w.started = false

		if err != nil {
			return fmt.Errorf("error closing fsnotify watcher: %w", err)
		}
	}

	return nil
}

// SetConfig sets the config reference.
func (w *Watcher) SetConfig(config contract.Config) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.config = config
}

// GetConfig returns the Config.
//
//nolint:ireturn,nolintlint // returning an interface is required by the contract API
func (w *Watcher) GetConfig() contract.Config {
	return w.config
}

// Compile time checks for interface.
var _ contract.Watcher = (*Watcher)(nil)
