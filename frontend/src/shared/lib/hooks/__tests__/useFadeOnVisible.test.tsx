import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render } from '@testing-library/react';
import { useRef } from 'react';
import { useFadeOnVisible } from '../useFadeOnVisible';

// IntersectionObserver のスタブ。observe された要素を記録し、任意のタイミングで
// 「交差した」ことにできるようにする（jsdom には実装が無い）。
type IOCallback = (entries: Array<{ isIntersecting: boolean; target: Element }>) => void;
let capturedCallback: IOCallback | null = null;
let observed: Element[] = [];
let unobserved: Element[] = [];
let disconnected = 0;

class StubIntersectionObserver {
  constructor(cb: IOCallback) {
    capturedCallback = cb;
  }
  observe(el: Element) {
    observed.push(el);
  }
  unobserve(el: Element) {
    unobserved.push(el);
  }
  disconnect() {
    disconnected += 1;
  }
}

/**
 * 可視判定は getBoundingClientRect で行うが、jsdom は常に 0 を返すため全要素が
 * 「画面内」になってしまう。id ベースでプロトタイプごとスタブし、走査より先に効かせる。
 */
function stubRectsById(visibleIds: string[]) {
  vi.spyOn(Element.prototype, 'getBoundingClientRect').mockImplementation(function (
    this: Element,
  ) {
    const visible = visibleIds.includes((this as HTMLElement).id);
    // jsdom の window.innerHeight は 768。画面外ははるか下に置く。
    return { top: visible ? 100 : 5000, bottom: visible ? 120 : 5020 } as DOMRect;
  });
}

function Harness({
  html,
  enabled = true,
  scanKey = 0,
}: {
  html: string;
  enabled?: boolean;
  /** 本文が同じまま再走査させたいとき用（DOM を作り直さずに contentKey だけ変える）。 */
  scanKey?: number;
}) {
  const ref = useRef<HTMLDivElement | null>(null);
  useFadeOnVisible(ref, enabled, scanKey);
  return <div ref={ref} dangerouslySetInnerHTML={{ __html: html }} />;
}

const TWO_SEGS =
  '<span class="fade-seg" id="a">見えている、</span><span class="fade-seg" id="b">画面外です。</span>';

describe('useFadeOnVisible (FRESTYLE-153)', () => {
  beforeEach(() => {
    capturedCallback = null;
    observed = [];
    unobserved = [];
    disconnected = 0;
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it('IntersectionObserver が無い環境では何も保留しない（従来どおり mount 時フェード）', () => {
    vi.stubGlobal('IntersectionObserver', undefined);
    stubRectsById([]); // 全部画面外でも保留しない
    const { container } = render(<Harness html={TWO_SEGS} />);
    // data-fade が一切付かない = CSS の mount アニメがそのまま走る。
    expect(container.querySelectorAll('[data-fade]')).toHaveLength(0);
  });

  it('画面内のチャンクは即フェード、画面外のチャンクは保留して監視する', () => {
    vi.stubGlobal('IntersectionObserver', StubIntersectionObserver);
    stubRectsById(['a']);
    const { container } = render(<Harness html={TWO_SEGS} />);

    const a = container.querySelector('#a') as HTMLElement;
    const b = container.querySelector('#b') as HTMLElement;
    expect(a.dataset.fade).toBe('now');
    expect(b.dataset.fade).toBe('defer');
    expect(observed).toContain(b);
    expect(observed).not.toContain(a);
  });

  it('保留したチャンクは画面に入った時点でフェードを再生する', () => {
    vi.stubGlobal('IntersectionObserver', StubIntersectionObserver);
    stubRectsById([]); // すべて画面外
    const { container } = render(<Harness html={TWO_SEGS} />);

    const b = container.querySelector('#b') as HTMLElement;
    expect(b.dataset.fade).toBe('defer');

    // スクロールして画面に入った。
    capturedCallback?.([{ isIntersecting: true, target: b }]);

    expect(b.dataset.fade).toBe('now');
    // 一度フェードさせたら監視を外す（再スクロールで再フェードしない）。
    expect(unobserved).toContain(b);
  });

  it('交差していない通知では保留のままにする', () => {
    vi.stubGlobal('IntersectionObserver', StubIntersectionObserver);
    stubRectsById([]);
    const { container } = render(<Harness html={TWO_SEGS} />);

    const b = container.querySelector('#b') as HTMLElement;
    capturedCallback?.([{ isIntersecting: false, target: b }]);
    expect(b.dataset.fade).toBe('defer');
  });

  it('判定済みのチャンクは再走査しても状態を変えない（再フェードしない）', () => {
    vi.stubGlobal('IntersectionObserver', StubIntersectionObserver);
    stubRectsById([]);
    const { container, rerender } = render(<Harness html={TWO_SEGS} scanKey={0} />);

    const b = container.querySelector('#b') as HTMLElement;
    capturedCallback?.([{ isIntersecting: true, target: b }]);
    expect(b.dataset.fade).toBe('now');

    // 本文が伸びて再走査が走っても、既判定の span は触らない（DOM は同一のまま）。
    rerender(<Harness html={TWO_SEGS} scanKey={1} />);
    expect(b.dataset.fade).toBe('now');
  });

  it('enabled が false のときは走査しない', () => {
    vi.stubGlobal('IntersectionObserver', StubIntersectionObserver);
    stubRectsById([]);
    const { container } = render(<Harness html={TWO_SEGS} enabled={false} />);
    expect(container.querySelectorAll('[data-fade]')).toHaveLength(0);
  });

  it('unmount で監視を解除する', () => {
    vi.stubGlobal('IntersectionObserver', StubIntersectionObserver);
    stubRectsById([]);
    const { unmount } = render(<Harness html={TWO_SEGS} />);
    unmount();
    expect(disconnected).toBeGreaterThan(0);
  });
});
