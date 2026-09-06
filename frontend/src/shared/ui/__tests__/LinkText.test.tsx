import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import LinkText from '../LinkText';

describe('LinkText', () => {
  it('リンクテキストが表示される', () => {
    render(
      <MemoryRouter>
        <LinkText to="/signup">サインアップ</LinkText>
      </MemoryRouter>
    );

    expect(screen.getByText('サインアップ')).toBeInTheDocument();
  });

  it('正しいリンク先が設定される', () => {
    render(
      <MemoryRouter>
        <LinkText to="/signup">サインアップ</LinkText>
      </MemoryRouter>
    );

    expect(screen.getByText('サインアップ').closest('a')).toHaveAttribute('href', '/signup');
  });

  it('aタグとしてレンダリングされる', () => {
    render(
      <MemoryRouter>
        <LinkText to="/test">テスト</LinkText>
      </MemoryRouter>
    );

    expect(screen.getByText('テスト').closest('a')).toBeTruthy();
  });

  it('テキストスタイルが適用される', () => {
    render(
      <MemoryRouter>
        <LinkText to="/test">テスト</LinkText>
      </MemoryRouter>
    );

    const link = screen.getByText('テスト').closest('a');
    // 白地に対し brand-500 は 3.67:1 で、小さな文字の基準 4.5:1 に届かないため 700 を使う。
    expect(link?.className).toContain('text-brand-700');
    expect(link?.className).toContain('font-medium');
  });

  it('異なるパスで正しいリンク先が設定される', () => {
    render(
      <MemoryRouter>
        <LinkText to="/forgot-password">パスワードリセット</LinkText>
      </MemoryRouter>
    );

    expect(screen.getByText('パスワードリセット').closest('a')).toHaveAttribute('href', '/forgot-password');
  });
});
