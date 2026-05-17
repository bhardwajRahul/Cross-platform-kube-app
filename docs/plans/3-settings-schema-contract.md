# Settings Schema Contract Plan

Created: 2026-05-17
Status: Temporary implementation plan

## Overview

Backend settings already expose schema metadata through
`GetAppSettingsSchema()`: key, type, default, current value, min, max, enum
values, validation format, and runtime side-effect flags. The frontend consumes
that schema during hydration, but it still carries many duplicated defaults,
min/max bounds, normalization rules, and input attributes.

The target model is backend-owned settings metadata with frontend controls built
from the schema wherever a preference is backend-owned. The frontend should keep
only minimal fallback constants needed before Wails is available or when schema
loading fails.

The desired end state is:

- Backend settings schema is the source of truth for persisted/runtime
  preference defaults, bounds, enum values, validation hints, and runtime
  side-effect flags.
- Frontend settings controls and `appPreferences` normalization use schema
  metadata instead of duplicated lockstep constants.
- Preference updates still go through `UpdateAppPreferences` with optimistic
  rollback.
- Local-only settings and bootstrap caches remain frontend-owned.

## Non-Goals

- Do not move transient UI state, such as the active Settings tab, into backend
  settings.
- Do not remove localStorage appearance bootstrap caches; they are needed before
  Wails reads are available.
- Do not replace `UpdateAppPreferences` with per-setting setters in new code.
- Do not change user-visible defaults or bounds unless a current mismatch is
  proven and intentionally corrected.

## Inventory

Architecture and guidance:

- `docs/architecture/data-access.md`
- `.agents/skills/app-shell/SKILL.md`
- `frontend/AGENTS.md`

Backend settings surfaces:

- `backend/app_settings.go`
- `backend/app_settings_test.go`
- Wails models generated from backend settings DTOs

Frontend settings state:

- `frontend/src/core/settings/appPreferences.ts`
- `frontend/src/core/app-state-access/readers.ts`
- `frontend/src/core/settings/appPreferences.test.ts`

Frontend Settings UI:

- `frontend/src/ui/settings`
- `frontend/src/ui/settings/sections/AdvancedSection.tsx`
- `frontend/src/ui/settings/sections/ObjectPanelSection.tsx`
- `frontend/src/ui/settings/sections/AppearanceSection.tsx`
- `frontend/src/ui/settings/sections/KubeconfigsSection.tsx`

Current broadness:

- `appPreferences.ts` has fallback defaults, min/max constants, schema
  hydration, normalization, event emission, and persistence logic in one module.
- Advanced settings inputs duplicate min/max/default behavior even after schema
  hydration.
- The backend schema currently describes more than 30 preferences.

## Phases

- [ ] Phase 1: Schema coverage tests
  - Add backend tests that every backend-owned `AppSettings` preference appears
    in `buildAppSettingsSchema`.
  - Add frontend tests that every schema entry used by `appPreferences` has a
    typed accessor and update path.
  - Add a test that frontend fallback constants match backend schema values only
    for the explicitly allowed bootstrap fallback keys.

- [ ] Phase 2: Preference metadata access layer
  - Add a small frontend schema metadata helper that returns typed default,
    current value, min, max, enum values, validation format, and runtime flag by
    key.
  - Keep schema fetch through `appStateAccess`.
  - Make failed schema loads explicit in diagnostics or logged error handling
    without breaking first paint.

- [ ] Phase 3: Normalize from schema
  - Replace duplicated numeric normalization in `appPreferences.ts` with
    schema-driven integer, enum, boolean, string, and color normalization
    helpers.
  - Keep special migration behavior where needed, such as zero-as-unset for old
    settings files.
  - Keep derived caches for appearance bootstrap localStorage, but document that
    they are not a source of truth.

- [ ] Phase 4: Schema-backed settings controls
  - Update Advanced settings controls to read `min`, `max`, `default`, and
    current value from schema metadata.
  - Update Object Panel layout controls to use schema metadata for persisted
    dimensions and positions.
  - Keep frontend-only controls outside schema-driven helpers.

- [ ] Phase 5: Persistence and rollback hardening
  - Ensure all backend-owned settings use `UpdateAppPreferences`.
  - Add tests for rollback after persistence failure, including emitted events
    and appearance bootstrap cache restoration.
  - Add tests for runtime-effect preferences so the frontend does not assume
    success before backend persistence completes.

- [ ] Phase 6: Documentation and skill update
  - Update `docs/architecture/data-access.md` Settings Contract with the final
    schema-driven frontend shape.
  - Update `.agents/skills/app-shell/SKILL.md` with the new checklist for
    adding settings.
  - Remove obsolete "keep in lockstep" comments once tests enforce the contract.

## Open Questions

- Which fallback constants are truly needed before Wails is available, and which
  can be removed entirely?
- Should schema metadata be exposed to Settings components directly, or only
  through typed hooks in `core/settings`?
- Should `UpdateAppPreferences` return schema metadata with the updated
  settings to avoid a separate schema refresh?
- Should runtime-effect flags drive UI copy or only validation and diagnostics?

## Validation Plan

- `go test ./backend`
- `npm run test --prefix frontend -- settings appPreferences`
- `npm run typecheck --prefix frontend`
- Browser or Storybook validation for Settings controls if UI behavior changes.
- `mage qc:prerelease`
- Inspect `git status --short` after the final gate because lint/fix steps may
  modify files.

## Progress Notes

- 2026-05-17: Plan created from app-review findings. No implementation started.
