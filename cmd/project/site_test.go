package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublishSiteVersionSkipsBeta(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	require.NoError(t, publishSiteVersion("v2.0.0-beta.1"))
}
