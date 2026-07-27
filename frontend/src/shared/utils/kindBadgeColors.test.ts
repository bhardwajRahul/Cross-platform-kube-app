import { describe, expect, it } from 'vitest';
import { getKindColorClass } from './kindBadgeColors';

describe('getKindColorClass', () => {
  it('hashes Unicode kind names by code point', () => {
    expect(getKindColorClass('Widget😀')).toBe('hash-color-24');
  });
});
