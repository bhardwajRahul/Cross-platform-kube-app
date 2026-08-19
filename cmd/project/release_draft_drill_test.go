package main

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfiguredReleaseDraftDrillUsesExplicitDisposableRepository(t *testing.T) {
	t.Setenv(releaseDraftDrillRepositoryEnv, " luxury-yacht/release-drills ")
	t.Setenv(releaseDraftDrillConfirmationEnv, releaseDraftDrillConfirmation)
	now := time.Date(2026, time.August, 18, 12, 0, 0, 123, time.FixedZone("test", -6*60*60))

	cfg := configuredReleaseDraftDrill(now)

	require.Equal(t, "luxury-yacht/release-drills", cfg.repository)
	require.Equal(t, "draft-recovery-drill-20260818T180000.000000123Z", cfg.tag)
	require.Equal(t, releaseDraftDrillConfirmation, cfg.confirmation)
}

func TestValidateReleaseDraftDrillConfigRejectsUnsafeTargets(t *testing.T) {
	valid := releaseDraftDrillConfig{
		repository:   "luxury-yacht/release-drills",
		tag:          "draft-recovery-drill-20260818T120000Z",
		confirmation: releaseDraftDrillConfirmation,
	}
	require.NoError(t, validateReleaseDraftDrillConfig(valid))

	tests := []struct {
		name   string
		mutate func(*releaseDraftDrillConfig)
		want   string
	}{
		{
			name: "missing confirmation",
			mutate: func(cfg *releaseDraftDrillConfig) {
				cfg.confirmation = ""
			},
			want: "explicit confirmation",
		},
		{
			name: "production repository",
			mutate: func(cfg *releaseDraftDrillConfig) {
				cfg.repository = projectReleaseRepo
			},
			want: "must not use the production release repository",
		},
		{
			name: "invalid repository",
			mutate: func(cfg *releaseDraftDrillConfig) {
				cfg.repository = "release-drills"
			},
			want: "owner/repository",
		},
		{
			name: "unscoped tag",
			mutate: func(cfg *releaseDraftDrillConfig) {
				cfg.tag = "v2.0.0"
			},
			want: "draft-recovery-drill-",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := valid
			test.mutate(&cfg)

			require.ErrorContains(t, validateReleaseDraftDrillConfig(cfg), test.want)
		})
	}
}

func TestRunReleaseDraftDrillCreatesInspectsAndDeletesFailedDraft(t *testing.T) {
	cfg := releaseDraftDrillConfig{
		repository:   "luxury-yacht/release-drills",
		tag:          "draft-recovery-drill-20260818T120000Z",
		confirmation: releaseDraftDrillConfirmation,
	}
	var calls [][]string
	releaseExists := false
	releaseIsDraft := false
	run := func(name string, args ...string) error {
		call := append([]string{name}, args...)
		calls = append(calls, call)
		switch {
		case isReleaseCommand(args, "create"):
			require.Contains(t, args, "--draft")
			releaseExists = true
			releaseIsDraft = true
			return nil
		case isReleaseCommand(args, "delete"):
			require.True(t, releaseExists)
			releaseExists = false
			releaseIsDraft = false
			return nil
		default:
			return errors.New("unexpected command")
		}
	}
	output := func(_ string, args ...string) (string, error) {
		require.Equal(t, []string{
			"release", "view", cfg.tag,
			"--repo", cfg.repository,
			"--json", "isDraft",
			"--jq", ".isDraft",
		}, args)
		if !releaseExists {
			return "", errors.New("release not found")
		}
		if releaseIsDraft {
			return "true", nil
		}
		return "false", nil
	}

	err := runReleaseDraftDrill(cfg, run, output)

	require.NoError(t, err)
	require.False(t, releaseExists)
	require.Len(t, calls, 2)
	require.Equal(t, []string{"gh", "release", "create", cfg.tag}, calls[0][:4])
	require.Equal(t, []string{"gh", "release", "delete", cfg.tag}, calls[1][:4])
}

func TestRunReleaseDraftDrillReportsCleanupFailure(t *testing.T) {
	cfg := releaseDraftDrillConfig{
		repository:   "luxury-yacht/release-drills",
		tag:          "draft-recovery-drill-20260818T120000Z",
		confirmation: releaseDraftDrillConfirmation,
	}
	releaseExists := false
	run := func(_ string, args ...string) error {
		switch {
		case isReleaseCommand(args, "create"):
			releaseExists = true
			return nil
		case isReleaseCommand(args, "delete"):
			return errors.New("delete failed")
		default:
			return errors.New("unexpected command")
		}
	}
	output := func(_ string, _ ...string) (string, error) {
		if !releaseExists {
			return "", errors.New("release not found")
		}
		return "true", nil
	}

	err := runReleaseDraftDrill(cfg, run, output)

	require.ErrorContains(t, err, "delete disposable draft")
	require.ErrorContains(t, err, "remove it manually before retrying")
}

func TestRunReleaseDraftDrillRejectsExistingReleaseWithoutMutation(t *testing.T) {
	cfg := releaseDraftDrillConfig{
		repository:   "luxury-yacht/release-drills",
		tag:          "draft-recovery-drill-20260818T120000Z",
		confirmation: releaseDraftDrillConfirmation,
	}
	run := func(_ string, _ ...string) error {
		t.Fatal("release mutation must not run when the tag already exists")
		return nil
	}

	err := runReleaseDraftDrill(cfg, run, func(_ string, _ ...string) (string, error) {
		return "true", nil
	})

	require.ErrorContains(t, err, "already exists")
}

func TestRunReleaseDraftDrillDeletesUnexpectedPublicRelease(t *testing.T) {
	cfg := releaseDraftDrillConfig{
		repository:   "luxury-yacht/release-drills",
		tag:          "draft-recovery-drill-20260818T120000Z",
		confirmation: releaseDraftDrillConfirmation,
	}
	releaseExists := false
	deleted := false
	run := func(_ string, args ...string) error {
		switch {
		case isReleaseCommand(args, "create"):
			releaseExists = true
			return nil
		case isReleaseCommand(args, "delete"):
			deleted = true
			releaseExists = false
			return nil
		default:
			return errors.New("unexpected command")
		}
	}
	output := func(_ string, _ ...string) (string, error) {
		if !releaseExists {
			return "", errors.New("release not found")
		}
		return "false", nil
	}

	err := runReleaseDraftDrill(cfg, run, output)

	require.ErrorContains(t, err, "became public")
	require.True(t, deleted)
}

func TestInspectReleaseDraftRejectsUnexpectedState(t *testing.T) {
	isDraft, err := inspectReleaseDraft("owner/repo", "tag", func(_ string, _ ...string) (string, error) {
		return "false", nil
	})
	require.NoError(t, err)
	require.False(t, isDraft)

	_, err = inspectReleaseDraft("owner/repo", "tag", func(_ string, _ ...string) (string, error) {
		return "unknown", nil
	})
	require.ErrorContains(t, err, "unexpected isDraft value")
}

func TestWriteReleaseDraftDrillFileWritesAndRejectsInvalidPattern(t *testing.T) {
	path, err := writeReleaseDraftDrillFile("release-draft-drill-test-*.txt", "payload")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.Remove(path)) })
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "payload", string(contents))

	_, err = writeReleaseDraftDrillFile("missing/draft-*.txt", "payload")
	require.Error(t, err)
}
