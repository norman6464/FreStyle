import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { usePracticeTimer } from '../usePracticeTimer';

describe('usePracticeTimer', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('初期状態は停止中で0秒', () => {
    const { result } = renderHook(() => usePracticeTimer());

    expect(result.current.seconds).toBe(0);
    expect(result.current.isRunning).toBe(false);
    expect(result.current.formattedTime).toBe('00:00');
  });

  it('startで計測開始し、秒数が増加する', () => {
    const { result } = renderHook(() => usePracticeTimer());

    act(() => {
      result.current.start();
    });

    expect(result.current.isRunning).toBe(true);

    act(() => {
      vi.advanceTimersByTime(3000);
    });

    expect(result.current.seconds).toBe(3);
  });

  it('stopで計測停止する', () => {
    const { result } = renderHook(() => usePracticeTimer());

    act(() => {
      result.current.start();
    });

    act(() => {
      vi.advanceTimersByTime(5000);
    });

    act(() => {
      result.current.stop();
    });

    expect(result.current.isRunning).toBe(false);
    expect(result.current.seconds).toBe(5);

    act(() => {
      vi.advanceTimersByTime(3000);
    });

    expect(result.current.seconds).toBe(5);
  });

  it('resetでタイマーをリセットする', () => {
    const { result } = renderHook(() => usePracticeTimer());

    act(() => {
      result.current.start();
    });

    act(() => {
      vi.advanceTimersByTime(10000);
    });

    act(() => {
      result.current.reset();
    });

    expect(result.current.seconds).toBe(0);
    expect(result.current.isRunning).toBe(false);
  });

  it('formattedTimeが正しいフォーマットで返される', () => {
    const { result } = renderHook(() => usePracticeTimer());

    act(() => {
      result.current.start();
    });

    act(() => {
      vi.advanceTimersByTime(65000); // 1分5秒
    });

    expect(result.current.formattedTime).toBe('01:05');
  });

  it('二重start防止: 既にrunning中のstartは無視される', () => {
    const { result } = renderHook(() => usePracticeTimer());

    act(() => {
      result.current.start();
    });

    act(() => {
      vi.advanceTimersByTime(3000);
    });

    // 二重start
    act(() => {
      result.current.start();
    });

    act(() => {
      vi.advanceTimersByTime(2000);
    });

    // 3 + 2 = 5秒（二重startで倍速にならない）
    expect(result.current.seconds).toBe(5);
  });

  it('stop後のstartで秒数が維持されたまま再開する', () => {
    const { result } = renderHook(() => usePracticeTimer());

    act(() => {
      result.current.start();
    });

    act(() => {
      vi.advanceTimersByTime(5000);
    });

    act(() => {
      result.current.stop();
    });

    expect(result.current.seconds).toBe(5);

    act(() => {
      result.current.start();
    });

    act(() => {
      vi.advanceTimersByTime(3000);
    });

    expect(result.current.seconds).toBe(8);
  });

  it('10分超のformattedTimeが正しく表示される', () => {
    const { result } = renderHook(() => usePracticeTimer());

    act(() => {
      result.current.start();
    });

    act(() => {
      vi.advanceTimersByTime(754000); // 12分34秒
    });

    expect(result.current.formattedTime).toBe('12:34');
  });
});
