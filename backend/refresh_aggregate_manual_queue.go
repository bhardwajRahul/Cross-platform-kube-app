package backend

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/luxury-yacht/app/backend/refresh"
	"github.com/luxury-yacht/app/backend/refresh/system"
)

// aggregateManualQueue routes cluster-scoped manual refresh jobs to per-cluster queues.
type aggregateManualQueue struct {
	clusterOrder []string
	queues       map[string]refresh.ManualQueue

	mu       sync.RWMutex
	configMu sync.RWMutex
	jobs     map[string]*aggregateManualJob
}

// aggregateManualJob tracks the child job created for a cluster-scoped refresh.
type aggregateManualJob struct {
	job         *refresh.ManualRefreshJob
	clusterJobs map[string]aggregateManualChildJob
}

// aggregateManualChildJob binds a child job to the queue that owns its status.
// A terminal child remains there; an unfinished child moves to a replacement
// queue when its cluster subsystem is rebuilt.
type aggregateManualChildJob struct {
	jobID string
	queue refresh.ManualQueue
}

type aggregateManualJobMigration struct {
	aggregateJobID string
	clusterID      string
	domain         string
	scope          string
	reason         string
	previous       aggregateManualChildJob
	replacement    refresh.ManualQueue
}

func newAggregateManualQueue(clusterOrder []string, subsystems map[string]*system.Subsystem) *aggregateManualQueue {
	queues := make(map[string]refresh.ManualQueue)
	for id, subsystem := range subsystems {
		if subsystem == nil || subsystem.ManualQueue == nil {
			continue
		}
		queues[id] = subsystem.ManualQueue
	}

	ordered := make([]string, 0, len(clusterOrder))
	for _, id := range clusterOrder {
		if _, ok := queues[id]; ok {
			ordered = append(ordered, id)
		}
	}
	if len(ordered) == 0 {
		for id := range queues {
			ordered = append(ordered, id)
		}
		sort.Strings(ordered)
	}
	return &aggregateManualQueue{
		clusterOrder: ordered,
		queues:       queues,
		jobs:         make(map[string]*aggregateManualJob),
	}
}

// Enqueue registers a manual refresh job for exactly one target cluster.
func (q *aggregateManualQueue) Enqueue(ctx context.Context, domain, scope, reason string) (*refresh.ManualRefreshJob, error) {
	if domain == "" {
		return nil, errors.New("domain is required")
	}
	queues := q.snapshotConfig()
	clusterIDs, scopeValue := refresh.SplitClusterScopeList(scope)
	target, err := q.resolveTarget(domain, clusterIDs, queues)
	if err != nil {
		return nil, err
	}
	queue := queues[target]
	if queue == nil {
		return nil, fmt.Errorf("manual queue unavailable for %s", target)
	}
	scoped := refresh.JoinClusterScope(target, scopeValue)
	job, err := queue.Enqueue(ctx, domain, scoped, reason)
	if err != nil {
		return nil, err
	}
	clusterJobs := map[string]aggregateManualChildJob{
		target: {jobID: job.ID, queue: queue},
	}

	aggregateJob := &refresh.ManualRefreshJob{
		ID:       generateAggregateJobID(),
		Domain:   domain,
		Scope:    scope,
		Reason:   reason,
		State:    refresh.JobStateQueued,
		QueuedAt: time.Now().UnixMilli(),
	}

	q.mu.Lock()
	q.jobs[aggregateJob.ID] = &aggregateManualJob{job: aggregateJob, clusterJobs: clusterJobs}
	q.mu.Unlock()

	return aggregateJob, nil
}

// Status returns the aggregate job state mirrored from its per-cluster child job.
func (q *aggregateManualQueue) Status(jobID string) (*refresh.ManualRefreshJob, bool) {
	q.mu.RLock()
	agg := q.jobs[jobID]
	if agg == nil {
		q.mu.RUnlock()
		return nil, false
	}
	base := *agg.job
	clusterJobs := make(map[string]aggregateManualChildJob, len(agg.clusterJobs))
	for id, child := range agg.clusterJobs {
		clusterJobs[id] = child
	}
	q.mu.RUnlock()

	return buildAggregateStatus(&base, clusterJobs), true
}

// Update stores the aggregated job when invoked directly.
func (q *aggregateManualQueue) Update(job *refresh.ManualRefreshJob) {
	if job == nil {
		return
	}
	q.mu.Lock()
	if agg, ok := q.jobs[job.ID]; ok {
		agg.job = job
	}
	q.mu.Unlock()
}

// Next blocks until the context is cancelled because aggregation is API-facing only.
func (q *aggregateManualQueue) Next(ctx context.Context) (*refresh.ManualRefreshJob, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (q *aggregateManualQueue) resolveTarget(
	domain string,
	clusterIDs []string,
	queues map[string]refresh.ManualQueue,
) (string, error) {
	if len(clusterIDs) == 0 {
		return "", fmt.Errorf("cluster scope is required for domain %s", domain)
	}
	if len(clusterIDs) > 1 {
		return "", fmt.Errorf("domain %s requires a single cluster scope (requested: %v)", domain, clusterIDs)
	}

	target := clusterIDs[0]
	if _, ok := queues[target]; !ok {
		return "", fmt.Errorf("cluster %s not active", target)
	}
	return target, nil
}

func buildAggregateStatus(
	base *refresh.ManualRefreshJob,
	clusterJobs map[string]aggregateManualChildJob,
) *refresh.ManualRefreshJob {
	status := aggregateChildStatus{}
	for clusterID, child := range clusterJobs {
		status.recordChild(clusterID, child)
	}

	base.State = status.state()
	base.Error = status.firstErr
	base.LatestVersion = status.maxVersion
	base.StartedAt = status.startedAt
	base.FinishedAt = status.finishedAt
	return base
}

type aggregateChildStatus struct {
	hasQueued    bool
	hasRunning   bool
	hasFailed    bool
	hasCancelled bool
	firstErr     string
	maxVersion   uint64
	startedAt    int64
	finishedAt   int64
}

func (s *aggregateChildStatus) recordChild(clusterID string, child aggregateManualChildJob) {
	if child.queue == nil {
		s.recordMissing(clusterID)
		return
	}
	job, ok := child.queue.Status(child.jobID)
	if !ok || job == nil {
		s.recordMissing(clusterID)
		return
	}
	s.recordState(job.State)
	s.recordJobTimes(job)
	if job.Error != "" && s.firstErr == "" {
		s.firstErr = fmt.Sprintf("%s: %s", clusterID, job.Error)
	}
}

func (s *aggregateChildStatus) recordMissing(clusterID string) {
	s.hasFailed = true
	if s.firstErr == "" {
		s.firstErr = fmt.Sprintf("cluster %s job missing", clusterID)
	}
}

func (s *aggregateChildStatus) recordState(state refresh.JobState) {
	switch state {
	case refresh.JobStateQueued:
		s.hasQueued = true
	case refresh.JobStateRunning:
		s.hasRunning = true
	case refresh.JobStateFailed:
		s.hasFailed = true
	case refresh.JobStateCancelled:
		s.hasCancelled = true
	}
}

func (s *aggregateChildStatus) recordJobTimes(job *refresh.ManualRefreshJob) {
	if job.LatestVersion > s.maxVersion {
		s.maxVersion = job.LatestVersion
	}
	if job.StartedAt > 0 && (s.startedAt == 0 || job.StartedAt < s.startedAt) {
		s.startedAt = job.StartedAt
	}
	if job.FinishedAt > s.finishedAt {
		s.finishedAt = job.FinishedAt
	}
}

func (s aggregateChildStatus) state() refresh.JobState {
	switch {
	case s.hasFailed:
		return refresh.JobStateFailed
	case s.hasCancelled:
		return refresh.JobStateCancelled
	case s.hasRunning:
		return refresh.JobStateRunning
	case s.hasQueued:
		return refresh.JobStateQueued
	default:
		return refresh.JobStateSucceeded
	}
}

func (q *aggregateManualQueue) snapshotConfig() map[string]refresh.ManualQueue {
	q.configMu.RLock()
	defer q.configMu.RUnlock()
	queues := make(map[string]refresh.ManualQueue, len(q.queues))
	for id, queue := range q.queues {
		queues[id] = queue
	}
	return queues
}

// UpdateConfig refreshes the aggregate manual queue wiring after selection changes.
func (q *aggregateManualQueue) UpdateConfig(clusterOrder []string, subsystems map[string]*system.Subsystem) {
	if q == nil {
		return
	}
	next := newAggregateManualQueue(clusterOrder, subsystems)
	q.configMu.Lock()
	q.clusterOrder = next.clusterOrder
	q.queues = next.queues
	q.configMu.Unlock()

	q.moveUnfinishedJobs(next.queues)
}

// moveUnfinishedJobs preserves refresh intent across a subsystem re-warm. The
// old manager has stopped consuming its queue, so an unfinished child must be
// re-enqueued on the replacement manager instead of remaining queued forever.
func (q *aggregateManualQueue) moveUnfinishedJobs(queues map[string]refresh.ManualQueue) {
	migrations := q.unfinishedJobMigrations(queues)
	for _, migration := range migrations {
		q.moveUnfinishedJob(migration)
	}
}

func (q *aggregateManualQueue) unfinishedJobMigrations(queues map[string]refresh.ManualQueue) []aggregateManualJobMigration {
	q.mu.RLock()
	defer q.mu.RUnlock()
	migrations := make([]aggregateManualJobMigration, 0)
	for aggregateJobID, aggregateJob := range q.jobs {
		if aggregateJob == nil || aggregateJob.job == nil {
			continue
		}
		migrations = append(migrations, migrationsForAggregateJob(aggregateJobID, aggregateJob, queues)...)
	}
	return migrations
}

func migrationsForAggregateJob(
	aggregateJobID string,
	aggregateJob *aggregateManualJob,
	queues map[string]refresh.ManualQueue,
) []aggregateManualJobMigration {
	result := make([]aggregateManualJobMigration, 0)
	_, scopeValue := refresh.SplitClusterScopeList(aggregateJob.job.Scope)
	for clusterID, child := range aggregateJob.clusterJobs {
		replacement := queues[clusterID]
		if !manualChildNeedsMigration(child, replacement) {
			continue
		}
		result = append(result, aggregateManualJobMigration{
			aggregateJobID: aggregateJobID,
			clusterID:      clusterID,
			domain:         aggregateJob.job.Domain,
			scope:          refresh.JoinClusterScope(clusterID, scopeValue),
			reason:         aggregateJob.job.Reason,
			previous:       child,
			replacement:    replacement,
		})
	}
	return result
}

func manualChildNeedsMigration(child aggregateManualChildJob, replacement refresh.ManualQueue) bool {
	if replacement == nil || child.queue == replacement {
		return false
	}
	status, ok := child.queue.Status(child.jobID)
	if !ok || status == nil {
		return true
	}
	return status.State == refresh.JobStateQueued || status.State == refresh.JobStateRunning
}

func (q *aggregateManualQueue) moveUnfinishedJob(migration aggregateManualJobMigration) {
	job, err := migration.replacement.Enqueue(
		context.Background(),
		migration.domain,
		migration.scope,
		migration.reason,
	)
	if err != nil || job == nil {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	aggregateJob := q.jobs[migration.aggregateJobID]
	if aggregateJob == nil {
		return
	}
	current, ok := aggregateJob.clusterJobs[migration.clusterID]
	if ok && current.jobID == migration.previous.jobID && current.queue == migration.previous.queue {
		aggregateJob.clusterJobs[migration.clusterID] = aggregateManualChildJob{
			jobID: job.ID,
			queue: migration.replacement,
		}
	}
}

// generateAggregateJobID returns a unique identifier for aggregate manual refresh jobs.
func generateAggregateJobID() string {
	return fmt.Sprintf("job-agg-%d", time.Now().UnixNano())
}
