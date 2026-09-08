import { describe, it, expect, vi, beforeEach } from 'vitest';
import authRepository from '../authRepository';
import apiClient from '@/shared/api/axios';

vi.mock('@/shared/api/axios');

const mockedApiClient = vi.mocked(apiClient);

describe('AuthRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('login: /auth/login をBodyなしで呼ぶ（Bearerヘッダはaxios側で付く）', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { message: 'ログインしました。' } });

    const result = await authRepository.login();

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v2/auth/login');
    expect(result).toEqual({ message: 'ログインしました。' });
  });

  it('getCurrentUser: 現在のユーザー情報を取得できる', async () => {
    const mockUser = { id: 1, email: 'test@example.com', name: 'テスト', sub: 'sub-123' };
    mockedApiClient.get.mockResolvedValue({ data: mockUser });

    const result = await authRepository.getCurrentUser();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v2/auth/me');
    expect(result).toEqual(mockUser);
  });

  it('probeCurrentUser: skipAuthRedirect を付けて呼ぶ（未ログインでも/loginへ飛ばさない）', async () => {
    const mockUser = { id: 1, email: 'test@example.com' };
    mockedApiClient.get.mockResolvedValue({ data: mockUser });

    const result = await authRepository.probeCurrentUser();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v2/auth/me', { skipAuthRedirect: true });
    expect(result).toEqual(mockUser);
  });
});
