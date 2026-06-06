import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import LoginPage from '../LoginPage';

// Hosted UI への遷移は window.location.href への代入で行うため、副作用を検証できるようにする。
const hrefSetter = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
  Object.defineProperty(window, 'location', {
    value: { set href(v: string) { hrefSetter(v); } },
    writable: true,
  });
});

function renderLoginPage() {
  return render(
    <MemoryRouter>
      <LoginPage />
    </MemoryRouter>
  );
}

describe('LoginPage', () => {
  it('Hosted UI ログインへの導線と公開ヘッダーが表示される', () => {
    renderLoginPage();

    expect(screen.getByRole('heading', { name: 'ログイン' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ログイン / 新規登録' })).toBeInTheDocument();
    // SRP のメール/パスワードフォームは廃止されている。
    expect(screen.queryByLabelText('メールアドレス')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('パスワード')).not.toBeInTheDocument();
  });

  it('利用申請への導線がある', () => {
    renderLoginPage();

    // ヘッダーと本文の双方に企業利用申請への導線がある。
    const applyLinks = screen.getAllByRole('link', { name: /利用申請/ });
    expect(applyLinks.length).toBeGreaterThan(0);
  });

  it('ログインボタンで Cognito Hosted UI へ遷移する', () => {
    renderLoginPage();

    screen.getByRole('button', { name: 'ログイン / 新規登録' }).click();
    expect(hrefSetter).toHaveBeenCalledTimes(1);
    expect(hrefSetter.mock.calls[0][0]).toContain('/oauth2/authorize');
  });
});
