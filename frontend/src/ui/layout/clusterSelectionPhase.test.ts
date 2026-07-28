import { describe, expect, it } from 'vitest';
import { getClusterSelectionPhase } from './clusterSelectionPhase';

describe('getClusterSelectionPhase', () => {
  it('distinguishes pending hydration from settled empty and active selections', () => {
    expect(getClusterSelectionPhase({ hasSelectedClusters: false, kubeconfigsLoading: true })).toBe(
      'pending'
    );
    expect(
      getClusterSelectionPhase({ hasSelectedClusters: false, kubeconfigsLoading: false })
    ).toBe('empty');
    expect(getClusterSelectionPhase({ hasSelectedClusters: true, kubeconfigsLoading: true })).toBe(
      'active'
    );
    expect(getClusterSelectionPhase({ hasSelectedClusters: true, kubeconfigsLoading: false })).toBe(
      'active'
    );
  });
});
