/**
 * Compile-time exhaustiveness guard.
 *
 * Call in the `default` branch of a switch (or the end of an if/else chain) over
 * a discriminated union. When every variant is handled, `value` narrows to
 * `never` and this compiles; if a new variant is added and left unhandled, the
 * argument is no longer `never` and the build fails — making the unhandled state
 * impossible to ship. It also throws at runtime if reached via an unchecked
 * boundary.
 */
export function assertNever(value: never, context = 'value'): never {
  throw new Error(`Unexpected ${context}: ${JSON.stringify(value)}`);
}
