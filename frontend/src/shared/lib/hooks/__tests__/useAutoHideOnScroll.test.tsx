import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useAutoHideOnScroll } from '../useAutoHideOnScroll';

function makeScrollable(): HTMLElement {
  const el = document.createElement('div');
  Object.defineProperty(el, 'scrollTop', { value: 0, writable: true });
  return el;
}

function scrollTo(el: HTMLElement, top: number) {
  (el as unknown as { scrollTop: number }).scrollTop = top;
  el.dispatchEvent(new Event('scroll'));
}

describe('useAutoHideOnScroll', () => {
  beforeEach(() => {
    vi.useFakeTimers({ now: 0 });
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  /** advance は cooldown（Date.now 基準）を経過させる。 */
  const advance = (ms: number) => vi.setSystemTime(Date.now() + ms);

  it('初期状態は表示（hidden=false）', () => {
    const { result } = renderHook(() => useAutoHideOnScroll());
    expect(result.current.hidden).toBe(false);
  });

  it('しきい値を超えて下へスクロールすると hidden になり、上へ戻すと表示に戻る', () => {
    const el = makeScrollable();
    const { result } = renderHook(() => useAutoHideOnScroll({ topThreshold: 80, delta: 8 }));
    act(() => result.current.scrollRef(el));

    act(() => scrollTo(el, 200)); // 下へ大きく
    expect(result.current.hidden).toBe(true);

    advance(400); // 切替クールダウンを経過させる
    act(() => scrollTo(el, 150)); // 上へ
    expect(result.current.hidden).toBe(false);
  });

  it('先頭付近（topThreshold 以内）では常に表示', () => {
    const el = makeScrollable();
    const { result } = renderHook(() => useAutoHideOnScroll({ topThreshold: 80, delta: 8 }));
    act(() => result.current.scrollRef(el));
    act(() => scrollTo(el, 200));
    expect(result.current.hidden).toBe(true);
    act(() => scrollTo(el, 40)); // 先頭付近へ（クールダウン中でも先頭では必ず表示）
    expect(result.current.hidden).toBe(false);
  });

  it('delta 未満の小刻みな揺れでは状態を変えない', () => {
    const el = makeScrollable();
    const { result } = renderHook(() => useAutoHideOnScroll({ topThreshold: 80, delta: 8 }));
    act(() => result.current.scrollRef(el));
    act(() => scrollTo(el, 200));
    expect(result.current.hidden).toBe(true);
    advance(400);
    act(() => scrollTo(el, 196)); // 4px 上（delta 未満）
    expect(result.current.hidden).toBe(true);
  });

  it('最下部のクランプ（隠す→表示域が伸びて scrollTop が押し戻される）で振動しない', () => {
    const el = makeScrollable();
    const { result } = renderHook(() => useAutoHideOnScroll({ topThreshold: 80, delta: 8, cooldownMs: 350 }));
    act(() => result.current.scrollRef(el));

    // 最下部まで下スクロール → hidden。直後にヘッダー分（64px）クランプされた偽イベントが来る。
    act(() => scrollTo(el, 1000));
    expect(result.current.hidden).toBe(true);
    act(() => scrollTo(el, 936)); // ブラウザによる自動クランプ（-64px の偽「上スクロール」）
    // クールダウン中は方向判定しない → 隠れたまま（振動しない）。
    expect(result.current.hidden).toBe(true);

    // クールダウン後、ユーザーが本当に上へスクロールすれば表示に戻る。
    advance(400);
    act(() => scrollTo(el, 900));
    expect(result.current.hidden).toBe(false);
  });

  it('監視対象が消えたら（ノート切替のローディング等）表示へ戻る', () => {
    const el = makeScrollable();
    const { result } = renderHook(() => useAutoHideOnScroll());
    act(() => result.current.scrollRef(el));
    act(() => scrollTo(el, 300));
    expect(result.current.hidden).toBe(true);
    act(() => result.current.scrollRef(null));
    expect(result.current.hidden).toBe(false);
  });
});
