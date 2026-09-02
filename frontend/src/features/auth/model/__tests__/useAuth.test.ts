import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import { createElement } from 'react';
import { useAuth } from '../useAuth';
import { AuthRepository } from '@/entities/user';
import authReducer from '@/entities/user/model/authSlice';

vi.mock('@/entities/user/api/authRepository');

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

  // 発行者側のセッション終了先が返ってきたら、そちらへ飛ぶ。
  // 手元の Cookie を消すだけだと発行者にログイン済みが残り、同じ端末で入り直せてしまう。
  it('logout: 発行者のセッション終了先が返れば そちらへ遷移する', async () => {
    mockedRepo.logout.mockResolvedValue({
      message: 'ログアウトしました。',
      endSessionUrl: 'https://issuer.test/oidc/v1/end_session',
    });
    const original = window.location;
    // location.href への代入を観測する（jsdom は実際の遷移をしない）。
    const assigned: string[] = [];
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: {
        ...original,
        set href(v: string) {
          assigned.push(v);
        },
        get href() {
          return assigned[assigned.length - 1] ?? '';
        },
      },
    });

    const { result } = renderHook(() => useAuth(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.logout();
    });

    Object.defineProperty(window, 'location', { configurable: true, value: original });

    expect(assigned).toContain('https://issuer.test/oidc/v1/end_session');
    // 発行者へ飛ばすので、SPA 内の /login へは送らない（二重遷移になる）。
    expect(mockNavigate).not.toHaveBeenCalledWith('/login');
  });

  it('logout: ログアウト成功時にログインページに遷移する', async () => {
    mockedRepo.logout.mockResolvedValue({ message: 'ログアウトしました。' });

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

  it('getCurrentUser: ユーザー情報取得失敗時にnullを返す', async () => {
    mockedRepo.getCurrentUser.mockRejectedValue(new Error('セッション切れ'));

    const { result } = renderHook(() => useAuth(), { wrapper: createWrapper() });

    let user: any;
    await act(async () => {
      user = await result.current.getCurrentUser();
    });

    expect(user).toBeNull();
    expect(result.current.error).toBe('ユーザー情報の取得に失敗しました。');
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
