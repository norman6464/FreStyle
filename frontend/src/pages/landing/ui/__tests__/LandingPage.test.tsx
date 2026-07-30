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
    // CTA（複数箇所にあるので存在のみ確認）
    expect(screen.getAllByRole('link', { name: /利用申請/ }).length).toBeGreaterThan(0);
  });

  it('ログイン済みなら /dashboard へリダイレクトする', () => {
    renderLanding(true);
    expect(screen.getByText('ダッシュボード')).toBeInTheDocument();
  });
});
