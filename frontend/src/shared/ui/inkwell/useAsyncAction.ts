import { useCallback, useEffect, useReducer } from 'react';

export type AsyncStatus = 'idle' | 'loading' | 'success' | 'error';

type Event = { type: 'SUBMIT' } | { type: 'RESOLVE' } | { type: 'REJECT' } | { type: 'RESET' };

// 相互排他な状態機械。idle 以外からの SUBMIT は無視され二重送信できない。
function reducer(status: AsyncStatus, event: Event): AsyncStatus {
  switch (status) {
    case 'idle':
      return event.type === 'SUBMIT' ? 'loading' : status;
    case 'loading':
      return event.type === 'RESOLVE' ? 'success' : event.type === 'REJECT' ? 'error' : status;
    case 'success':
    case 'error':
      return event.type === 'RESET' ? 'idle' : status;
    default:
      return status;
  }
}

interface Options {
  /** success を表示してから idle に戻すまでの ms。 */
  successResetMs?: number;
  /** error を表示してから idle に戻すまでの ms。 */
  errorResetMs?: number;
}

/**
 * 非同期アクションを idle → loading →（success | error）→ idle の状態機械で扱う。
 * loading 中の再実行は状態機械が弾く（二重送信防止）。success/error は自動で idle へ戻す。
 */
export function useAsyncAction(action: () => Promise<void>, options: Options = {}) {
  const { successResetMs = 1200, errorResetMs = 2000 } = options;
  const [status, dispatch] = useReducer(reducer, 'idle');

  const run = useCallback(async () => {
    if (status !== 'idle') return; // 副作用も含めて再入を防ぐ
    dispatch({ type: 'SUBMIT' });
    try {
      await action();
      dispatch({ type: 'RESOLVE' });
    } catch {
      dispatch({ type: 'REJECT' });
    }
  }, [status, action]);

  useEffect(() => {
    if (status === 'success') {
      const t = setTimeout(() => dispatch({ type: 'RESET' }), successResetMs);
      return () => clearTimeout(t);
    }
    if (status === 'error') {
      const t = setTimeout(() => dispatch({ type: 'RESET' }), errorResetMs);
      return () => clearTimeout(t);
    }
  }, [status, successResetMs, errorResetMs]);

  return { status, run };
}
