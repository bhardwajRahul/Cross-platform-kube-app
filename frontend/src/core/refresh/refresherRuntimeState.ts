export type RefreshInvocation = 'automatic' | 'foreground' | 'manual';
export type RefresherTimer = number | ReturnType<typeof globalThis.setTimeout>;

export type RefreshExecutionSummary = {
  successCount: number;
  failures: Array<{ error: Error; timedOut: boolean }>;
};

export type RefresherExecution = {
  id: number;
  controller: AbortController;
  promise: Promise<RefreshExecutionSummary>;
  invocation: RefreshInvocation;
};

export type RefresherIntentState =
  | { status: 'disabled' }
  | { status: 'paused' }
  | { status: 'enabled' };

export type RefresherTimingState =
  | { status: 'idle'; intervalTimer?: RefresherTimer }
  | { status: 'cooldown'; cooldownTimer: RefresherTimer; intervalTimer?: RefresherTimer };

export type RefresherExecutionState =
  | { status: 'idle' }
  | ({ status: 'running' } & RefresherExecution);

export type RefresherRuntimeState = {
  intent: RefresherIntentState;
  timing: RefresherTimingState;
  execution: RefresherExecutionState;
};

export type RefresherRuntimeEvent =
  | { type: 'enabled' }
  | { type: 'disabled' }
  | { type: 'paused' }
  | { type: 'idle'; intervalTimer?: RefresherTimer }
  | { type: 'interval-replaced'; intervalTimer: RefresherTimer }
  | { type: 'refresh-started'; execution: RefresherExecution }
  | { type: 'refresh-finished'; executionId: number }
  | { type: 'cooldown-started'; cooldownTimer: RefresherTimer }
  | { type: 'cooldown-finished'; cooldownTimer: RefresherTimer };

export const initialRefresherRuntimeState = (enabled: boolean): RefresherRuntimeState => ({
  intent: enabled ? { status: 'enabled' } : { status: 'disabled' },
  timing: { status: 'idle' },
  execution: { status: 'idle' },
});

const intervalTimerFor = (timing: RefresherTimingState): RefresherTimer | undefined =>
  timing.intervalTimer;

export const transitionRefresherRuntimeState = (
  state: RefresherRuntimeState,
  event: RefresherRuntimeEvent
): RefresherRuntimeState => {
  switch (event.type) {
    case 'enabled':
      return state.intent.status !== 'enabled'
        ? { ...state, intent: { status: 'enabled' } }
        : state;
    case 'disabled':
      return state.intent.status === 'disabled' &&
        state.timing.status === 'idle' &&
        state.execution.status === 'idle'
        ? state
        : {
            intent: { status: 'disabled' },
            timing: { status: 'idle' },
            execution: { status: 'idle' },
          };
    case 'paused':
      return state.intent.status === 'disabled' || state.intent.status === 'paused'
        ? state
        : { ...state, intent: { status: 'paused' }, timing: { status: 'idle' } };
    case 'idle':
      return {
        ...state,
        timing: {
          status: 'idle',
          ...(event.intervalTimer !== undefined ? { intervalTimer: event.intervalTimer } : {}),
        },
      };
    case 'interval-replaced':
      return {
        ...state,
        timing: { ...state.timing, intervalTimer: event.intervalTimer },
      };
    case 'refresh-started':
      return {
        ...state,
        execution: { status: 'running', ...event.execution },
      };
    case 'refresh-finished':
      return state.execution.status !== 'running' || state.execution.id !== event.executionId
        ? state
        : { ...state, execution: { status: 'idle' } };
    case 'cooldown-started':
      return {
        ...state,
        timing: {
          status: 'cooldown',
          cooldownTimer: event.cooldownTimer,
          ...(intervalTimerFor(state.timing) !== undefined
            ? { intervalTimer: intervalTimerFor(state.timing) }
            : {}),
        },
      };
    case 'cooldown-finished':
      if (
        state.timing.status !== 'cooldown' ||
        state.timing.cooldownTimer !== event.cooldownTimer
      ) {
        return state;
      }
      return {
        ...state,
        timing: {
          status: 'idle',
          ...(state.timing.intervalTimer !== undefined
            ? { intervalTimer: state.timing.intervalTimer }
            : {}),
        },
      };
  }
};

export const isRefresherEnabled = (state: RefresherRuntimeState): boolean =>
  state.intent.status !== 'disabled';

export const refresherIntervalTimer = (state: RefresherRuntimeState): RefresherTimer | undefined =>
  intervalTimerFor(state.timing);

export const refresherCooldownTimer = (state: RefresherRuntimeState): RefresherTimer | undefined =>
  state.timing.status === 'cooldown' ? state.timing.cooldownTimer : undefined;
