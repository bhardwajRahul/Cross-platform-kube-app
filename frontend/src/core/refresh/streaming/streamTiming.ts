const UINT32_SPAN = 2 ** 32;

// Reconnect jitter exists to keep clients from retrying in lockstep, so it is
// drawn from the platform CSPRNG rather than Math.random: the sequence is not
// predictable from observing one client's retry timing. Returns [0, 1), the
// same range Math.random provides.
const unitRandom = (): number => {
  const buffer = new Uint32Array(1);
  crypto.getRandomValues(buffer);
  return buffer[0] / UINT32_SPAN;
};

export const streamReconnectDelay = (
  attempt: number,
  options: {
    baseMs?: number;
    maxMs?: number;
    minMs?: number;
    jitterMs?: number;
    jitterFactor?: number;
    round?: boolean;
  } = {}
): number => {
  const baseMs = options.baseMs ?? 1000;
  const maxMs = options.maxMs ?? 30_000;
  const minMs = options.minMs ?? 0;
  const backoff = Math.min(maxMs, baseMs * 2 ** attempt);
  const absoluteJitter = options.jitterMs ? unitRandom() * options.jitterMs : 0;
  const proportionalJitter = options.jitterFactor
    ? backoff * ((unitRandom() * 2 - 1) * options.jitterFactor)
    : 0;
  const delay = Math.max(minMs, backoff + absoluteJitter + proportionalJitter);
  return options.round ? Math.round(delay) : delay;
};
