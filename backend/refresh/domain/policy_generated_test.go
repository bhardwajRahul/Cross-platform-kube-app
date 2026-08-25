package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratedSnapshotCacheBypassPolicy(t *testing.T) {
	bypassDomains := []string{
		"namespace-metrics",
		"nodes",
		"namespace-workloads",
		"pods",
		"object-details",
		"object-maintenance",
	}
	for _, domainName := range bypassDomains {
		require.Truef(t, BypassesSnapshotCache(domainName), "domain %s", domainName)
	}
	require.False(t, BypassesSnapshotCache("namespaces"))
	require.False(t, BypassesSnapshotCache("unknown-domain"))
}

func TestGeneratedSingleflightBypassPolicy(t *testing.T) {
	require.True(t, BypassesSingleflight("object-maintenance"))
	require.False(t, BypassesSingleflight("namespaces"))
	require.False(t, BypassesSingleflight("unknown-domain"))
}

func TestGeneratedPoliciesReturnDefensiveSourceClockCopies(t *testing.T) {
	policies := Policies()
	require.NotEmpty(t, policies)
	policies[0].SourceClocks[0] = SourceClockMetric

	policy, ok := LookupPolicy(policies[0].Domain)
	require.True(t, ok)
	require.Equal(t, SourceClockObject, policy.SourceClocks[0])
}
