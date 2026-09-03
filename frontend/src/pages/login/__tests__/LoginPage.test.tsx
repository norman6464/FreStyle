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
  it('発行者のログイン画面へ送るボタンだけを置く', () => {
    renderLoginPage();

    expect(screen.getByRole('heading', { name: 'ログイン' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ログインする' })).toBeInTheDocument();
  });

  // パスワードを受け取るのは発行者のログイン画面の役目。アプリが受け取ると、
  // 二要素・ロックアウト・パスワードの強さといった発行者側の守りを素通りする
  // 経路を自分で開くことになる。
  it('メールとパスワードの入力欄を置かない', () => {
    renderLoginPage();

    expect(screen.queryByLabelText('メールアドレス')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('パスワード')).not.toBeInTheDocument();
    expect(screen.queryByRole('form', { name: 'ログインフォーム' })).not.toBeInTheDocument();
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
