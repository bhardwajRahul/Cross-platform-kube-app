package snapshot

import (
	"errors"
	"strings"

	"github.com/luxury-yacht/app/backend/refresh"
)

const namespaceSnapshotScopePrefix = "namespace:"

// NamespaceSnapshotScope is the parsed identity for namespace-scoped snapshot
// builders. Namespace is empty only when AllNamespaces is true.
type NamespaceSnapshotScope struct {
	ClusterID      string
	Namespace      string
	AllNamespaces  bool
	CanonicalScope string
}

func parseNamespaceSnapshotScope(scope, requiredMessage string) (NamespaceSnapshotScope, error) {
	clusterID, trimmed := refresh.SplitClusterScope(scope)
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return NamespaceSnapshotScope{}, errors.New(requiredMessage)
	}

	if isAllNamespaceScope(trimmed) {
		return NamespaceSnapshotScope{
			ClusterID:      clusterID,
			AllNamespaces:  true,
			CanonicalScope: refresh.JoinClusterScope(clusterID, namespaceSnapshotScopePrefix+"all"),
		}, nil
	}

	namespace, err := parseNamespaceScopeValue(trimmed, requiredMessage)
	if err != nil {
		return NamespaceSnapshotScope{}, err
	}
	return NamespaceSnapshotScope{
		ClusterID:      clusterID,
		Namespace:      namespace,
		CanonicalScope: refresh.JoinClusterScope(clusterID, namespaceSnapshotScopePrefix+namespace),
	}, nil
}

func parseNamespaceScopeValue(scope, requiredMessage string) (string, error) {
	_, scopeValue := refresh.SplitClusterScope(scope)
	namespace := strings.TrimSpace(scopeValue)
	if strings.HasPrefix(namespace, namespaceSnapshotScopePrefix) {
		namespace = strings.TrimPrefix(namespace, namespaceSnapshotScopePrefix)
		namespace = strings.TrimLeft(namespace, ":")
	}
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return "", errors.New(requiredMessage)
	}
	return namespace, nil
}

func isAllNamespaceScope(scope string) bool {
	_, scopeValue := refresh.SplitClusterScope(scope)
	value := strings.TrimSpace(strings.ToLower(scopeValue))
	if value == "" {
		return false
	}
	if strings.HasPrefix(value, namespaceSnapshotScopePrefix) {
		value = strings.TrimLeft(strings.TrimPrefix(value, namespaceSnapshotScopePrefix), ":")
	}
	switch value {
	case "all", "*":
		return true
	default:
		return false
	}
}
