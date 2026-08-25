I’d prioritize these refactors in this order. They preserve the current refresh semantics—cluster scoping, retained snapshots, signal-driven refetch, and polling fallback—while reducing lifecycle and state-management risk.

1. **Unify refresh-generation activation and teardown** — highest value, medium risk

   **Status (2026-08-24): implemented.** `backend/refresh_generation.go` now owns generation startup, pre-publication rollback, permission-revalidation ownership, manager lifetime, and reverse-order stop. Initial setup, selection updates, auth/governor rebuilds, cooling, cluster removal, and global teardown use that shared lifecycle.

   Before this refactor, initial startup started permission revalidation but the auth/governor rebuild path did not. Initial setup and selection updates now begin the shared activation before publication ([refresh_setup.go](/Volumes/git/luxury-yacht/app/backend/refresh_setup.go:50), [refresh_update.go](/Volumes/git/luxury-yacht/app/backend/refresh_update.go:166)); auth/governor rebuilds enter the same contract before routing replacement ([cluster_auth_contract.go](/Volumes/git/luxury-yacht/app/backend/cluster_auth_contract.go:100)).

   The transactional activation owns manager startup, store reconciliation, permission revalidation, and rollback until the caller publishes routing; commit transfers exact-generation cancellation ownership to the coordinator ([refresh_generation.go](/Volumes/git/luxury-yacht/app/backend/refresh_generation.go:39), [refresh_generation.go](/Volumes/git/luxury-yacht/app/backend/refresh_generation.go:93)). Every teardown path uses the reverse-order generation stop ([refresh_generation.go](/Volumes/git/luxury-yacht/app/backend/refresh_generation.go:248)).

2. **Turn the authored domain contract into executable policy** — high value, low-to-medium risk

   **Status (2026-08-24): implemented.** The authored JSON now includes frontend registration order and scheduling alongside its existing cache, source-clock, orchestration, timing, and priority metadata ([refresh-domain-contract.json](/Volumes/git/luxury-yacht/app/backend/refresh/domain/refresh-domain-contract.json:743)). The generator validates the policy vocabulary, uniqueness, and ordering before emitting typed Go and TypeScript tables ([domain_contract.go](/Volumes/git/luxury-yacht/app/backend/internal/genrefreshcontracts/domain_contract.go:119), [policy_render.go](/Volumes/git/luxury-yacht/app/backend/internal/genrefreshcontracts/policy_render.go:17), [render.go](/Volumes/git/luxury-yacht/app/backend/internal/genrefreshcontracts/render.go:97)).

   Backend registration callbacks are keyed by domain and ordered by the generated policy, while snapshot-cache bypass reads the generated cache policy ([registrations.go](/Volumes/git/luxury-yacht/app/backend/refresh/system/registrations.go:291), [service.go](/Volumes/git/luxury-yacht/app/backend/refresh/snapshot/service.go:477)). Frontend orchestration callbacks are keyed by generated orchestrator kind; generated policy drives registration order, scheduling, descriptor timing, and metric demand ([domainRegistrations.ts](/Volumes/git/luxury-yacht/app/frontend/src/core/refresh/domainRegistrations.ts:45), [domainRegistry.ts](/Volumes/git/luxury-yacht/app/frontend/src/core/refresh/domainRegistry.ts:182), [orchestrator.ts](/Volumes/git/luxury-yacht/app/frontend/src/core/refresh/orchestrator.ts:1653)).

3. **Make snapshot singleflight independent of the first caller’s cancellation** — high value, contained change

   **Status (2026-08-24): implemented.** `snapshotBuildFlights` gives each caller
   an independent wait context while reference-counting the shared build. A
   canceled leader no longer cancels a live follower; the build stops when every
   waiter leaves ([service_flights.go](/Volumes/git/luxury-yacht/app/backend/refresh/snapshot/service_flights.go:12), [service.go](/Volumes/git/luxury-yacht/app/backend/refresh/snapshot/service.go:164)).

   The snapshot service wraps the full permission/readiness/build path in a
   rotatable generation cancellation epoch. Generation teardown cancels current
   flights before producers stop, while later requests keep the same service
   usable for governor-cooled retained reads ([service.go](/Volumes/git/luxury-yacht/app/backend/refresh/snapshot/service.go:189), [manager.go](/Volumes/git/luxury-yacht/app/backend/refresh/system/manager.go:535), [refresh_generation.go](/Volumes/git/luxury-yacht/app/backend/refresh_generation.go:257)). Regression tests cover leader/follower cancellation, all-waiter cancellation, readiness cancellation, generation reuse, cache-bypass isolation, and permission-specific cache keys.

4. **Model frontend refresh as one state machine per `(clusterId, domain, scope)`** — very high simplification, medium-to-high migration risk

   **Status (2026-08-24): in progress; fetch lifecycle seam implemented.**
   `ScopedFetchState` is now a discriminated `idle`/`fetching` state reduced by
   explicit fetch-started, stream-signal, settled, and canceled events. Request
   ownership and the one-trailing-signal latch no longer depend on an optional
   mutable flag, and stale settle/cancel events cannot clear a replacement
   request ([refreshRuntime.ts](/Volumes/git/luxury-yacht/app/frontend/src/core/refresh/refreshRuntime.ts:17), [refreshRuntime.test.ts](/Volumes/git/luxury-yacht/app/frontend/src/core/refresh/refreshRuntime.test.ts:9)).

   `ClusterRefreshRuntime` keeps the existing orchestrator boundary while
   routing start, latch, settle, teardown, and iteration through that reducer;
   the orchestrator's request scheduling and trailing refetch behavior stay on
   their existing APIs ([refreshRuntime.ts](/Volumes/git/luxury-yacht/app/frontend/src/core/refresh/refreshRuntime.ts:321), [orchestrator.ts](/Volumes/git/luxury-yacht/app/frontend/src/core/refresh/orchestrator.ts:1374)).

   Remaining seams are `RefreshManager` timing intent, lease/scope enablement,
   activation/readiness, stream lifecycle/health, permission/auth state, and
   metric demand. Migrate them one at a time behind current APIs, preserving
   navigation, remount, multi-cluster, retained-state, and polling behavior.

   Before this migration, `RefreshManager` owned timer, promise, abort, enabled,
   and context-transition state while the orchestrator separately owned
   readiness, leases, auth failures, metric demand, and scope enablement;
   `ClusterRefreshRuntime` independently tracked in-flight requests.

5. **Normalize resource-stream messages through a pure protocol reducer** — high debugging payoff, medium risk

   Each subscription currently carries numerous independently mutable epochs, errors, queue, timer, reset, and resync fields ([resourceStreamSubscriptions.ts](/Volumes/git/luxury-yacht/app/frontend/src/core/refresh/streaming/resourceStreamSubscriptions.ts:10)). Signal and legacy frames follow separate transition paths ([resourceStreamManager.ts](/Volumes/git/luxury-yacht/app/frontend/src/core/refresh/streaming/resourceStreamManager.ts:399), [resourceStreamManager.ts](/Volumes/git/luxury-yacht/app/frontend/src/core/refresh/streaming/resourceStreamManager.ts:469)), while health is inferred afterward from combinations of those fields ([resourceStreamManager.ts](/Volumes/git/luxury-yacht/app/frontend/src/core/refresh/streaming/resourceStreamManager.ts:701)). One concrete complexity symptom: resync clears `updateQueue`, then immediately derives source versions from that cleared queue ([resourceStreamManager.ts](/Volumes/git/luxury-yacht/app/frontend/src/core/refresh/streaming/resourceStreamManager.ts:1057)).

   Normalize both wire formats into canonical events and reduce them into explicit states such as connecting, awaiting acknowledgement, synchronized, resyncing, permission-blocked, and stopping. Keep network, timer, and store writes as emitted effects. Lock behavior with transition-table tests for replay, reset, quiet subscriptions, permission denial, overflow, replacement, and reconnect.

Audit validation baseline (before implementation): focused backend refresh tests passed; focused frontend refresh/streaming tests passed `43` files and `529` tests; targeted rebuild/governor tests passed. No files were changed, and final `git status --short` was empty.
