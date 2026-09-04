import { describe, it, expect, vi, beforeEach } from 'vitest';
import authRepository from '../authRepository';
import apiClient from '@/shared/api/axios';

vi.mock('@/shared/api/axios');

const mockedApiClient = vi.mocked(apiClient);

describe('AuthRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('callback: PKCE の検証値と nonce を送る', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { message: 'ログインしました。' } });

    const result = await authRepository.callback({
      code: 'auth-code-123',
      codeVerifier: 'verifier-abc',
      nonce: 'nonce-xyz',
    });

    // 検証値を送らないと、公開クライアントでは交換そのものが通らない。
    // nonce はバックエンドが id_token の中身と突き合わせる。
    expect(mockedApiClient.post).toHaveBeenCalledWith(
      '/api/v2/auth/login',
      { code: 'auth-code-123', codeVerifier: 'verifier-abc', nonce: 'nonce-xyz' },
      { skipAuthRedirect: true },
    );
    expect(result).toEqual({ message: 'ログインしました。' });
  });

  it('logout: 発行者側のセッション終了先を受け取る', async () => {
    mockedApiClient.post.mockResolvedValue({
      data: { message: 'ログアウトしました。', endSessionUrl: 'https://issuer.test/logout' },
    });

    const result = await authRepository.logout();

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v2/auth/logout');
    // Cookie を消すだけでは発行者側のセッションが残り、同じ端末で入り直せてしまう。
    expect(result.endSessionUrl).toBe('https://issuer.test/logout');
  });

  it('getCurrentUser: 現在のユーザー情報を取得できる', async () => {
    const mockUser = { id: 1, email: 'test@example.com', name: 'テスト', sub: 'sub-123' };
    mockedApiClient.get.mockResolvedValue({ data: mockUser });

    const result = await authRepository.getCurrentUser();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v2/auth/me');
    expect(result).toEqual(mockUser);
  });

  it('refreshToken: トークンリフレッシュできる', async () => {
    mockedApiClient.post.mockResolvedValue({});

    await authRepository.refreshToken();

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v2/auth/refresh');
  });
});
