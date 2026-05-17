# Permission And Capability Contract Plan

Created: 2026-05-17
Status: Temporary implementation plan

## Overview

Luxury Yacht has two intentionally separate permission systems:

- Backend refresh-domain permissions decide which snapshot and stream domains can
  list or watch Kubernetes resources.
- UI permission and capability checks decide which actions are shown, enabled,
  denied, or explained in object menus and workflow controls.

Both systems are valid, but the contract is currently spread across backend
domain registrations, backend runtime permission checks, frontend permission
specs, capability descriptors, diagnostics feature labels, and tests. The target
model is a single permission contract layer that keeps those independent runtime
systems aligned without merging their caches or request paths.

The desired end state is:

- Resource and action permission descriptors live in one reviewed catalog or are
  parity-checked from one catalog.
- Refresh-domain permission requirements, UI permission specs, diagnostics
  feature labels, and action/capability descriptors cannot drift silently.
- Restricted-RBAC behavior remains visible through permission-denied refresh
  domains, disabled action reasons, and diagnostics.
- CRD and built-in GVK identity stays explicit through every permission and
  capability boundary.

## Non-Goals

- Do not merge the refresh subsystem SSAR checker with the UI permission SSRR
  store. They have different runtime jobs and cache behavior.
- Do not change Kubernetes RBAC semantics or make optimistic authorization
  decisions in the frontend.
- Do not replace action-specific backend mutation checks with frontend gating.
  Backend write paths must keep enforcing permissions.
- Do not add a broad resource-model rewrite unless parity checks prove the
  current descriptor shape cannot express the contract.

## Inventory

Architecture and guidance:

- `docs/architecture/permissions.md`
- `.agents/skills/permissions-capabilities/SKILL.md`
- `.agents/context/code-map.md`

Backend refresh permission surfaces:

- `backend/refresh/snapshot/permission_checks.go`
- `backend/refresh/snapshot/permission.go`
- `backend/refresh/snapshot/service.go`
- `backend/refresh/system/registrations.go`
- `backend/refresh/system/permission_gate.go`
- `backend/refresh/permissions/checker.go`

Backend UI/action permission surfaces:

- `backend/capabilities`
- `backend/app_permissions.go`
- `backend/resource_permission.go`
- Backend mutation paths under `backend/resources`, `backend/object_yaml*.go`,
  `backend/portforward*.go`, and `backend/shell_sessions.go`

Frontend permission and diagnostics surfaces:

- `frontend/src/core/capabilities/permissionSpecs.ts`
- `frontend/src/core/capabilities/permissionStore.ts`
- `frontend/src/core/capabilities/hooks.ts`
- `frontend/src/core/capabilities/catalog.ts`
- `frontend/src/core/refresh/components/diagnostics/diagnosticsPanelConfig.ts`
- `frontend/src/modules/object-panel/components/ObjectPanel/hooks/useObjectPanelCapabilities.ts`
- `frontend/src/shared/hooks/useObjectActions.tsx`
- `frontend/src/shared/components/kubernetes/ActionsMenu.tsx`

Current broadness:

- `backend/refresh/system/registrations.go` has 32 domain registration entries.
- `frontend/src/core/capabilities/permissionSpecs.ts` has 18 feature groups.
- Diagnostics feature filtering is manually keyed by feature strings.

## Phases

- [ ] Phase 1: Contract inventory and mismatch tests
  - Add focused tests that enumerate refresh-domain permission requirements,
    frontend permission feature labels, capability catalog labels, and
    diagnostics feature filters.
  - Fail on missing diagnostics mappings for any permission feature that should
    be visible.
  - Fail on built-in permission specs whose kind cannot resolve to a known
    group/version.
  - Add backend tests that confirm every registered refresh domain has an
    explicit runtime permission policy or an intentional exemption.

- [ ] Phase 2: Descriptor normalization
  - Introduce shared backend helper types for refresh permission resource
    requirements so `permission_checks.go` and `registrations.go` do not repeat
    group/resource lists in incompatible formats.
  - Preserve `requireAll` and `requireAny` behavior.
  - Keep partial-data permission structs in each snapshot builder, but generate
    or derive their allowed-resource keys from the same normalized resource
    descriptors.

- [ ] Phase 3: Frontend feature-label centralization
  - Move permission feature labels and diagnostics grouping keys into a small
    typed catalog.
  - Update `permissionSpecs.ts`, diagnostics feature maps, and capability catalog
    references to consume the typed labels.
  - Keep view-specific filtering behavior unchanged, including browse showing
    all scoped permissions.

- [ ] Phase 4: Action and capability parity
  - Add parity tests for object action descriptors, object-panel capabilities,
    node maintenance actions, shell/debug controls, port-forward actions, and
    backend mutation permission checks.
  - Ensure every UI-visible mutating action has a matching backend permission
    check and a denied/pending UI reason.
  - Ensure CRD action checks always carry explicit group and version.

- [ ] Phase 5: Restricted-RBAC diagnostics hardening
  - Add or update restricted-RBAC tests that prove denied domains remain visible
    in refresh diagnostics.
  - Add frontend tests showing Effective Permissions and Capabilities Checks
    remain populated for the active cluster and view.
  - Verify partial-data domains still show available resources when some
    resources are denied.

- [ ] Phase 6: Documentation and skill update
  - Update `docs/architecture/permissions.md` to describe the new catalog or
    parity-check contract.
  - Update `.agents/skills/permissions-capabilities/SKILL.md` with the new
    add/change checklist.
  - Remove obsolete warnings that are replaced by enforced tests.

## Open Questions

- Should the canonical permission descriptor catalog live in backend Go only,
  frontend TypeScript only, or as generated JSON consumed by both sides?
- Should diagnostics feature labels be user-facing strings, stable enum keys
  with display labels, or both?
- Which action families should be in the first parity pass: all current actions,
  or the highest-risk mutating workflows first?
- How should CRD-specific lazy permissions be represented in a static parity
  test without requiring live discovery?

## Validation Plan

- `go test ./backend/capabilities ./backend/refresh/snapshot ./backend/refresh/system ./backend`
- `npm run test --prefix frontend -- capabilities ObjectPanel ActionsMenu diagnostics`
- `npm run typecheck --prefix frontend`
- `mage qc:prerelease`
- Inspect `git status --short` after the final gate because lint/fix steps may
  modify files.

## Progress Notes

- 2026-05-17: Plan created from app-review findings. No implementation started.
