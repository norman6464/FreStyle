import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useSmoothReveal } from '../useSmoothReveal';

// vitest の fake timers は Date も偽装するため、EMA(到着間隔)と放出タイマーの両方を
// advanceTimersByTime で決定的に検証できる。
describe('useSmoothReveal', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  function setup(content: string, active: boolean) {
    return renderHook(({ c, a }: { c: string; a: boolean }) => useSmoothReveal(c, a), {
      initialProps: { c: content, a: active },
    });
  }

  it('mount 時は全文を即時表示する(履歴・確定メッセージ・remount)', () => {
    const { result } = setup('こんにちは。今日は晴れです。', false);
    expect(result.current.text).toBe('こんにちは。今日は晴れです。');
    expect(result.current.settled).toBe(true);
  });

  it('ストリーミング中、最初のチャンクは待たずに即時放出される', () => {
    const { result, rerender } = setup('', true);
    expect(result.current.text).toBe('');
    act(() => {
      rerender({ c: 'こんにちは、', a: true });
    });
    // 「考え中」からの最初のチャンクは fb を待たない。
    expect(result.current.text).toBe('こんにちは、');
  });

  it('後続のチャンクは適応間隔のタイマーで 1 つずつ放出される', () => {
    const { result, rerender } = setup('', true);
    act(() => {
      rerender({ c: 'こんにちは、', a: true });
    });
    act(() => {
      vi.advanceTimersByTime(100);
      rerender({ c: 'こんにちは、今日は晴れ。明日は雨。', a: true });
    });
    // まだタイマー前なので前回までの表示のまま。
    expect(result.current.text).toBe('こんにちは、');
    // fb = (EMA + 600) / (残チャンク数 + 1)。十分進めると 1 チャンクずつ増える。
    act(() => {
      vi.advanceTimersByTime(400);
    });
    expect(result.current.text).toBe('こんにちは、今日は晴れ。');
    act(() => {
      vi.advanceTimersByTime(600);
    });
    expect(result.current.text).toBe('こんにちは、今日は晴れ。明日は雨。');
    expect(result.current.settled).toBe(false); // まだ active
  });

  it('句読点で終わらない未完チャンクはストリーミング中は保留される', () => {
    const { result, rerender } = setup('', true);
    act(() => {
      rerender({ c: 'こんにちは、', a: true });
    });
    act(() => {
      vi.advanceTimersByTime(100);
      rerender({ c: 'こんにちは、続きがまだ途中', a: true });
    });
    act(() => {
      vi.advanceTimersByTime(5000);
    });
    // 「続きがまだ途中」は境界(句読点/改行)で終わっていないため出さない。
    expect(result.current.text).toBe('こんにちは、');
  });

  it('完了(active false)で保留分も含めて流し切り settled になる', () => {
    const { result, rerender } = setup('', true);
    act(() => {
      rerender({ c: 'こんにちは、', a: true });
    });
    act(() => {
      vi.advanceTimersByTime(100);
      rerender({ c: 'こんにちは、続きがまだ途中', a: true });
    });
    act(() => {
      rerender({ c: 'こんにちは、続きがまだ途中', a: false });
    });
    expect(result.current.settled).toBe(false);
    act(() => {
      vi.advanceTimersByTime(5000);
    });
    expect(result.current.text).toBe('こんにちは、続きがまだ途中');
    expect(result.current.settled).toBe(true);
  });

  it('最初のチャンク放出前に active=false + 句読点なしの content が来たら即時に全表示する(エラー文)', () => {
    // useAskAi の error 修正後の状態遷移を再現: visible='' active=true から、
    // 句読点を含まないエラー文で active=false になる。 未完チャンク保留は active 中のみなので、
    // 完了状態では丸ごと放出される(「考え中」固定にならない)。
    const { result, rerender } = setup('', true);
    expect(result.current.text).toBe('');
    act(() => {
      rerender({ c: 'ネットワークエラーが発生しました', a: false });
    });
    expect(result.current.text).toBe('ネットワークエラーが発生しました');
    expect(result.current.settled).toBe(true);
  });

  it('表示済み prefix の延長でない content(エラー置換)は即スナップする', () => {
    const { result, rerender } = setup('', true);
    act(() => {
      rerender({ c: '回答の途中、', a: true });
    });
    expect(result.current.text).toBe('回答の途中、');
    act(() => {
      rerender({ c: '（エラー）ネットワークエラーが発生しました', a: true });
    });
    expect(result.current.text).toBe('（エラー）ネットワークエラーが発生しました');
  });

  it('長い無放出期間の後に大量の backlog が来たら catch-up でまとめて出す', () => {
    const { result, rerender } = setup('', true);
    act(() => {
      rerender({ c: '一。', a: true });
    });
    expect(result.current.text).toBe('一。');
    // 前回の放出から 10 秒経過してから大量のチャンクが届く(接続の停滞など)。
    act(() => {
      vi.advanceTimersByTime(10_000);
      rerender({ c: '一。二。三。四。五。六。七。八。', a: true });
    });
    // 最初の tick で elapsed が fb×10 を超えているため、1 個ずつではなく一括で追いつく。
    act(() => {
      vi.advanceTimersByTime(300);
    });
    expect(result.current.text).toBe('一。二。三。四。五。六。七。八。');
  });

  it('unmount で進行中のタイマーが破棄される', () => {
    const { rerender, unmount } = setup('', true);
    act(() => {
      rerender({ c: 'こんにちは、', a: true });
    });
    act(() => {
      vi.advanceTimersByTime(100);
      rerender({ c: 'こんにちは、今日は晴れ。', a: true });
    });
    expect(vi.getTimerCount()).toBeGreaterThan(0);
    unmount();
    expect(vi.getTimerCount()).toBe(0);
  });
});
