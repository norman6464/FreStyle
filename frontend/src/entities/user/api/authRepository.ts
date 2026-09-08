import apiClient, { type PublicSafeRequestConfig } from '@/shared/api/axios';
import { AUTH } from '@/shared/config/apiRoutes';

/**
 * 認証リポジトリ
 *
 * <p>役割:</p>
 * <ul>
 *   <li>認証関連のAPI呼び出しを抽象化</li>
 *   <li>セッション確立、自己情報取得</li>
 * </ul>
 *
 * <p>インフラ層（Infrastructure Layer）:</p>
 * <ul>
 *   <li>外部APIとの通信を担当</li>
 *   <li>Domain層に依存せず、独立している</li>
 * </ul>
 *
 * backend はセッション用の Cookie を発行しない。認証は毎リクエスト
 * `Authorization: Bearer <ID トークン>` で行う（`shared/api/axios.ts` が自動で付ける）。
 * ID トークンの取得元（GCIP のクライアント SDK / ローカルの Dex）はこのリポジトリの
 * 関知するところではない。
 */

export interface UserInfo {
  id: number;
  email?: string;
  name?: string;
  sub?: string;
  /** /auth/me が返す表示名 */
  displayName?: string;
  /** 所属ワークスペースの UUID。未所属の運営ユーザーでは返らない。 */
  workspaceId?: string;
}

class AuthRepository {
  /**
   * セッションを確立する。
   *
   * サインイン/サインアップ直後に一度呼ぶ。backend が Bearer の ID トークンを検証し、
   * 初回なら users 行と個人ワークスペースを作る（自己サインアップ）。既存ユーザーなら
   * 実質 no-op（何度呼んでも安全）。
   */
  async login(): Promise<{ message: string }> {
    const response = await apiClient.post(AUTH.login);
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
   * ログイン画面へ追い出さないためにこちらを使う。
   */
  async probeCurrentUser(): Promise<UserInfo> {
    const config: PublicSafeRequestConfig = { skipAuthRedirect: true };
    const response = await apiClient.get(AUTH.me, config);
    return response.data;
  }
}

export default new AuthRepository();
