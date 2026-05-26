### Changed

- YAML editor now shows protected fields in edit mode, disallows changes to those keys/values, and shows a hover tooltip explaining why they are protected.
- YAML editor extracted to a shared component to reduce code duplication across views that display/edit YAML.
- Object panel YAML toolbar icons now sit together as a single group next to the search controls instead of being split across the header. The `managedFields` toggle stays visible while editing but is disabled, with a tooltip explaining why.

### Fixed

- YAML editing now checks the same `patch` permission that the save path uses and shows the permission denial reason on the disabled edit action. This prevents a user with insufficient permissions from being allowed to enter edit mode, but not being allowed to save changes.
- Editor save verification no longer warns when the live object only differs by protected server-owned fields, generated Deployment/ReplicaSet annotations, or `kubectl.kubernetes.io/last-applied-configuration`. This reduces the amount of unimportant diffs reported by the editor on save.
