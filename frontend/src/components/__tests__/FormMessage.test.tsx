import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
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
});
