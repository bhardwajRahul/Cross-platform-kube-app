package backend

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/luxury-yacht/app/backend/refresh/system"
)

// ResetRuntimeState unpublishes refresh transports before clearing transient
// resource caches and the app-owned cache tree. It is safe to call repeatedly.
func (a *RefreshCoordinator) ResetRuntimeState() error {
	if a == nil {
		return nil
	}
	a.teardownRefreshSubsystem()
	if a.resources != nil {
		a.resources.clearCaches()
	}
	if a.preferences == nil {
		return nil
	}
	cacheRoot, err := a.preferences.cacheDirPath()
	if err != nil {
		return err
	}
	return os.RemoveAll(cacheRoot)
}

func (a *RefreshCoordinator) teardownRefreshSubsystem() {
	// Reverse publication before stopping producers so no new request or stream
	// can resolve a generation while it is being torn down.
	a.refreshService.Store(nil)
	aggregates := a.refreshAggregates.Load()
	a.refreshAggregates.Store(nil)
	if aggregates != nil && aggregates.resources != nil {
		aggregates.resources.Stop()
	}
	subsystems := a.replaceRefreshSubsystems(nil)
	a.stopObjectCatalog()

	for clusterID, subsystem := range subsystems {
		a.stopRefreshGeneration(clusterID, subsystem)
	}
	a.stopRemainingRefreshGenerationRuntimes()
	a.stopRefreshRuntimeContext()

	a.setTelemetryRecorder(nil)
}

func (a *RefreshCoordinator) handlePermissionIssues(issues []system.PermissionIssue) {
	if a == nil || a.logger == nil {
		return
	}
	for _, issue := range issues {
		if issue.Err == nil {
			continue
		}
		a.logger.Warn(
			fmt.Sprintf("Refresh domain %s unavailable (%s): %v", issue.Domain, issue.Resource, issue.Err),
			"Refresh",
		)
		// NOTE: Per-cluster auth recovery is now handled by the auth manager via 401 responses.
		// Permission issues without cluster context are logged but not auto-recovered.
	}
}

// transportFailureState tracks transport failures for a single cluster.
// This allows isolated recovery per-cluster without affecting other clusters.
type transportFailureState struct {
	mu                sync.Mutex
	failureCount      int
	windowStart       time.Time
	rebuildInProgress bool
	lastRebuild       time.Time
}
