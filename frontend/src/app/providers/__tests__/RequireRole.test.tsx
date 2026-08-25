import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import type { ReactNode } from 'react';
import RequireRole from '../RequireRole';
// reducer だけが要るので、Public API ではなく slice を直接読む。
// @/entities/user 経由だと barrel が全 repository と axios まで引き込み、
// このファイル単体のカバレッジ分母が 38 → 180 statements に膨らむ。
import authReducer from '@/entities/user/model/authSlice';

type Auth = { loading?: boolean; isAdmin?: boolean; role?: string | null };

/**
 * 通過・拒否は「どの画面が出たか」で判定する。実画面と同じく、ルート配下の中身を
 * 名前付きランドマーク（main）として置き、role + アクセシブルな名前で引く。
 */
const GATED_NAME = '管理画面の中身';
const REDIRECT_NAME = 'ダッシュボード';

const content = <main aria-label={GATED_NAME}>管理者向けの内容</main>;

/** ゲートを通れたか。通れていなければ null。 */
const gatedScreen = () => screen.queryByRole('main', { name: GATED_NAME });
/** /dashboard へ戻されたか。戻されていなければ null。 */
const redirectScreen = () => screen.queryByRole('main', { name: REDIRECT_NAME });

function renderGate(gate: ReactNode, auth: Auth) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: {
        isAuthenticated: true,
        loading: auth.loading ?? false,
        isAdmin: auth.isAdmin ?? false,
        role: auth.role ?? null,
        aiChatEnabledForTrainees: true,
      },
    },
  });

  return render(
    <Provider store={store}>
      <MemoryRouter initialEntries={['/admin/x']}>
        <Routes>
          <Route path="/admin/x" element={gate} />
          <Route
            path="/dashboard"
            element={<main aria-label={REDIRECT_NAME}>ダッシュボードの内容</main>}
          />
        </Routes>
      </MemoryRouter>
    </Provider>,
  );
}

describe('RequireRole', () => {
  it('allow に含まれる role なら children を表示する', () => {
    renderGate(<RequireRole allow={['super_admin']}>{content}</RequireRole>, {
      role: 'super_admin',
    });
    expect(gatedScreen()).toBeInTheDocument();
    expect(redirectScreen()).not.toBeInTheDocument();
  });

  it('allow が複数のときは、そのいずれかの role なら通す', () => {
    renderGate(
      <RequireRole allow={['super_admin', 'company_admin']}>{content}</RequireRole>,
      { role: 'company_admin' },
    );
    expect(gatedScreen()).toBeInTheDocument();
    expect(redirectScreen()).not.toBeInTheDocument();
  });

  it('allow に含まれない role は /dashboard へリダイレクトする', () => {
    renderGate(<RequireRole allow={['super_admin']}>{content}</RequireRole>, {
      role: 'company_admin',
      isAdmin: true,
    });
    expect(redirectScreen()).toBeInTheDocument();
    expect(gatedScreen()).not.toBeInTheDocument();
  });

  it('role 未確定（null）も allow に含まれないので /dashboard へリダイレクトする', () => {
    renderGate(<RequireRole allow={['super_admin']}>{content}</RequireRole>, { role: null });
    expect(redirectScreen()).toBeInTheDocument();
    expect(gatedScreen()).not.toBeInTheDocument();
  });

  it('認証情報の確認中はローディングを出し、children もリダイレクトも出さない', () => {
    renderGate(<RequireRole allow={['super_admin']}>{content}</RequireRole>, {
      loading: true,
      role: 'super_admin',
    });
    expect(screen.getByRole('status', { name: '読み込み中' })).toBeInTheDocument();
    expect(screen.getByText('認証情報を確認中...')).toBeInTheDocument();
    expect(gatedScreen()).not.toBeInTheDocument();
    expect(redirectScreen()).not.toBeInTheDocument();
  });

  it('requireAdminFlag を付けると、role が合っていても isAdmin=false はリダイレクトする', () => {
    renderGate(
      <RequireRole allow={['super_admin']} requireAdminFlag>
        {content}
      </RequireRole>,
      { role: 'super_admin', isAdmin: false },
    );
    expect(redirectScreen()).toBeInTheDocument();
    expect(gatedScreen()).not.toBeInTheDocument();
  });

  it('requireAdminFlag を付けても、role と isAdmin の両方が揃えば通す', () => {
    renderGate(
      <RequireRole allow={['super_admin']} requireAdminFlag>
        {content}
      </RequireRole>,
      { role: 'super_admin', isAdmin: true },
    );
    expect(gatedScreen()).toBeInTheDocument();
    expect(redirectScreen()).not.toBeInTheDocument();
  });

  it('allow="any" は role を問わず、isAdmin だけで通す', () => {
    renderGate(
      <RequireRole allow="any" requireAdminFlag>
        {content}
      </RequireRole>,
      { role: 'trainee', isAdmin: true },
    );
    expect(gatedScreen()).toBeInTheDocument();
    expect(redirectScreen()).not.toBeInTheDocument();
  });

  it('allow="any" でも isAdmin=false はリダイレクトする', () => {
    renderGate(
      <RequireRole allow="any" requireAdminFlag>
        {content}
      </RequireRole>,
      { role: 'super_admin', isAdmin: false },
    );
    expect(redirectScreen()).toBeInTheDocument();
    expect(gatedScreen()).not.toBeInTheDocument();
  });
});
