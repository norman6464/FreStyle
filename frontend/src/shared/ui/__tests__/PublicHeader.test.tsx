import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import PublicHeader from '../PublicHeader';

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <PublicHeader />
    </MemoryRouter>
  );
}

describe('PublicHeader', () => {
  it('ログイン画面ではアカウント作成への導線がある', () => {
    renderAt('/login');
    const signup = screen.getByRole('link', { name: /アカウントを作成/ });
    expect(signup).toHaveAttribute('href', '/signup');
  });

  it('サインアップ画面では自己参照リンクを出さず、ログインへの導線を出す', () => {
    renderAt('/signup');
    expect(screen.queryByRole('link', { name: /アカウントを作成/ })).not.toBeInTheDocument();
    const login = screen.getByRole('link', { name: /ログイン/ });
    expect(login).toHaveAttribute('href', '/login');
  });
});
