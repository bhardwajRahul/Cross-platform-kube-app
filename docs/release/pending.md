### Fixed

- Permission denials, expired credentials, and missing Kubernetes objects no longer fire Sentry error reports.
- Canceled permission checks no longer produce bursts of duplicate error reports after a cluster connection fails.
