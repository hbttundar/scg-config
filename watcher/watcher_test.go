package watcher_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hbttundar/scg-config/config"
)

func TestWatcher(t *testing.T) {
	t.Parallel()

	t.Run("WatchWithCallback", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "test.yaml")

		require.NoError(t, os.WriteFile(configFile, []byte("test: value"), 0o600))

		cfg := config.New()
		watcher := cfg.Watcher()

		defer func() { _ = watcher.Close() }()

		callbackCalled := make(chan struct{}, 1)

		require.NoError(t, watcher.AddFile(configFile, func() {
			select {
			case callbackCalled <- struct{}{}:
			default:
			}
		}))

		time.Sleep(200 * time.Millisecond)

		require.NoError(t, os.WriteFile(configFile, []byte("test: modified"), 0o600))
		require.NoError(t, os.Chtimes(configFile, time.Now(), time.Now()))

		select {
		case <-callbackCalled:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatal("Callback was not called within timeout")
		}
	})

	t.Run("CloseWatcher", func(t *testing.T) {
		t.Parallel()

		cfg := config.New()
		watcher := cfg.Watcher()
		require.NoError(t, watcher.Close())
	})

	t.Run("WatchNonExistentFile", func(t *testing.T) {
		t.Parallel()

		cfg := config.New()
		watcher := cfg.Watcher()

		defer func() { _ = watcher.Close() }()

		err := watcher.AddFile("/non/existent/file.yaml", func() {})
		require.Error(t, err)
	})
}
