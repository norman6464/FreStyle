import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import LoginPage from '../ui/LoginPage';

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

  it('アカウント作成への導線がヘッダーと本文の両方にある', () => {
    renderLoginPage();
    // ヘッダーと本文の 2 箇所。
    const signupLinks = screen.getAllByRole('link', { name: /アカウントを作成/ });
    expect(signupLinks.length).toBeGreaterThanOrEqual(2);
  });
});
