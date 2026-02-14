import { render, screen, fireEvent, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import Tooltip from '../Tooltip';

describe('Tooltip', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('子要素が表示される', () => {
    render(
      <Tooltip content="説明文">
        <button>ボタン</button>
      </Tooltip>
    );
    expect(screen.getByText('ボタン')).toBeInTheDocument();
  });

  it('初期状態ではツールチップが非表示', () => {
    render(
      <Tooltip content="説明文">
        <button>ボタン</button>
      </Tooltip>
    );
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
  });

  it('マウスオーバーでツールチップが表示される', () => {
    render(
      <Tooltip content="説明文">
        <button>ボタン</button>
      </Tooltip>
    );
    fireEvent.mouseEnter(screen.getByText('ボタン').parentElement!);
    act(() => {
      vi.advanceTimersByTime(300);
    });
    expect(screen.getByRole('tooltip')).toBeInTheDocument();
    expect(screen.getByText('説明文')).toBeInTheDocument();
  });

  it('マウスアウトでツールチップが非表示になる', () => {
    render(
      <Tooltip content="説明文">
        <button>ボタン</button>
      </Tooltip>
    );
    fireEvent.mouseEnter(screen.getByText('ボタン').parentElement!);
    act(() => {
      vi.advanceTimersByTime(300);
    });
    expect(screen.getByRole('tooltip')).toBeInTheDocument();
    fireEvent.mouseLeave(screen.getByText('ボタン').parentElement!);
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
  });

  it('遅延前にマウスアウトするとツールチップが表示されない', () => {
    render(
      <Tooltip content="説明文">
        <button>ボタン</button>
      </Tooltip>
    );
    fireEvent.mouseEnter(screen.getByText('ボタン').parentElement!);
    act(() => {
      vi.advanceTimersByTime(100);
    });
    fireEvent.mouseLeave(screen.getByText('ボタン').parentElement!);
    act(() => {
      vi.advanceTimersByTime(300);
    });
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
  });

  it('position=bottomでツールチップが下に表示される', () => {
    render(
      <Tooltip content="下部説明" position="bottom">
        <button>ボタン</button>
      </Tooltip>
    );
    fireEvent.mouseEnter(screen.getByText('ボタン').parentElement!);
    act(() => {
      vi.advanceTimersByTime(300);
    });
    const tooltip = screen.getByRole('tooltip');
    expect(tooltip.className).toContain('top-full');
    expect(tooltip.className).toContain('mt-1.5');
  });

  it('position=topでツールチップが上に表示される', () => {
    render(
      <Tooltip content="上部説明">
        <button>ボタン</button>
      </Tooltip>
    );
    fireEvent.mouseEnter(screen.getByText('ボタン').parentElement!);
    act(() => {
      vi.advanceTimersByTime(300);
    });
    const tooltip = screen.getByRole('tooltip');
    expect(tooltip.className).toContain('bottom-full');
    expect(tooltip.className).toContain('mb-1.5');
  });

  it('ツールチップにcontent文字列が表示される', () => {
    render(
      <Tooltip content="ヘルプテキスト">
        <span>アイコン</span>
      </Tooltip>
    );
    fireEvent.mouseEnter(screen.getByText('アイコン').parentElement!);
    act(() => {
      vi.advanceTimersByTime(300);
    });
    expect(screen.getByText('ヘルプテキスト')).toBeInTheDocument();
  });
});
