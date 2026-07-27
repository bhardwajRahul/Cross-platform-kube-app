import { describe, expect, it } from 'vitest';
import { parseBracketedLogPrefix } from './logLineMetadata';

describe('parseBracketedLogPrefix', () => {
  it('separates a bracketed label, complete prefix, and remainder', () => {
    expect(parseBracketedLogPrefix('[api-7]   ready')).toEqual({
      label: 'api-7',
      prefix: '[api-7]   ',
      remainder: 'ready',
    });
  });

  it('rejects missing and empty bracketed labels', () => {
    expect(parseBracketedLogPrefix('api-7 ready')).toBeNull();
    expect(parseBracketedLogPrefix('[] ready')).toBeNull();
    expect(parseBracketedLogPrefix('[api-7 ready')).toBeNull();
  });
});
