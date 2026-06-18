### Added

- Error notifications now have a "Copy error" button that copies the full error — message, technical details, suggestions, and context — to the clipboard for easy pasting into bug reports.
- Ingress rule hosts and Gateway listener hostnames in the object panel now show small `https`/`http` links that open the host in your browser. Each valid scheme is offered explicitly — a TLS-covered Ingress host gets both `https` and `http`, a plaintext host gets `http`, and each Gateway listener gets its own scheme/port — so the scheme is always visible. Wildcard hosts stay as plain text.

### Fixed

- Tables now load as soon as their own data is ready during cluster connection: one slow or failing watch (for example a misbehaving CRD or restricted resource) no longer delays every other view's first load.
- Switching the active cluster no longer briefly shows the previous cluster's rows in a table before the new cluster's data loads.
