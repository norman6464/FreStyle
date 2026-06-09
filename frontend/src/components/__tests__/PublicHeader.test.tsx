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
  it('ログインページでは CTA が「新規登録」(→/signup)', () => {
    renderAt('/login');
    const cta = screen.getByRole('link', { name: '新規登録' });
    expect(cta).toHaveAttribute('href', '/signup');
    expect(screen.queryByRole('link', { name: 'ログイン' })).not.toBeInTheDocument();
  });

  it('新規登録ページでは CTA が「ログイン」(→/login)', () => {
    renderAt('/signup');
    const cta = screen.getByRole('link', { name: 'ログイン' });
    expect(cta).toHaveAttribute('href', '/login');
    expect(screen.queryByRole('link', { name: '新規登録' })).not.toBeInTheDocument();
  });

  it('企業の利用申請への導線は常にある', () => {
    renderAt('/login');
    expect(screen.getByRole('link', { name: /企業の利用申請/ })).toBeInTheDocument();
  });
});
