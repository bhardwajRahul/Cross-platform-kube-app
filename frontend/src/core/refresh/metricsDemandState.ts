export type MetricsDemandState =
  | {
      status: 'idle';
      appliedKey: string;
      retryDelayMs: number;
    }
  | {
      status: 'requesting';
      appliedKey: string;
      requestKey: string;
      retryDelayMs: number;
    }
  | {
      status: 'waiting-retry';
      appliedKey: string;
      retryKey: string;
      retryTimer: ReturnType<typeof setTimeout>;
      retryDelayMs: number;
    };

export type MetricsDemandEvent =
  | { type: 'request-started'; key: string }
  | { type: 'request-succeeded'; key: string; initialDelayMs?: number }
  | { type: 'request-abandoned'; key: string; initialDelayMs?: number }
  | {
      type: 'retry-scheduled';
      key: string;
      timer: ReturnType<typeof setTimeout>;
      maxDelayMs: number;
    }
  | { type: 'retry-fired'; key: string }
  | { type: 'retry-cancelled' };

export const initialMetricsDemandState = (initialDelayMs = 1_000): MetricsDemandState => ({
  status: 'idle',
  appliedKey: '',
  retryDelayMs: initialDelayMs,
});

export const transitionMetricsDemandState = (
  state: MetricsDemandState,
  event: MetricsDemandEvent
): MetricsDemandState => {
  switch (event.type) {
    case 'request-started':
      return state.status === 'idle'
        ? { ...state, status: 'requesting', requestKey: event.key }
        : state;
    case 'request-succeeded':
      if (state.status !== 'requesting' || state.requestKey !== event.key) {
        return state;
      }
      return {
        status: 'idle',
        appliedKey: event.key,
        retryDelayMs: event.initialDelayMs ?? 1_000,
      };
    case 'request-abandoned':
      if (state.status !== 'requesting' || state.requestKey !== event.key) {
        return state;
      }
      return {
        status: 'idle',
        appliedKey: state.appliedKey,
        retryDelayMs: event.initialDelayMs ?? 1_000,
      };
    case 'retry-scheduled':
      if (state.status !== 'requesting' || state.requestKey !== event.key) {
        return state;
      }
      return {
        status: 'waiting-retry',
        appliedKey: state.appliedKey,
        retryKey: event.key,
        retryTimer: event.timer,
        retryDelayMs: Math.min(state.retryDelayMs * 2, event.maxDelayMs),
      };
    case 'retry-fired':
      if (state.status !== 'waiting-retry' || state.retryKey !== event.key) {
        return state;
      }
      return {
        status: 'idle',
        appliedKey: state.appliedKey,
        retryDelayMs: state.retryDelayMs,
      };
    case 'retry-cancelled':
      return state.status === 'waiting-retry'
        ? {
            status: 'idle',
            appliedKey: state.appliedKey,
            retryDelayMs: state.retryDelayMs,
          }
        : state;
  }
};
