import { describe, it, expect } from 'vitest';
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

    act(() => scrollTo(el, 150)); // 上へ
    expect(result.current.hidden).toBe(false);
  });

  it('先頭付近（topThreshold 以内）では常に表示', () => {
    const el = makeScrollable();
    const { result } = renderHook(() => useAutoHideOnScroll({ topThreshold: 80, delta: 8 }));
    act(() => result.current.scrollRef(el));
    act(() => scrollTo(el, 200));
    expect(result.current.hidden).toBe(true);
    act(() => scrollTo(el, 40)); // 先頭付近へ
    expect(result.current.hidden).toBe(false);
  });

  it('delta 未満の小刻みな揺れでは状態を変えない', () => {
    const el = makeScrollable();
    const { result } = renderHook(() => useAutoHideOnScroll({ topThreshold: 80, delta: 8 }));
    act(() => result.current.scrollRef(el));
    act(() => scrollTo(el, 200));
    expect(result.current.hidden).toBe(true);
    act(() => scrollTo(el, 196)); // 4px 上（delta 未満）
    expect(result.current.hidden).toBe(true);
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
