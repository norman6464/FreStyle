import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useAutoHideOnScroll } from '../useAutoHideOnScroll';

function makeScrollable({ scrollHeight = 10000, clientHeight = 800 } = {}): HTMLElement {
  const el = document.createElement('div');
  Object.defineProperty(el, 'scrollTop', { value: 0, writable: true });
  Object.defineProperty(el, 'scrollHeight', { value: scrollHeight, writable: true });
  Object.defineProperty(el, 'clientHeight', { value: clientHeight, writable: true });
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

  it('下へスクロールすると hidden になり、上へ累積 showAfterUp 戻したときだけ表示に戻る', () => {
    const el = makeScrollable();
    const { result } = renderHook(() =>
      useAutoHideOnScroll({ topThreshold: 80, delta: 8, showAfterUp: 160 }),
    );
    act(() => result.current.scrollRef(el));

    act(() => scrollTo(el, 600)); // 下へ大きく
    expect(result.current.hidden).toBe(true);

    advance(400); // 切替クールダウンを経過させる
    act(() => scrollTo(el, 550)); // 上へ 50px（累積 50 < 160）: まだ出さない
    expect(result.current.hidden).toBe(true);
    act(() => scrollTo(el, 480)); // さらに 70px（累積 120 < 160）: まだ出さない
    expect(result.current.hidden).toBe(true);
    act(() => scrollTo(el, 430)); // さらに 50px（累積 170 >= 160）: 表示に戻る
    expect(result.current.hidden).toBe(false);
  });

  it('上方向の累積は下へ動くとリセットされる', () => {
    const el = makeScrollable();
    const { result } = renderHook(() =>
      useAutoHideOnScroll({ topThreshold: 80, delta: 8, showAfterUp: 160 }),
    );
    act(() => result.current.scrollRef(el));
    act(() => scrollTo(el, 600));
    expect(result.current.hidden).toBe(true);
    advance(400);
    act(() => scrollTo(el, 500)); // 上へ 100（累積 100）
    act(() => scrollTo(el, 560)); // 下へ 60 → 累積リセット
    advance(400);
    act(() => scrollTo(el, 460)); // 上へ 100（累積 100 < 160）: リセット済みなので出ない
    expect(result.current.hidden).toBe(true);
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

    // クールダウン後、ユーザーが本当に大きく（累積 showAfterUp 以上）上へ戻せば表示に戻る。
    advance(400);
    act(() => scrollTo(el, 700));
    expect(result.current.hidden).toBe(false);
  });

  it('最下部付近（隠すとクランプが起きる余地しか無い位置）では隠さない', () => {
    // scrollHeight 1000 / clientHeight 800 → スクロール余地は 200px。
    const el = makeScrollable({ scrollHeight: 1000, clientHeight: 800 });
    const { result } = renderHook(() => useAutoHideOnScroll({ topThreshold: 80, delta: 8, hideReserve: 88 }));
    act(() => result.current.scrollRef(el));

    // 残り余地 200-150=50px（< 88）: 隠すと最下部に張り付くため隠さない。
    act(() => scrollTo(el, 150));
    expect(result.current.hidden).toBe(false);

    // 余地が十分な文書なら同じ位置でも隠れる（対照）。
    const big = makeScrollable({ scrollHeight: 5000, clientHeight: 800 });
    const other = renderHook(() => useAutoHideOnScroll({ topThreshold: 80, delta: 8, hideReserve: 88 }));
    act(() => other.result.current.scrollRef(big));
    act(() => scrollTo(big, 150));
    expect(other.result.current.hidden).toBe(true);
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
