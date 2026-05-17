# Operation Lifecycle Registry Plan

Created: 2026-05-17
Status: Temporary implementation plan

## Overview

Several user-facing workflows create live or long-running cluster-scoped work:

- Shell exec sessions and debug-container attach flows.
- Port-forward sessions with reconnect behavior.
- Node drain and maintenance jobs.

Each workflow has its own backend store, event names, frontend status handling,
and cluster cleanup logic. The target model is a small backend operation
lifecycle registry that gives cluster-scoped runtime work one cleanup and
visibility contract while preserving workflow-specific implementation details.

The desired end state is:

- Every live operation records cluster identity, operation type, stable ID,
  status, timestamps, and cleanup behavior.
- Cluster removal, kubeconfig clearing, auth/client teardown, and app shutdown
  call one backend cleanup path.
- Frontend status panels consume consistent runtime-read snapshots and events.
- Workflow-specific state, such as shell backlog, port-forward local ports, and
  drain events, remains owned by the workflow implementation.

## Non-Goals

- Do not collapse shell, port-forward, and node drain implementation into one
  generic runner.
- Do not close shell sessions when an object panel unmounts; current session
  continuity is intentional.
- Do not remove workflow-specific event streams such as shell output or
  port-forward status.
- Do not change Kubernetes operation behavior unless lifecycle tests prove a
  current bug.

## Inventory

Architecture and workflow docs:

- `docs/workflows/shell-debug.md`
- `docs/workflows/logs/overview.md`
- `docs/architecture/data-access.md`
- `.agents/skills/operations-workflows/SKILL.md`

Backend shell surfaces:

- `backend/shell_sessions.go`
- `backend/shell_sessions_test.go`
- `backend/shell_sessions_error_test.go`
- `backend/resources/pods/debug.go`

Backend port-forward surfaces:

- `backend/portforward.go`
- `backend/portforward_types.go`
- `backend/portforward_test.go`

Backend node-maintenance surfaces:

- `backend/nodemaintenance`
- `backend/resources_nodes.go`
- `backend/refresh/snapshot/node_maintenance.go`
- `backend/resource_permission_test.go`

Cluster lifecycle cleanup callers:

- `backend/cluster_clients.go`
- `backend/kubeconfigs.go`
- `frontend/src/ui/layout/ClusterTabs.tsx`

Frontend runtime/session surfaces:

- `frontend/src/core/app-state-access/readers.ts`
- `frontend/src/modules/object-panel/components/ObjectPanel/Shell/ShellTab.tsx`
- `frontend/src/modules/port-forward`
- `frontend/src/ui/status/SessionsStatus.tsx`
- `frontend/src/shared/components/modals/DrainNodeModal.tsx`
- `frontend/src/shared/hooks/useNodeMaintenanceActions.tsx`

Current broadness:

- Backend cluster teardown calls shell and port-forward cleanup in at least
  three places.
- Frontend cluster tab close also coordinates shell and port-forward cleanup.
- Runtime session list reads are centralized through `appStateAccess`, but
  mutating operations and event contracts remain workflow-specific.
- `docs/workflows/shell-debug.md` references session-panel paths that are not
  present in the current frontend tree.

## Phases

- [ ] Phase 1: Lifecycle contract tests
  - Add backend tests proving cluster removal stops shell sessions, port
    forwards, and active node-maintenance work through one expected cleanup
    contract.
  - Add tests for idempotent cleanup when the same cluster is removed through
    multiple paths.
  - Add frontend tests that status surfaces remove sessions after list events
    and do not keep stale entries for the active cluster.

- [ ] Phase 2: Backend registry shape
  - Introduce a small backend registry for cluster-scoped operations with:
    operation type, ID, cluster ID, optional namespace/name, status, started
    time, and cleanup callback.
  - Keep the registry separate from refresh state and app settings.
  - Register shell sessions, port-forward sessions, and drain jobs without
    changing their workflow-specific stores yet.

- [ ] Phase 3: Single backend cleanup entrypoint
  - Add a single backend cleanup function for removed clusters.
  - Replace repeated shell and port-forward cleanup calls in kubeconfig/client
    lifecycle code with the new entrypoint.
  - Decide whether cluster-tab close should call only selection change and let
    backend lifecycle cleanup handle runtime operations, or whether it should
    call an explicit backend "close cluster" command.

- [ ] Phase 4: Unified runtime snapshots
  - Add a runtime operation list snapshot for status surfaces, or add parity
    tests that prove existing shell and port-forward list snapshots remain
    consistent with the registry.
  - Keep shell backlog and port-forward local port updates on their existing
    workflow APIs.
  - Keep drain details on `object-maintenance`, but connect active job presence
    to the registry.

- [ ] Phase 5: Frontend status simplification
  - Update `SessionsStatus` and port-forward panels to use the final runtime
    snapshot/list contract.
  - Keep object-panel shell reattach behavior intact.
  - Keep active cluster filtering based on `clusterId`.
  - Add tests for cross-cluster shell jump, active cluster switch, and removed
    cluster cleanup.

- [ ] Phase 6: Documentation and skill update
  - Update `docs/workflows/shell-debug.md` so referenced frontend paths match
    the current implementation.
  - Add an operation lifecycle section to the relevant workflow docs.
  - Update `.agents/skills/operations-workflows/SKILL.md` with the registry and
    cleanup checklist.

## Open Questions

- Should node drain cancellation be part of cluster cleanup, or should completed
  and historical drain jobs remain visible until their bounded history expires?
- Should the registry emit a single `runtime-operations:list` event, or should
  it only enforce backend cleanup while existing workflow events stay public?
- Should cluster-tab close become a backend command that can atomically stop
  operations and update selected kubeconfigs?
- How much historical status should survive after an operation is cleaned up?

## Validation Plan

- `go test ./backend ./backend/resources/pods ./backend/resources/nodes ./backend/nodemaintenance`
- `npm run test --prefix frontend -- Shell port-forward SessionsStatus drain`
- `npm run typecheck --prefix frontend`
- Browser validation for shell/port-forward status behavior if frontend status
  surfaces change.
- `mage qc:prerelease`
- Inspect `git status --short` after the final gate because lint/fix steps may
  modify files.

## Progress Notes

- 2026-05-17: Plan created from app-review findings. No implementation started.
