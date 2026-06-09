import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import LoginPage from '../LoginPage';

function renderLoginPage() {
  return render(
    <MemoryRouter>
      <LoginPage />
    </MemoryRouter>
  );
}

describe('LoginPage', () => {
  it('メール/パスワードフォームと公開ヘッダーが表示される', () => {
    renderLoginPage();

    expect(screen.getByRole('heading', { name: 'ログイン' })).toBeInTheDocument();
    expect(screen.getByRole('form', { name: 'ログインフォーム' })).toBeInTheDocument();
    expect(screen.getByLabelText('メールアドレス')).toBeInTheDocument();
    expect(screen.getByLabelText('パスワード')).toBeInTheDocument();
    // フォーム送信ボタン（ヘッダーの「ログイン」はリンクなので button では一意になる）。
    expect(screen.getByRole('button', { name: 'ログイン' })).toBeInTheDocument();
  });

  it('Google ログイン導線がある', () => {
    renderLoginPage();
    expect(screen.getByRole('button', { name: /Google/ })).toBeInTheDocument();
  });

  it('利用申請への導線がある（ヘッダー + 本文）', () => {
    renderLoginPage();
    const applyLinks = screen.getAllByRole('link', { name: /利用申請/ });
    expect(applyLinks.length).toBeGreaterThan(0);
  });
});
