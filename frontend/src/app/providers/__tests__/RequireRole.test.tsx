import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import type { ReactNode } from 'react';
import RequireRole from '../RequireRole';
import authReducer from '@/entities/user/model/authSlice';

type Auth = { loading?: boolean; isAdmin?: boolean; role?: string | null };

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
          <Route path="/dashboard" element={<div>ダッシュボード</div>} />
        </Routes>
      </MemoryRouter>
    </Provider>,
  );
}

const content = <div>管理画面の中身</div>;

describe('RequireRole', () => {
  it('allow に含まれる role なら children を表示する', () => {
    renderGate(<RequireRole allow={['super_admin']}>{content}</RequireRole>, {
      role: 'super_admin',
    });
    expect(screen.getByText('管理画面の中身')).toBeInTheDocument();
    expect(screen.queryByText('ダッシュボード')).not.toBeInTheDocument();
  });

  it('allow が複数のときは、そのいずれかの role なら通す', () => {
    renderGate(
      <RequireRole allow={['super_admin', 'company_admin']}>{content}</RequireRole>,
      { role: 'company_admin' },
    );
    expect(screen.getByText('管理画面の中身')).toBeInTheDocument();
  });

  it('allow に含まれない role は /dashboard へリダイレクトする', () => {
    renderGate(<RequireRole allow={['super_admin']}>{content}</RequireRole>, {
      role: 'company_admin',
      isAdmin: true,
    });
    expect(screen.getByText('ダッシュボード')).toBeInTheDocument();
    expect(screen.queryByText('管理画面の中身')).not.toBeInTheDocument();
  });

  it('role 未確定（null）も allow に含まれないので /dashboard へリダイレクトする', () => {
    renderGate(<RequireRole allow={['super_admin']}>{content}</RequireRole>, { role: null });
    expect(screen.getByText('ダッシュボード')).toBeInTheDocument();
  });

  it('認証情報の確認中はローディングを出し、children もリダイレクトも出さない', () => {
    renderGate(<RequireRole allow={['super_admin']}>{content}</RequireRole>, {
      loading: true,
      role: 'super_admin',
    });
    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.getByText('認証情報を確認中...')).toBeInTheDocument();
    expect(screen.queryByText('管理画面の中身')).not.toBeInTheDocument();
    expect(screen.queryByText('ダッシュボード')).not.toBeInTheDocument();
  });

  it('requireAdminFlag を付けると、role が合っていても isAdmin=false はリダイレクトする', () => {
    renderGate(
      <RequireRole allow={['super_admin']} requireAdminFlag>
        {content}
      </RequireRole>,
      { role: 'super_admin', isAdmin: false },
    );
    expect(screen.getByText('ダッシュボード')).toBeInTheDocument();
  });

  it('requireAdminFlag を付けても、role と isAdmin の両方が揃えば通す', () => {
    renderGate(
      <RequireRole allow={['super_admin']} requireAdminFlag>
        {content}
      </RequireRole>,
      { role: 'super_admin', isAdmin: true },
    );
    expect(screen.getByText('管理画面の中身')).toBeInTheDocument();
  });

  it('allow="any" は role を問わず、isAdmin だけで通す', () => {
    renderGate(
      <RequireRole allow="any" requireAdminFlag>
        {content}
      </RequireRole>,
      { role: 'trainee', isAdmin: true },
    );
    expect(screen.getByText('管理画面の中身')).toBeInTheDocument();
  });

  it('allow="any" でも isAdmin=false はリダイレクトする', () => {
    renderGate(
      <RequireRole allow="any" requireAdminFlag>
        {content}
      </RequireRole>,
      { role: 'super_admin', isAdmin: false },
    );
    expect(screen.getByText('ダッシュボード')).toBeInTheDocument();
    expect(screen.queryByText('管理画面の中身')).not.toBeInTheDocument();
  });
});
