import { describe, expect, it } from 'vitest';
import { normalizeDropdownValue } from './dropdownValue';

describe('normalizeDropdownValue', () => {
  it('preserves multi-select values', () => {
    const values = ['one', 'two'];
    expect(normalizeDropdownValue(values)).toBe(values);
  });

  it('wraps a single selected value', () => {
    expect(normalizeDropdownValue('one')).toEqual(['one']);
  });

  it('normalizes an empty single-select value to an empty array', () => {
    expect(normalizeDropdownValue('')).toEqual([]);
  });
});
