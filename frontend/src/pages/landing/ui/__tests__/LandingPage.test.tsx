import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import LandingPage from '../LandingPage';
import authReducer from '@/entities/user/model/authSlice';

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
        name: /新卒ITエンジニア向け研修プラットフォーム/,
      }),
    ).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'FreStyle とは' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: '主な機能' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'よくある質問' })).toBeInTheDocument();
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

  it('ログイン済みなら /dashboard へリダイレクトする', () => {
    renderLanding(true);
    expect(screen.getByText('ダッシュボード')).toBeInTheDocument();
  });

  it('LP 自身がスクロールコンテナを持つ（body overflow hidden 下でスクロール不能になった回帰: FRESTYLE-223）', () => {
    const { container } = renderLanding(false);
    // ルート要素そのものがスクロールコンテナであること（子孫のどこかでは不十分）
    expect(container.firstElementChild).toHaveClass('h-full', 'overflow-y-auto');
  });
});
