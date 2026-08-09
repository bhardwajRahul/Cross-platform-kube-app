### Changed

- The app no longer throws an error when it cannot find a valid kubeconfig file. It now shows an informative warning with a link to the Settings panel to update the list of kubeconfig directories.

### Fixed

- When a Kubernetes watch is denied due to the user's permissions, this is the app behaving as designed and is no longer reported as an application error.
- Expected Kubernetes authentication and connectivity failures no longer create Sentry issues, while operational exceptions, unrecognized failures, and client-side deadline overruns remain reportable.
- Fixed a regression that would prevent clusters from recovering after expired authentication was refreshed. For example, an expired SSO token would cause an expected authentication warning. Refreshing the SSO token would cause the app to start to load the cluster, but it would never complete.
