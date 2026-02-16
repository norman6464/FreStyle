import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, act, fireEvent } from '@testing-library/react';
import FormMessage from '../FormMessage';

describe('FormMessage', () => {
  it('エラーメッセージを表示する', () => {
    render(<FormMessage message={{ type: 'error', text: 'エラーが発生しました' }} />);

    expect(screen.getByText('エラーが発生しました')).toBeInTheDocument();
  });

  it('成功メッセージを表示する', () => {
    render(<FormMessage message={{ type: 'success', text: '保存しました' }} />);

    expect(screen.getByText('保存しました')).toBeInTheDocument();
  });

  it('messageがnullの場合何も表示しない', () => {
    const { container } = render(<FormMessage message={null} />);

    expect(container.firstChild).toBeNull();
  });

  it('エラーメッセージにエラースタイルが適用される', () => {
    render(<FormMessage message={{ type: 'error', text: 'エラー' }} />);

    const el = screen.getByText('エラー').closest('div');
    expect(el?.className).toContain('rose');
  });

  it('成功メッセージに成功スタイルが適用される', () => {
    render(<FormMessage message={{ type: 'success', text: '成功' }} />);

    const el = screen.getByText('成功').closest('div');
    expect(el?.className).toContain('emerald');
  });

  it('エラーメッセージにExclamationCircleIconが表示される', () => {
    render(<FormMessage message={{ type: 'error', text: 'エラー' }} />);

    expect(screen.getByText('エラー').closest('div')?.querySelector('svg')).toBeInTheDocument();
  });

  it('成功メッセージにCheckCircleIconが表示される', () => {
    render(<FormMessage message={{ type: 'success', text: '成功' }} />);

    expect(screen.getByText('成功').closest('div')?.querySelector('svg')).toBeInTheDocument();
  });

  it('長いメッセージテキストが正しく表示される', () => {
    const longText = 'エラー'.repeat(50);
    render(<FormMessage message={{ type: 'error', text: longText }} />);

    expect(screen.getByText(longText)).toBeInTheDocument();
  });

  it('HTMLタグを含むテキストがエスケープされる', () => {
    render(<FormMessage message={{ type: 'error', text: '<script>alert("xss")</script>' }} />);

    expect(screen.getByText('<script>alert("xss")</script>')).toBeInTheDocument();
  });

  it('エラーメッセージにborder-rose-800クラスが含まれる', () => {
    render(<FormMessage message={{ type: 'error', text: 'テスト' }} />);

    const el = screen.getByText('テスト').closest('div');
    expect(el?.className).toContain('border');
  });

  describe('自動消去', () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it('5秒後にonDismissが呼ばれる', () => {
      const onDismiss = vi.fn();
      render(<FormMessage message={{ type: 'error', text: 'エラー' }} onDismiss={onDismiss} />);

      act(() => {
        vi.advanceTimersByTime(5000);
      });

      expect(onDismiss).toHaveBeenCalledOnce();
    });

    it('onDismissが未指定の場合はタイマーが動作しない', () => {
      render(<FormMessage message={{ type: 'error', text: 'エラー' }} />);

      act(() => {
        vi.advanceTimersByTime(5000);
      });

      expect(screen.getByText('エラー')).toBeInTheDocument();
    });

    it('messageがnullの場合タイマーが設定されない', () => {
      const onDismiss = vi.fn();
      render(<FormMessage message={null} onDismiss={onDismiss} />);

      act(() => {
        vi.advanceTimersByTime(5000);
      });

      expect(onDismiss).not.toHaveBeenCalled();
    });
  });

  describe('閉じるボタン', () => {
    it('onDismiss指定時に閉じるボタンが表示される', () => {
      const onDismiss = vi.fn();
      render(<FormMessage message={{ type: 'error', text: 'エラー' }} onDismiss={onDismiss} />);

      expect(screen.getByRole('button', { name: '閉じる' })).toBeInTheDocument();
    });

    it('閉じるボタンクリックでonDismissが呼ばれる', () => {
      const onDismiss = vi.fn();
      render(<FormMessage message={{ type: 'error', text: 'エラー' }} onDismiss={onDismiss} />);

      fireEvent.click(screen.getByRole('button', { name: '閉じる' }));

      expect(onDismiss).toHaveBeenCalledOnce();
    });

    it('onDismiss未指定時に閉じるボタンが表示されない', () => {
      render(<FormMessage message={{ type: 'error', text: 'エラー' }} />);

      expect(screen.queryByRole('button', { name: '閉じる' })).not.toBeInTheDocument();
    });
  });
});
