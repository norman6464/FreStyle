import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import AuthInitializer from '../AuthInitializer';
import authReducer from '../../store/authSlice';
import authRepository from '../../repositories/AuthRepository';

vi.mock('../../repositories/AuthRepository');

function renderWithStore(initialLoading = true) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: { isAuthenticated: false, loading: initialLoading },
    },
  });
  return {
    store,
    ...render(
      <Provider store={store}>
        <AuthInitializer>
          <div>認証済みコンテンツ</div>
        </AuthInitializer>
      </Provider>,
    ),
  };
}

describe('AuthInitializer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('ローディング中はLoadingコンポーネントを表示する', () => {
    vi.mocked(authRepository.getCurrentUser).mockResolvedValue({
      id: 1,
      email: 'test@example.com',
      name: 'テスト',
      sub: 'sub-123',
    });

    renderWithStore(true);

    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('認証成功時にchildrenを表示する', async () => {
    vi.mocked(authRepository.getCurrentUser).mockResolvedValue({
      id: 1,
      email: 'test@example.com',
      name: 'テスト',
      sub: 'sub-123',
    });

    renderWithStore(true);

    await waitFor(() => {
      expect(screen.getByText('認証済みコンテンツ')).toBeInTheDocument();
    });
  });

  it('認証失敗時もchildrenを表示する', async () => {
    vi.mocked(authRepository.getCurrentUser).mockRejectedValue(new Error('Unauthorized'));

    renderWithStore(true);

    await waitFor(() => {
      expect(screen.getByText('認証済みコンテンツ')).toBeInTheDocument();
    });
  });

  it('認証成功時にisAuthenticatedがtrueになる', async () => {
    vi.mocked(authRepository.getCurrentUser).mockResolvedValue({
      id: 1,
      email: 'test@example.com',
      name: 'テスト',
      sub: 'sub-123',
    });

    const { store } = renderWithStore(true);

    await waitFor(() => {
      expect(store.getState().auth.isAuthenticated).toBe(true);
      expect(store.getState().auth.loading).toBe(false);
    });
  });

  it('認証失敗時にisAuthenticatedがfalseになる', async () => {
    vi.mocked(authRepository.getCurrentUser).mockRejectedValue(new Error('Unauthorized'));

    const { store } = renderWithStore(true);

    await waitFor(() => {
      expect(store.getState().auth.isAuthenticated).toBe(false);
      expect(store.getState().auth.loading).toBe(false);
    });
  });
});
