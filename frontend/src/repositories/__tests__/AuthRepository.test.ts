import { describe, it, expect, vi, beforeEach } from 'vitest';
import authRepository from '../AuthRepository';
import apiClient from '../../lib/axios';

vi.mock('../../lib/axios');

const mockedApiClient = vi.mocked(apiClient);

describe('AuthRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('login: ログインできる', async () => {
    const mockUser = { id: 1, email: 'test@example.com', name: 'テスト', sub: 'sub-123' };
    mockedApiClient.post.mockResolvedValue({ data: mockUser });

    const result = await authRepository.login({ email: 'test@example.com', password: 'password123' });

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/auth/cognito/login', { email: 'test@example.com', password: 'password123' });
    expect(result).toEqual(mockUser);
  });

  it('signup: サインアップできる', async () => {
    mockedApiClient.post.mockResolvedValue({});

    await authRepository.signup({ email: 'test@example.com', password: 'password123', name: 'テスト' });

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/auth/cognito/signup', { email: 'test@example.com', password: 'password123', name: 'テスト' });
  });

  it('confirmSignup: サインアップ確認ができる', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { message: '確認成功' } });

    const result = await authRepository.confirmSignup({ email: 'test@example.com', code: '123456' });

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/auth/cognito/confirm', { email: 'test@example.com', code: '123456' });
    expect(result).toEqual({ message: '確認成功' });
  });

  it('callback: OAuthコールバックを処理できる', async () => {
    const mockData = { user: { id: 1, name: 'テスト' } };
    mockedApiClient.post.mockResolvedValue({ data: mockData });

    const result = await authRepository.callback('auth-code-123');

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/auth/cognito/callback', { code: 'auth-code-123' });
    expect(result).toEqual(mockData);
  });

  it('forgotPassword: パスワード再設定リクエストを送信できる', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { message: '確認コードを送信しました' } });

    const result = await authRepository.forgotPassword({ email: 'test@example.com' });

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/auth/cognito/forgot-password', { email: 'test@example.com' });
    expect(result).toEqual({ message: '確認コードを送信しました' });
  });

  it('confirmForgotPassword: パスワード再設定確認ができる', async () => {
    mockedApiClient.post.mockResolvedValue({ data: { message: 'パスワードをリセットしました' } });

    const result = await authRepository.confirmForgotPassword({
      email: 'test@example.com',
      confirmationCode: '123456',
      newPassword: 'newPassword123',
    });

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/auth/cognito/confirm-forgot-password', {
      email: 'test@example.com',
      confirmationCode: '123456',
      newPassword: 'newPassword123',
    });
    expect(result).toEqual({ message: 'パスワードをリセットしました' });
  });

  it('logout: ログアウトできる', async () => {
    mockedApiClient.post.mockResolvedValue({});

    await authRepository.logout();

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/auth/cognito/logout');
  });

  it('getCurrentUser: 現在のユーザー情報を取得できる', async () => {
    const mockUser = { id: 1, email: 'test@example.com', name: 'テスト', sub: 'sub-123' };
    mockedApiClient.get.mockResolvedValue({ data: mockUser });

    const result = await authRepository.getCurrentUser();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/auth/cognito/me');
    expect(result).toEqual(mockUser);
  });

  it('refreshToken: トークンリフレッシュできる', async () => {
    mockedApiClient.post.mockResolvedValue({});

    await authRepository.refreshToken();

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/auth/cognito/refresh-token');
  });
});
