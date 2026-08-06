import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import LandingPage from '../LandingPage';
import authReducer from '@/entities/user/model/authSlice';
import AuthRepository from '@/entities/user/api/authRepository';

// マウント時の認証確認(ログイン済みなら /dashboard へ送る)をモックする。
// 既定は 401(未ログイン)扱い。公開ページなので probeCurrentUser を使う(FRESTYLE-225)。
vi.mock('@/entities/user/api/authRepository', () => ({
  default: { probeCurrentUser: vi.fn().mockRejectedValue(new Error('unauthorized')) },
}));
const mockGetCurrentUser = vi.mocked(AuthRepository.probeCurrentUser);

function renderLanding(isAuthenticated: boolean) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: { isAuthenticated, loading: false, isAdmin: false, role: null },
    },
  });
  return render(
    <Provider store={store}>
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route path="/" element={<LandingPage />} />
          <Route path="/dashboard" element={<div>ダッシュボード</div>} />
        </Routes>
      </MemoryRouter>
    </Provider>,
  );
}

describe('LandingPage', () => {
  it('未ログインでは公開LPのヒーロー・主要セクションを表示する', () => {
    renderLanding(false);
    expect(
      screen.getByRole('heading', {
        level: 1,
        name: /「わかる」で終わらせない/,
      }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole('heading', { name: /研修をつなぎ、学びの完了まで届ける機能群/ }),
    ).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /学びっぱなしにさせない/ })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'FreStyle で変わる新人研修' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: '立場ごとに、見える景色を用意' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: '企業利用を前提にした設計' })).toBeInTheDocument();
    // CTA は遷移先まで検証する（存在確認だけでは誤配線を検知できない）
    const applyLinks = screen.getAllByRole('link', { name: /利用申請/ });
    expect(applyLinks.length).toBeGreaterThan(0);
    for (const link of applyLinks) {
      expect(link).toHaveAttribute('href', '/company-application');
    }
    const loginLinks = screen.getAllByRole('link', { name: /ログイン/ });
    expect(loginLinks.length).toBeGreaterThan(0);
    for (const link of loginLinks) {
      expect(link).toHaveAttribute('href', '/login');
    }
  });

  it('Before/After 比較表を行見出し付きの table で表示する', () => {
    renderLanding(false);
    const table = screen.getByRole('table');
    expect(table).toBeInTheDocument();
    expect(screen.getByRole('rowheader', { name: '従来の研修' })).toBeInTheDocument();
    expect(screen.getByRole('rowheader', { name: 'FreStyle' })).toBeInTheDocument();
    for (const col of ['教材と学び方', '質問対応', '進捗の把握']) {
      expect(screen.getByRole('columnheader', { name: col })).toBeInTheDocument();
    }
  });

  it('ログイン済みなら /dashboard へリダイレクトする', () => {
    renderLanding(true);
    expect(screen.getByText('ダッシュボード')).toBeInTheDocument();
  });

  it('初回ロードで認証確認が成功したら /dashboard へ自動遷移する(FRESTYLE-225)', async () => {
    // Cookie でログイン済みのユーザーが直接 / を開いたケース: store は未認証で始まるが、
    // マウント時の /auth/me 確認が成功したらダッシュボードへ送られる。
    mockGetCurrentUser.mockResolvedValueOnce({ isAdmin: false, role: 'trainee' });
    renderLanding(false);
    expect(await screen.findByText('ダッシュボード')).toBeInTheDocument();
  });

  it('未ログイン(401)なら LP を表示し続ける', async () => {
    renderLanding(false);
    // 認証確認が reject された後も LP のまま
    expect(
      await screen.findByRole('heading', { level: 1, name: /「わかる」で終わらせない/ }),
    ).toBeInTheDocument();
    expect(screen.queryByText('ダッシュボード')).not.toBeInTheDocument();
  });

  it('LP 自身がスクロールコンテナを持つ（body overflow hidden 下でスクロール不能になった回帰: FRESTYLE-223）', () => {
    const { container } = renderLanding(false);
    // ルート要素そのものがスクロールコンテナであること（子孫のどこかでは不十分）
    expect(container.firstElementChild).toHaveClass('h-full', 'overflow-y-auto');
  });
});
