### Added

### Changed

### Fixed

- Object Panel → Details now refreshes when the underlying object changes (e.g. a
  Deployment's container image tag). The details snapshot previously carried a constant
  source-version ETag, so the backend kept replying "not modified" and the panel showed
  stale content for the rest of the session. The object's `resourceVersion` is now the
  details source clock, and the header-metadata cache is evicted on change (which also
  keeps the "Last Modified" field current). Applies to every kind, including custom
  resources.
