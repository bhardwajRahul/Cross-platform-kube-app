package backend

import (
	"errors"
	"fmt"
	"os"

	"github.com/luxury-yacht/app/internal/appstate"
)

// staticAppStateCleaner removes the app-owned config and cache roots after
// their live owners have quiesced and reset their in-memory state.
type staticAppStateCleaner struct {
	appName string
}

func newStaticAppStateCleaner(appName string) *staticAppStateCleaner {
	return &staticAppStateCleaner{appName: appName}
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
