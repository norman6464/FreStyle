import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDebounce } from '../useDebounce';

describe('useDebounce', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('初期値をそのまま返す', () => {
    const { result } = renderHook(() => useDebounce('初期値', 300));
    expect(result.current).toBe('初期値');
  });

  it('指定ミリ秒後にデバウンスされた値が更新される', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'A', delay: 300 } }
    );

    rerender({ value: 'B', delay: 300 });

    // まだ更新されていない
    expect(result.current).toBe('A');

    act(() => { vi.advanceTimersByTime(300); });

    expect(result.current).toBe('B');
  });

  it('遅延中に値が変わるとタイマーがリセットされる', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'A', delay: 300 } }
    );

    rerender({ value: 'B', delay: 300 });
    act(() => { vi.advanceTimersByTime(200); });

    rerender({ value: 'C', delay: 300 });
    act(() => { vi.advanceTimersByTime(200); });

    // Bにはならない
    expect(result.current).toBe('A');

    act(() => { vi.advanceTimersByTime(100); });

    expect(result.current).toBe('C');
  });

  it('異なるdelayで動作する', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'A', delay: 500 } }
    );

    rerender({ value: 'B', delay: 500 });

    act(() => { vi.advanceTimersByTime(300); });
    expect(result.current).toBe('A');

    act(() => { vi.advanceTimersByTime(200); });
    expect(result.current).toBe('B');
  });

  it('数値型でも動作する', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 0, delay: 300 } }
    );

    rerender({ value: 42, delay: 300 });

    act(() => { vi.advanceTimersByTime(300); });

    expect(result.current).toBe(42);
  });
});
