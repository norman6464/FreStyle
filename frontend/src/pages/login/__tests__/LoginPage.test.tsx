import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
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
    // 「在ること」だけでなく「押せること」まで見る。設定が揃っているのに
    // 押せない状態も、押せるのに何も起きない状態も、ここで落ちる。
    expect(screen.getByRole('button', { name: 'ログインする' })).toBeEnabled();
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
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

/*
 * 設定が欠けているときの姿。
 *
 * ボタンを消さず、押せない状態のまま理由を添える。消してしまうと外から見て
 * 「壊れているのか、意図的に止めているのか」が区別できない。
 *
 * ここで検査しているのは表示だけで、**押せて何も起きない状態を作らせない**のは
 * 型（useOidcLogin の合併に start が無い枝）の役目。表示の検査を外しても
 * その保証は残るが、逆は成り立たない。
 */
describe('LoginPage（認可の設定が欠けているとき）', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_OIDC_AUTHORIZE_URI', '');
    vi.stubEnv('VITE_OIDC_CLIENT_ID', '');
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it('ログインボタンは消えず、押せない状態で残る', () => {
    renderLoginPage();
    const button = screen.getByRole('button', { name: 'ログインする' });
    expect(button).toBeInTheDocument();
    expect(button).toBeDisabled();
  });

  it('Google の導線も押せない', () => {
    renderLoginPage();
    expect(screen.getByRole('button', { name: /Google/ })).toBeDisabled();
  });

  it('押せない理由を画面に出す', () => {
    renderLoginPage();
    const notice = screen.getByRole('status');
    expect(notice).toHaveTextContent('現在ログインを受け付けていません');
  });

  // 欠けている設定の名前は人が読む文には出さない（利用者に意味が無い）。
  // 運用する側が要素を見れば分かるよう、属性には載せる。
  it('欠けている設定の名前は文ではなく属性に載せる', () => {
    renderLoginPage();
    const notice = screen.getByRole('status');
    expect(notice.textContent).not.toContain('VITE_');
    expect(notice.getAttribute('data-missing')).toContain('VITE_OIDC_AUTHORIZE_URI');
    expect(notice.getAttribute('data-missing')).toContain('VITE_OIDC_CLIENT_ID');
  });
});
