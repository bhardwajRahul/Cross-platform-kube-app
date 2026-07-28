export type ClusterSelectionPhase = 'pending' | 'empty' | 'active';

interface ClusterSelectionPhaseInput {
  hasSelectedClusters: boolean;
  kubeconfigsLoading: boolean;
}

export const getClusterSelectionPhase = ({
  hasSelectedClusters,
  kubeconfigsLoading,
}: ClusterSelectionPhaseInput): ClusterSelectionPhase => {
  if (hasSelectedClusters) {
    return 'active';
  }
  return kubeconfigsLoading ? 'pending' : 'empty';
};
