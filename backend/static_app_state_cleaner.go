package backend

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/luxury-yacht/app/internal/appstate"
)

const appAtomicWriteTempPrefix = ".tmp-"

// staticAppStateCleaner removes stale atomic-write artifacts at process
// startup and removes the app-owned config and cache roots after their live
// owners have quiesced during Factory Reset.
type staticAppStateCleaner struct {
	appName string
}

func newStaticAppStateCleaner(appName string) *staticAppStateCleaner {
	return &staticAppStateCleaner{appName: appName}
}

func (c *staticAppStateCleaner) CleanupStaleWrites() error {
	if c == nil {
		return nil
	}
	manifest, err := appstate.Resolve(c.appName)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(manifest.ConfigRoot)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read app config root %s: %w", manifest.ConfigRoot, err)
	}

	var failures []error
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !isAppAtomicWriteTemp(entry.Name()) {
			continue
		}
		path := filepath.Join(manifest.ConfigRoot, entry.Name())
		if removeErr := os.Remove(path); removeErr != nil {
			failures = append(failures, fmt.Errorf("remove stale app state write %s: %w", path, removeErr))
		}
	}
	return errors.Join(failures...)
}

func isAppAtomicWriteTemp(name string) bool {
	suffix := strings.TrimPrefix(name, appAtomicWriteTempPrefix)
	if suffix == name || suffix == "" {
		return false
	}
	for _, character := range suffix {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func (c *staticAppStateCleaner) Reset() error {
	if c == nil {
		return nil
	}
	manifest, err := appstate.Resolve(c.appName)
	if err != nil {
		return err
	}
	var failures []error
	for _, root := range manifest.StaticRoots() {
		if removeErr := os.RemoveAll(root); removeErr != nil {
			failures = append(failures, fmt.Errorf("remove app state root %s: %w", root, removeErr))
		}
	}
	return errors.Join(failures...)
}
