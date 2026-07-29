### Changed

- Clicking empty space in a table de-selects highlighted row. In the Workloads table, this also clears the filters on the Pods table that are set by selecting a row.

### Fixed

- The "no active clusters" overlay no longer appears for a moment while saved cluster selections are loading.
- State files can no longer accidentally be wiped out by other instances of the app.
- Editing or applying object YAML now fails with a clear error when it is invoked without an operation context, instead of continuing on a detached background context.
- Background permission-cache refreshes are always time-bounded and are no longer cancelled when the request that triggered them finishes.
- CSV exports are written with owner-only permissions instead of being readable by every account on the machine (macOS and Linux only, no change on Windows).
- Stream reconnect jitter is drawn from the platform cryptographic RNG, so reconnect timing cannot be predicted by observing a single client.

### Developer Notes

- Switched to `mise` for toolchain management.
- Added sonarcloud properties file.
