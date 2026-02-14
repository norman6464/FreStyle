import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import ScrollToTop from '../ScrollToTop';

describe('ScrollToTop', () => {
  let scrollContainer: HTMLDivElement;

  beforeEach(() => {
    scrollContainer = document.createElement('div');
    scrollContainer.id = 'scroll-target';
    Object.defineProperty(scrollContainer, 'scrollTop', { value: 0, writable: true });
    scrollContainer.scrollTo = vi.fn();
    document.body.appendChild(scrollContainer);
  });

  afterEach(() => {
    document.body.removeChild(scrollContainer);
  });

  it('初期状態ではボタンが非表示', () => {
    render(<ScrollToTop targetId="scroll-target" />);
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('スクロール後にボタンが表示される', () => {
    render(<ScrollToTop targetId="scroll-target" threshold={100} />);
    Object.defineProperty(scrollContainer, 'scrollTop', { value: 150 });
    act(() => {
      fireEvent.scroll(scrollContainer);
    });
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('ボタンクリックでscrollToが呼ばれる', () => {
    render(<ScrollToTop targetId="scroll-target" threshold={100} />);
    Object.defineProperty(scrollContainer, 'scrollTop', { value: 200 });
    act(() => {
      fireEvent.scroll(scrollContainer);
    });
    fireEvent.click(screen.getByRole('button'));
    expect(scrollContainer.scrollTo).toHaveBeenCalledWith({ top: 0, behavior: 'smooth' });
  });

  it('aria-labelが設定される', () => {
    render(<ScrollToTop targetId="scroll-target" threshold={0} />);
    Object.defineProperty(scrollContainer, 'scrollTop', { value: 10 });
    act(() => {
      fireEvent.scroll(scrollContainer);
    });
    expect(screen.getByRole('button')).toHaveAttribute('aria-label', 'ページ上部に戻る');
  });

  it('閾値以下にスクロールするとボタンが非表示になる', () => {
    render(<ScrollToTop targetId="scroll-target" threshold={100} />);
    // スクロールして表示
    Object.defineProperty(scrollContainer, 'scrollTop', { value: 200 });
    act(() => {
      fireEvent.scroll(scrollContainer);
    });
    expect(screen.getByRole('button')).toBeInTheDocument();

    // 上に戻って非表示
    Object.defineProperty(scrollContainer, 'scrollTop', { value: 50 });
    act(() => {
      fireEvent.scroll(scrollContainer);
    });
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('デフォルトのthresholdが200である', () => {
    render(<ScrollToTop targetId="scroll-target" />);
    // 199ではまだ表示されない
    Object.defineProperty(scrollContainer, 'scrollTop', { value: 199 });
    act(() => {
      fireEvent.scroll(scrollContainer);
    });
    expect(screen.queryByRole('button')).not.toBeInTheDocument();

    // 200で表示される
    Object.defineProperty(scrollContainer, 'scrollTop', { value: 200 });
    act(() => {
      fireEvent.scroll(scrollContainer);
    });
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('ターゲット要素が存在しない場合エラーにならない', () => {
    expect(() => render(<ScrollToTop targetId="nonexistent" />)).not.toThrow();
  });

  it('ちょうど閾値でボタンが表示される', () => {
    render(<ScrollToTop targetId="scroll-target" threshold={100} />);
    Object.defineProperty(scrollContainer, 'scrollTop', { value: 100 });
    act(() => {
      fireEvent.scroll(scrollContainer);
    });
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('ボタンにfixed配置クラスが適用される', () => {
    render(<ScrollToTop targetId="scroll-target" threshold={0} />);
    Object.defineProperty(scrollContainer, 'scrollTop', { value: 10 });
    act(() => {
      fireEvent.scroll(scrollContainer);
    });
    expect(screen.getByRole('button').className).toContain('fixed');
    expect(screen.getByRole('button').className).toContain('bottom-6');
    expect(screen.getByRole('button').className).toContain('right-6');
  });

  it('アンマウント時にスクロールイベントリスナーがクリーンアップされる', () => {
    const removeSpy = vi.spyOn(scrollContainer, 'removeEventListener');
    const { unmount } = render(<ScrollToTop targetId="scroll-target" />);
    unmount();
    expect(removeSpy).toHaveBeenCalledWith('scroll', expect.any(Function));
    removeSpy.mockRestore();
  });
});
