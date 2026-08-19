package backend

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStaticAppStateCleanerRejectsAnEmptyAppNameWithoutRemovingTheUserConfigRoot(t *testing.T) {
	setTestConfigEnv(t)
	configRoot, err := os.UserConfigDir()
	require.NoError(t, err)
	sentinel := filepath.Join(configRoot, "keep.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(sentinel), 0o700))
	require.NoError(t, os.WriteFile(sentinel, []byte("{}"), 0o600))

	err = newStaticAppStateCleaner("").Reset()

	require.ErrorContains(t, err, "empty app name")
	require.FileExists(t, sentinel)
}

func TestNilStaticAppStateCleanerResetIsSafe(t *testing.T) {
	var cleaner *staticAppStateCleaner
	require.NoError(t, cleaner.Reset())
}
