import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import Protected from '../Protected';
import authReducer from '@/entities/user/model/authSlice';

function renderWithAuth(isAuthenticated: boolean) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: { isAuthenticated, loading: false, isAdmin: false, role: null },
    },
  });

  return render(
    <Provider store={store}>
      <MemoryRouter initialEntries={['/protected']}>
        <Routes>
          <Route
            path="/protected"
            element={
              <Protected>
                <div>保護されたコンテンツ</div>
              </Protected>
            }
          />
          <Route path="/login" element={<div>ログインページ</div>} />
        </Routes>
      </MemoryRouter>
    </Provider>,
  );
}

describe('Protected', () => {
  it('認証済みの場合、childrenを表示する', () => {
    renderWithAuth(true);
    expect(screen.getByText('保護されたコンテンツ')).toBeInTheDocument();
  });

  it('認証済みの場合、ログインページにリダイレクトしない', () => {
    renderWithAuth(true);
    expect(screen.queryByText('ログインページ')).not.toBeInTheDocument();
  });

  it('未認証の場合、/login にリダイレクトする', () => {
    renderWithAuth(false);
    expect(screen.getByText('ログインページ')).toBeInTheDocument();
  });

  it('未認証の場合、childrenを表示しない', () => {
    renderWithAuth(false);
    expect(screen.queryByText('保護されたコンテンツ')).not.toBeInTheDocument();
  });

  // super_admin が trainee 向けルート (/chat/ask-ai 等) にアクセスしたら /admin/companies に飛ばす。
  it('role=super_admin が /chat/ask-ai にアクセスすると /admin/companies にリダイレクト', () => {
    const store = configureStore({
      reducer: { auth: authReducer },
      preloadedState: {
        auth: {
          isAuthenticated: true,
          loading: false,
          isAdmin: true,
          role: 'super_admin',
        },
      },
    });
    render(
      <Provider store={store}>
        <MemoryRouter initialEntries={['/chat/ask-ai']}>
          <Routes>
            <Route
              path="/chat/ask-ai"
              element={
                <Protected>
                  <div>AI チャット画面</div>
                </Protected>
              }
            />
            <Route path="/admin/companies" element={<div>会社一覧</div>} />
          </Routes>
        </MemoryRouter>
      </Provider>,
    );
    expect(screen.getByText('会社一覧')).toBeInTheDocument();
    expect(screen.queryByText('AI チャット画面')).not.toBeInTheDocument();
  });

  // super_admin でも /admin 配下は通る。
  it('role=super_admin が /admin/companies にアクセスすると children を表示', () => {
    const store = configureStore({
      reducer: { auth: authReducer },
      preloadedState: {
        auth: {
          isAuthenticated: true,
          loading: false,
          isAdmin: true,
          role: 'super_admin',
        },
      },
    });
    render(
      <Provider store={store}>
        <MemoryRouter initialEntries={['/admin/companies']}>
          <Routes>
            <Route
              path="/admin/companies"
              element={
                <Protected>
                  <div>会社一覧</div>
                </Protected>
              }
            />
          </Routes>
        </MemoryRouter>
      </Provider>,
    );
    expect(screen.getByText('会社一覧')).toBeInTheDocument();
  });

  // trainee は trainee ルートにアクセス可能。
  it('role=trainee は /chat/ask-ai にアクセスできる', () => {
    const store = configureStore({
      reducer: { auth: authReducer },
      preloadedState: {
        auth: {
          isAuthenticated: true,
          loading: false,
          isAdmin: false,
          role: 'trainee',
        },
      },
    });
    render(
      <Provider store={store}>
        <MemoryRouter initialEntries={['/chat/ask-ai']}>
          <Routes>
            <Route
              path="/chat/ask-ai"
              element={
                <Protected>
                  <div>AI チャット画面</div>
                </Protected>
              }
            />
          </Routes>
        </MemoryRouter>
      </Provider>,
    );
    expect(screen.getByText('AI チャット画面')).toBeInTheDocument();
  });
});
