import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import { createElement } from 'react';
import { useAuth } from '../useAuth';
import AuthRepository from '../../repositories/AuthRepository';
import authReducer from '../../store/authSlice';

vi.mock('../../repositories/AuthRepository');

const mockedRepo = vi.mocked(AuthRepository);

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

function createWrapper() {
  const store = configureStore({ reducer: { auth: authReducer } });
  return ({ children }: { children: React.ReactNode }) =>
    createElement(Provider, { store },
      createElement(MemoryRouter, null, children)
    );
}

describe('useAuth', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('login: ログイン成功時にtrueを返す', async () => {
    const mockUser = { id: 1, email: 'test@example.com', name: 'テスト', sub: 'sub-123' };
    mockedRepo.login.mockResolvedValue(mockUser);

    const { result } = renderHook(() => useAuth(), { wrapper: createWrapper() });

    let success: boolean = false;
    await act(async () => {
      success = await result.current.login({ email: 'test@example.com', password: 'password' });
    });

    expect(success).toBe(true);
    expect(result.current.isAuthenticated).toBe(true);
  });

  it('login: ログイン失敗時にfalseとerrorを返す', async () => {
    mockedRepo.login.mockRejectedValue(new Error('認証失敗'));

    const { result } = renderHook(() => useAuth(), { wrapper: createWrapper() });

    let success: boolean = true;
    await act(async () => {
      success = await result.current.login({ email: 'test@example.com', password: 'wrong' });
    });

    expect(success).toBe(false);
    expect(result.current.error).toBe('認証失敗');
  });

  it('signup: サインアップ成功時にtrueを返す', async () => {
    mockedRepo.signup.mockResolvedValue(undefined);

    const { result } = renderHook(() => useAuth(), { wrapper: createWrapper() });

    let success: boolean = false;
    await act(async () => {
      success = await result.current.signup({ email: 'test@example.com', password: 'password', name: 'テスト' });
    });

    expect(success).toBe(true);
  });

  it('logout: ログアウト成功時にログインページに遷移する', async () => {
    mockedRepo.logout.mockResolvedValue(undefined);

    const { result } = renderHook(() => useAuth(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.logout();
    });

    expect(mockNavigate).toHaveBeenCalledWith('/login');
    expect(result.current.isAuthenticated).toBe(false);
  });

  it('getCurrentUser: ユーザー情報取得成功時に認証状態を設定する', async () => {
    const mockUser = { id: 1, email: 'test@example.com', name: 'テスト', sub: 'sub-123' };
    mockedRepo.getCurrentUser.mockResolvedValue(mockUser);

    const { result } = renderHook(() => useAuth(), { wrapper: createWrapper() });

    let user: any;
    await act(async () => {
      user = await result.current.getCurrentUser();
    });

    expect(user).toEqual(mockUser);
    expect(result.current.isAuthenticated).toBe(true);
  });

  it('refreshToken: リフレッシュ失敗時にログインページに遷移する', async () => {
    mockedRepo.refreshToken.mockRejectedValue(new Error('トークン期限切れ'));

    const { result } = renderHook(() => useAuth(), { wrapper: createWrapper() });

    let success: boolean = true;
    await act(async () => {
      success = await result.current.refreshToken();
    });

    expect(success).toBe(false);
    expect(mockNavigate).toHaveBeenCalledWith('/login');
  });
});
