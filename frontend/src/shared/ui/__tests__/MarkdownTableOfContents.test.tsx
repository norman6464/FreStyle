import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import MarkdownTableOfContents from '../MarkdownTableOfContents';

// jsdom は IntersectionObserver を実装していないためスタブする。
// コールバックを捕捉して、任意のタイミングで交差イベントを再現できるようにする。
type IOCallback = (entries: Array<Partial<IntersectionObserverEntry>>) => void;
let capturedCallback: IOCallback | null = null;
const observed: Element[] = [];

class StubIntersectionObserver {
  constructor(cb: IOCallback) {
    capturedCallback = cb;
  }
  observe(el: Element) {
    observed.push(el);
  }
  disconnect() {}
  unobserve() {}
}

describe('MarkdownTableOfContents', () => {
  beforeEach(() => {
    capturedCallback = null;
    observed.length = 0;
    vi.stubGlobal('IntersectionObserver', StubIntersectionObserver);
  });
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('h1〜h3 を抽出してレベルに応じたアンカーリンクを描画する', () => {
    const content = ['# はじめに', '', '## セットアップ', '', '### 前提条件', '', '#### 深すぎる見出し'].join('\n');
    render(<MarkdownTableOfContents content={content} />);
    const nav = screen.getByRole('navigation', { name: '目次' });
    expect(nav).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'はじめに' })).toHaveAttribute('href', '#はじめに');
    expect(screen.getByRole('link', { name: 'セットアップ' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '前提条件' })).toBeInTheDocument();
    // h4 以下は目次に載せない。
    expect(screen.queryByRole('link', { name: '深すぎる見出し' })).not.toBeInTheDocument();
  });

  it('コードフェンス内の # 行は見出しとして扱わない', () => {
    const content = ['# 本物の見出し', '```bash', '# これはコメント', '```'].join('\n');
    render(<MarkdownTableOfContents content={content} />);
    expect(screen.getByRole('link', { name: '本物の見出し' })).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'これはコメント' })).not.toBeInTheDocument();
  });

  it('インライン記法(コード/太字/リンク)を剥がして表示する', () => {
    const content = '## `git status` と **確認** と [参考](https://example.com)';
    render(<MarkdownTableOfContents content={content} />);
    expect(screen.getByRole('link', { name: 'git status と 確認 と 参考' })).toBeInTheDocument();
  });

  it('見出しが無い場合は何も描画しない', () => {
    const { container } = render(<MarkdownTableOfContents content={'本文だけ。見出しなし。'} />);
    expect(container).toBeEmptyDOMElement();
    const empty = render(<MarkdownTableOfContents content={''} />);
    expect(empty.container).toBeEmptyDOMElement();
  });

  it('交差した見出しがハイライトされる(IntersectionObserver)', () => {
    render(
      <>
        {/* rehype-slug が本文側に生成する id 付き見出しを再現する。 */}
        <h2 id="setup" />
        <h2 id="usage" />
        <MarkdownTableOfContents content={'## Setup\n\n## Usage'} />
      </>,
    );
    expect(observed).toHaveLength(2);

    act(() => {
      capturedCallback?.([
        {
          isIntersecting: true,
          target: document.getElementById('usage')!,
          boundingClientRect: { top: 10 } as DOMRectReadOnly,
        },
      ]);
    });
    expect(screen.getByRole('link', { name: 'Usage' })).toHaveClass('border-taupe-500');
    expect(screen.getByRole('link', { name: 'Setup' })).not.toHaveClass('border-taupe-500');
  });
});
