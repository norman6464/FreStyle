import apiClient from '../lib/axios';

/**
 * 認証リポジトリ
 *
 * <p>役割:</p>
 * <ul>
 *   <li>認証関連のAPI呼び出しを抽象化</li>
 *   <li>ログイン、ログアウト、ユーザー情報取得</li>
 * </ul>
 *
 * <p>インフラ層（Infrastructure Layer）:</p>
 * <ul>
 *   <li>外部APIとの通信を担当</li>
 *   <li>Domain層に依存せず、独立している</li>
 * </ul>
 */

export interface LoginRequest {
  email: string;
  password: string;
}

export interface SignupRequest {
  email: string;
  password: string;
  name: string;
}

export interface ConfirmSignupRequest {
  email: string;
  code: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ConfirmForgotPasswordRequest {
  email: string;
  confirmationCode: string;
  newPassword: string;
}

export interface UserInfo {
  id: number;
  email: string;
  name: string;
  sub: string;
}

class AuthRepository {
  /**
   * ログイン
   */
  async login(request: LoginRequest): Promise<UserInfo> {
    const response = await apiClient.post('/api/auth/cognito/login', request);
    return response.data;
  }

  /**
   * サインアップ
   */
  async signup(request: SignupRequest): Promise<void> {
    await apiClient.post('/api/auth/cognito/signup', request);
  }

  /**
   * サインアップ確認
   */
  async confirmSignup(request: ConfirmSignupRequest): Promise<{ message: string }> {
    const response = await apiClient.post('/api/auth/cognito/confirm', request);
    return response.data;
  }

  /**
   * OAuthコールバック
   */
  async callback(code: string): Promise<{ success: string }> {
    const response = await apiClient.post('/api/auth/cognito/callback', { code });
    return response.data;
  }

  /**
   * パスワード再設定リクエスト
   */
  async forgotPassword(request: ForgotPasswordRequest): Promise<{ message: string }> {
    const response = await apiClient.post('/api/auth/cognito/forgot-password', request);
    return response.data;
  }

  /**
   * パスワード再設定確認
   */
  async confirmForgotPassword(request: ConfirmForgotPasswordRequest): Promise<{ message: string }> {
    const response = await apiClient.post('/api/auth/cognito/confirm-forgot-password', request);
    return response.data;
  }

  /**
   * ログアウト
   */
  async logout(): Promise<void> {
    await apiClient.post('/api/auth/cognito/logout');
  }

  /**
   * 現在のユーザー情報取得
   */
  async getCurrentUser(): Promise<UserInfo> {
    const response = await apiClient.get('/api/auth/cognito/me');
    return response.data;
  }

  /**
   * トークンリフレッシュ
   */
  async refreshToken(): Promise<void> {
    await apiClient.post('/api/auth/cognito/refresh-token');
  }
}

export default new AuthRepository();
