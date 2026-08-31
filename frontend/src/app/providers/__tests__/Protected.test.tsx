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

  // super_admin が trainee 向けルート (/code-editor 等) にアクセスしたらホームへ飛ばす。
  // ノート（/notes）は旧ナレッジを統合した共有の面なので super_admin にも開く（対象外）。
  it('role=super_admin が /code-editor にアクセスするとホームにリダイレクト', () => {
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
        <MemoryRouter initialEntries={['/code-editor']}>
          <Routes>
            <Route
              path="/code-editor"
              element={
                <Protected>
                  <div>演習画面</div>
                </Protected>
              }
            />
            <Route path="/dashboard" element={<div>ホーム</div>} />
          </Routes>
        </MemoryRouter>
      </Provider>,
    );
    expect(screen.getByText('ホーム')).toBeInTheDocument();
    expect(screen.queryByText('演習画面')).not.toBeInTheDocument();
  });

  it('role=super_admin でも /notes は表示できる（旧ナレッジを統合した共有の面）', () => {
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
        <MemoryRouter initialEntries={['/notes']}>
          <Routes>
            <Route
              path="/notes"
              element={
                <Protected>
                  <div>ノート画面</div>
                </Protected>
              }
            />
            <Route path="/" element={<div>ホーム</div>} />
          </Routes>
        </MemoryRouter>
      </Provider>,
    );
    expect(screen.getByText('ノート画面')).toBeInTheDocument();
    expect(screen.queryByText('ホーム')).not.toBeInTheDocument();
  });

  // super_admin でも /admin 配下は通る。
  it('role=super_admin が /admin/members にアクセスすると children を表示', () => {
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
        <MemoryRouter initialEntries={['/admin/members']}>
          <Routes>
            <Route
              path="/admin/members"
              element={
                <Protected>
                  <div>従業員一覧</div>
                </Protected>
              }
            />
          </Routes>
        </MemoryRouter>
      </Provider>,
    );
    expect(screen.getByText('従業員一覧')).toBeInTheDocument();
  });

  // trainee は trainee ルートにアクセス可能。
  it('role=trainee は /notes にアクセスできる', () => {
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
        <MemoryRouter initialEntries={['/notes']}>
          <Routes>
            <Route
              path="/notes"
              element={
                <Protected>
                  <div>ノート画面</div>
                </Protected>
              }
            />
          </Routes>
        </MemoryRouter>
      </Provider>,
    );
    expect(screen.getByText('ノート画面')).toBeInTheDocument();
  });
});
