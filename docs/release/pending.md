### Added

**Global views**

- When more than one cluster is open, a new Global tab will be displayed with views that synthesize data from all open clusters.
  - **Clusters** view shows status, summary info, and aggregated metrics for all open clusters.
  - **Namespaces** view shows status, summary info, and metrics per-namespace in all open clusters.

**Attention view**

- Clusters now have an Attention view that shows items in that cluster that may require your attention.
  - Attention findings can be ignored individually, per cluster, or globally (right-click on an item to ignore it).
  - The sidebar shows summary badges for info/warning/error counts.

**Combined Workloads and Pods into a single view**
- The Workloads view now includes pods in a split-pane view. Selecting a workload in the top pane filters which pods are visible in the bottom pane.
  - The Pods pane can be collapsed to allow more screen space for workloads.

**Filter chips**

- When filters are enabled in any table view, those filters are displayed as chips to easily see which filters are being applied to the view.
  - This brings table filters to parity with the filter chips that already exist for container logs.

### Changed

- Overhauled the Favorites system to properly support all views, including the new split-pane workloads view.
  - **⚠️ The new Favorites system required a schema change.** The app will attempt to migrate your existing favorites automatically. If it cannot migrate them, it will delete them. If this happens, you'll have to manually recreate them. I tested the migration but it's possible that I missed something, so apologies in advance if this happens to you.
- Improved container log scrolling. When you manually scroll in the logs, it will now properly retain your place in the viewport, continue to buffer new logs in the background, and present a button to resume scrolling.
- Pagination controls do not appear unless there are more than 25 rows in the table (25 is the smallest pagination size).
- Added Type, Reason, and Source filters to Events views.
- Pod Status and Pod Signal in the Cluster Overview now link to the new Attention view, with the proper filters applied.
- Standardized absent table values as a dimmed hyphen across all table views.
- Every multi-select filter is now searchable and has Select All/Select None controls.

### Fixed

- Dropdowns are now viewport-aware and zoom-aware so should correctly render in all window sizes and zoom levels.
- Fixed YAML editor and and object map data freshness. The app is now aware of when versionless snapshot payloads change, so deleted or updated objects no longer remain visible from a stale response from the backend.
- Many improvements to the refresh system (https://github.com/luxury-yacht/app/pull/266) to improve the app's performance and stability.
