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
});
