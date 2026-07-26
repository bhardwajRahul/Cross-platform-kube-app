/** Normalizes the single- or multi-select value emitted by Dropdown. */
export const normalizeDropdownValue = (value: string | string[]): string[] => {
  if (Array.isArray(value)) {
    return value;
  }
  if (value) {
    return [value];
  }
  return [];
};
