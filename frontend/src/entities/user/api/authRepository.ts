import apiClient, { type PublicSafeRequestConfig } from '@/shared/api/axios';
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

export interface UserInfo {
  id: number;
  email?: string;
  name?: string;
  sub?: string;
  groups?: string[];
  /** /auth/me が返す表示名 */
  displayName?: string;
  /** 所属ワークスペースの UUID。未所属の運営ユーザーでは返らない。 */
  workspaceId?: string;
}

class AuthRepository {
  /**
   * 認可コードの交換。
   *
   * codeVerifier と nonce は、認可を始めたときにこのブラウザが作って手元に置いた値
   * （features/auth/lib/oidcAuthUrl）。バックエンドはこれを使って
   * 「この応答が、この人が始めた認可の応答か」を確かめる。
   */
  async callback(params: {
    code: string;
    codeVerifier: string;
    nonce: string;
  }): Promise<{ message: string }> {
    const body = {
      code: params.code,
      codeVerifier: params.codeVerifier,
      nonce: params.nonce,
    };
    // 交換失敗(401)も正常な応答として呼び出し側で扱う（ログイン画面へ戻して案内するため）。
    const config: PublicSafeRequestConfig = { skipAuthRedirect: true };
    const response = await apiClient.post(AUTH.callback, body, config);
    return response.data;
  }

  /**
   * ログアウト
   */
  async logout(): Promise<{ message: string; endSessionUrl?: string }> {
    const response = await apiClient.post(AUTH.logout);
    return response.data;
  }

  /**
   * 現在のユーザー情報取得
   */
  async getCurrentUser(): Promise<UserInfo> {
    const response = await apiClient.get(AUTH.me);
    return response.data;
  }

  /**
   * 公開ページ用の認証確認。
   *
   * getCurrentUser との違いは「未ログインでも /login へ飛ばさない」ことだけ。
   * 公開ページでは 401 が正常な答えなので、訪問者や検索エンジンのクローラを
   * ログイン画面へ追い出さないためにこちらを使う（FRESTYLE-225）。
   */
  async probeCurrentUser(): Promise<UserInfo> {
    const config: PublicSafeRequestConfig = { skipAuthRedirect: true };
    const response = await apiClient.get(AUTH.me, config);
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
