import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import SignupPage from '../ui/SignupPage';

function renderSignupPage() {
  return render(
    <MemoryRouter>
      <SignupPage />
    </MemoryRouter>
  );
}

describe('SignupPage', () => {
  it('見出しとメールで始めるボタンが表示される', () => {
    renderSignupPage();

    expect(screen.getByRole('heading', { name: 'アカウントを作成' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'メールで始める' })).toBeInTheDocument();
  });

  it('Google での開始導線がある', () => {
    renderSignupPage();
    expect(screen.getByRole('button', { name: /Google/ })).toBeInTheDocument();
  });

  it('ログインへの導線がある', () => {
    renderSignupPage();
    const loginLink = screen.getByRole('link', { name: 'ログイン' });
    expect(loginLink).toHaveAttribute('href', '/login');
  });
});
