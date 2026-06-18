/**
 * frontend/src/modules/object-panel/components/ObjectPanel/Details/Overview/shared/hostLink.ts
 *
 * URL building and scheme inference for clickable Ingress/Gateway hostnames.
 * Kept free of React so the logic can be unit-tested in isolation.
 */

export type UrlScheme = 'http' | 'https';

/** A wildcard host (e.g. *.example.com) is not a concrete, browsable address. */
const isWildcardHost = (host: string): boolean => host.includes('*');

const defaultPortFor = (scheme: UrlScheme): number => (scheme === 'https' ? 443 : 80);

interface BuildHostUrlOptions {
  host: string;
  /** Defaults to https when the scheme can't be inferred by the caller. */
  scheme?: UrlScheme;
  port?: number;
}

/**
 * Builds a browsable URL for a hostname, or returns null when the host can't be
 * opened (empty or a wildcard). The default port for the scheme is omitted.
 */
export const buildHostUrl = ({
  host,
  scheme = 'https',
  port,
}: BuildHostUrlOptions): string | null => {
  const trimmed = host?.trim();
  if (!trimmed || isWildcardHost(trimmed)) {
    return null;
  }
  const portSuffix = port && port > 0 && port !== defaultPortFor(scheme) ? `:${port}` : '';
  return `${scheme}://${trimmed}${portSuffix}`;
};

/** True when `host` is covered by a TLS host entry (exact or single-label wildcard). */
const hostMatchesTlsEntry = (host: string, entry: string): boolean => {
  if (entry === host) {
    return true;
  }
  if (entry.startsWith('*.')) {
    const suffix = entry.slice(1); // ".example.com"
    if (!host.endsWith(suffix)) {
      return false;
    }
    // A DNS wildcard matches exactly one leading label.
    const label = host.slice(0, host.length - suffix.length);
    return label.length > 0 && !label.includes('.');
  }
  return false;
};

/**
 * Returns the schemes worth offering for an Ingress rule host, in display
 * order. A TLS-covered host (exact or single-label wildcard match) is reachable
 * over https and usually also http (which typically redirects), so both are
 * offered. A host with no TLS coverage is plain HTTP only — https there would
 * fail with no certificate, so it is not offered.
 */
export const ingressHostSchemes = (host: string, tlsHosts: string[]): UrlScheme[] =>
  tlsHosts.some((entry) => hostMatchesTlsEntry(host, entry)) ? ['https', 'http'] : ['http'];

/**
 * Maps a Gateway listener protocol to a web scheme. Only HTTP and HTTPS have an
 * unambiguous browsable scheme; TLS/TCP/UDP return null so the hostname stays
 * plain text rather than producing a link that likely won't resolve.
 */
export const listenerScheme = (protocol: string | undefined): UrlScheme | null => {
  switch (protocol?.toUpperCase()) {
    case 'HTTPS':
      return 'https';
    case 'HTTP':
      return 'http';
    default:
      return null;
  }
};
