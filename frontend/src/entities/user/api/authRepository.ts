import apiClient from '@/shared/api/axios';
import { AUTH } from '@/shared/config/apiRoutes';

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
  email?: string;
  name?: string;
  sub?: string;
  groups?: string[];
  isAdmin?: boolean;
  /** バックエンド users テーブルの role: super_admin / company_admin / trainee */
  role?: string;
  /** 所属する company の ID。SuperAdmin は null になり得る。 */
  companyId?: number | null;
  /** /auth/me が返す表示名（招待時に displayName として登録された値） */
  displayName?: string;
  /** 自社が trainee の AI エージェント機能を有効にしているか。 */
  aiChatEnabledForTrainees?: boolean;
}

class AuthRepository {
  /**
   * ログイン
   */
  async login(request: LoginRequest): Promise<UserInfo> {
    const response = await apiClient.post(AUTH.login, request);
    return response.data;
  }

  /**
   * OAuth コールバック。
   *
   * invitationToken は招待マジックリンク経由のサインアップで sessionStorage から
   * 引き渡される UUID。指定がある場合 backend は email より優先して照合する。
   * 未指定（直接 /login から入った既存ユーザー等）は省略可。
   */
  async callback(code: string, invitationToken?: string | null): Promise<{ success: string }> {
    const body: { code: string; invitationToken?: string } = { code };
    if (invitationToken) body.invitationToken = invitationToken;
    const response = await apiClient.post(AUTH.callback, body);
    return response.data;
  }

  /**
   * パスワード再設定リクエスト
   */
  async forgotPassword(request: ForgotPasswordRequest): Promise<{ message: string }> {
    const response = await apiClient.post(AUTH.forgotPassword, request);
    return response.data;
  }

  /**
   * パスワード再設定確認
   */
  async confirmForgotPassword(request: ConfirmForgotPasswordRequest): Promise<{ message: string }> {
    const response = await apiClient.post(AUTH.confirmForgotPassword, request);
    return response.data;
  }

  /**
   * ログアウト
   */
  async logout(): Promise<void> {
    await apiClient.post(AUTH.logout);
  }

  /**
   * 現在のユーザー情報取得
   */
  async getCurrentUser(): Promise<UserInfo> {
    const response = await apiClient.get(AUTH.me);
    return response.data;
  }

  /**
   * トークンリフレッシュ
   */
  async refreshToken(): Promise<void> {
    await apiClient.post(AUTH.refreshToken);
  }
}

export default new AuthRepository();
