import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import Protected from '../Protected';
import authReducer from '../../store/authSlice';

function renderWithAuth(isAuthenticated: boolean) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: { isAuthenticated, loading: false },
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
});
