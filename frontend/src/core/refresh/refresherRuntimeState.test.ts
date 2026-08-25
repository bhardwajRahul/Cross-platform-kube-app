import { describe, expect, it } from 'vitest';

import {
  initialRefresherRuntimeState,
  transitionRefresherRuntimeState,
} from './refresherRuntimeState';

describe('transitionRefresherRuntimeState', () => {
  it('keeps pause intent and ignores a stale execution completion', () => {
    const first = {
      id: 1,
      controller: new AbortController(),
      promise: Promise.resolve({ successCount: 0, failures: [] }),
      invocation: 'automatic' as const,
    };
    const replacement = { ...first, id: 2, controller: new AbortController() };
    let state = transitionRefresherRuntimeState(initialRefresherRuntimeState(true), {
      type: 'refresh-started',
      execution: first,
    });
    state = transitionRefresherRuntimeState(state, { type: 'paused' });
    state = transitionRefresherRuntimeState(state, {
      type: 'refresh-started',
      execution: replacement,
    });

    const afterStaleCompletion = transitionRefresherRuntimeState(state, {
      type: 'refresh-finished',
      executionId: first.id,
    });
    expect(afterStaleCompletion).toBe(state);
    expect(afterStaleCompletion.intent).toEqual({ status: 'paused' });
    expect(afterStaleCompletion.execution).toEqual({ status: 'running', ...replacement });
  });
});
