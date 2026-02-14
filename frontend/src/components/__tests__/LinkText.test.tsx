import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import LinkText from '../LinkText';

describe('LinkText', () => {
  it('リンクテキストが表示される', () => {
    render(
      <MemoryRouter>
        <LinkText to="/signup">新規登録</LinkText>
      </MemoryRouter>
    );

    expect(screen.getByText('新規登録')).toBeInTheDocument();
  });

  it('正しいリンク先が設定される', () => {
    render(
      <MemoryRouter>
        <LinkText to="/signup">新規登録</LinkText>
      </MemoryRouter>
    );

    expect(screen.getByText('新規登録').closest('a')).toHaveAttribute('href', '/signup');
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
    expect(link?.className).toContain('text-primary-500');
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
