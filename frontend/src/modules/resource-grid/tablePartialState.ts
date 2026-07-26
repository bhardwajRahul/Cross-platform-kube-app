import type { SnapshotStats } from '@/core/refresh/client';

interface LocalPartialLabelOptions {
  stats?: SnapshotStats | null;
  fallback: string;
  sourceLabel?: string;
  sourceVerb?: 'is' | 'are';
}

const cleanWarnings = (stats?: SnapshotStats | null): string[] =>
  (stats?.warnings ?? []).map((warning) => warning.trim()).filter(Boolean);

export const buildLocalPartialDataLabel = ({
  stats,
  fallback,
  sourceLabel = 'This table',
  sourceVerb = 'is',
}: LocalPartialLabelOptions): string => {
  const warnings = cleanWarnings(stats);
  let windowLabel: string;

  if (warnings.length > 0) {
    windowLabel = warnings.join(' ');
  } else if (stats?.truncated && stats.totalItems && stats.totalItems > stats.itemCount) {
    windowLabel = `Showing ${stats.itemCount} of ${stats.totalItems} rows.`;
  } else {
    windowLabel = fallback;
  }

  return `${windowLabel} ${sourceLabel} ${sourceVerb} a bounded local window. Search, filters, sort, copy, and actions apply only to the visible rows.`;
};
