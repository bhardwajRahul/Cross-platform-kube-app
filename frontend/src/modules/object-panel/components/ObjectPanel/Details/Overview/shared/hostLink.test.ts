/**
 * frontend/src/modules/object-panel/components/ObjectPanel/Details/Overview/shared/hostLink.test.ts
 *
 * Test suite for hostLink — URL building and scheme inference for clickable
 * Ingress/Gateway hostnames.
 */

import { describe, expect, it } from 'vitest';

import { buildHostUrl, ingressHostSchemes, listenerScheme } from './hostLink';

describe('buildHostUrl', () => {
  it('defaults to https and omits the default port', () => {
    expect(buildHostUrl({ host: 'example.com' })).toBe('https://example.com');
    expect(buildHostUrl({ host: 'example.com', scheme: 'https', port: 443 })).toBe(
      'https://example.com'
    );
  });

  it('uses http when requested and omits port 80', () => {
    expect(buildHostUrl({ host: 'example.com', scheme: 'http' })).toBe('http://example.com');
    expect(buildHostUrl({ host: 'example.com', scheme: 'http', port: 80 })).toBe(
      'http://example.com'
    );
  });

  it('includes a non-default port', () => {
    expect(buildHostUrl({ host: 'example.com', scheme: 'http', port: 8080 })).toBe(
      'http://example.com:8080'
    );
    expect(buildHostUrl({ host: 'example.com', scheme: 'https', port: 8443 })).toBe(
      'https://example.com:8443'
    );
  });

  it('returns null for wildcard hosts (not a concrete address)', () => {
    expect(buildHostUrl({ host: '*.example.com' })).toBeNull();
  });

  it('returns null for empty or whitespace hosts', () => {
    expect(buildHostUrl({ host: '' })).toBeNull();
    expect(buildHostUrl({ host: '   ' })).toBeNull();
  });

  it('trims surrounding whitespace', () => {
    expect(buildHostUrl({ host: '  example.com  ' })).toBe('https://example.com');
  });
});

describe('ingressHostSchemes', () => {
  it('offers https and http for a TLS-covered host', () => {
    expect(ingressHostSchemes('example.com', ['example.com'])).toEqual(['https', 'http']);
  });

  it('offers https and http when covered by a single-label wildcard TLS host', () => {
    expect(ingressHostSchemes('app.example.com', ['*.example.com'])).toEqual(['https', 'http']);
  });

  it('offers only http for a host not covered by TLS', () => {
    expect(ingressHostSchemes('example.com', [])).toEqual(['http']);
    expect(ingressHostSchemes('other.com', ['example.com'])).toEqual(['http']);
    // Wildcard only matches a single leading label, not multi-level subdomains.
    expect(ingressHostSchemes('a.b.example.com', ['*.example.com'])).toEqual(['http']);
  });
});

describe('listenerScheme', () => {
  it('maps HTTP and HTTPS protocols to a scheme', () => {
    expect(listenerScheme('HTTP')).toBe('http');
    expect(listenerScheme('HTTPS')).toBe('https');
    expect(listenerScheme('https')).toBe('https');
  });

  it('returns null for protocols without an unambiguous web scheme', () => {
    expect(listenerScheme('TLS')).toBeNull();
    expect(listenerScheme('TCP')).toBeNull();
    expect(listenerScheme('UDP')).toBeNull();
    expect(listenerScheme(undefined)).toBeNull();
    expect(listenerScheme('')).toBeNull();
  });
});
