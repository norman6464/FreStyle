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

export interface LoginRequest {
  email: string;
  password: string;
}

/** loginWithChallenge の結果。通常成功か、初回パスワード設定が必要か。 */
export type LoginOutcome =
  | { kind: 'success' }
  | { kind: 'new_password_required'; session: string };

/** 一時パスワード初回ログインの新パスワード設定リクエスト。 */
export interface NewPasswordRequest {
  email: string;
  session: string;
  newPassword: string;
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
  /** /auth/me が返す表示名（招待時に displayName として登録された値） */
  displayName?: string;
}

class AuthRepository {
  /**
   * ログイン
   *
   * 401(認証情報が違う)は正常な応答なので、共通インターセプターの
   * 「401 → トークンリフレッシュ → 失敗ならログイン画面へ強制遷移」を無効にする。
   * これがないと画面が再読み込みされ、呼び出し側のエラーメッセージが表示されない。
   */
  async login(request: LoginRequest): Promise<UserInfo> {
    const config: PublicSafeRequestConfig = { skipAuthRedirect: true };
    const response = await apiClient.post(AUTH.login, request, config);
    return response.data;
  }

  /**
   * メール / パスワードでログインし、結果を判別して返す。
   *
   * 一時パスワードでの初回ログインは backend が 200 で
   * `{ challenge: 'NEW_PASSWORD_REQUIRED', session }` を返す（トークンはまだ発行されない）。
   * その場合は session を持ち帰り、submitNewPassword で新パスワードを設定する。
   * それ以外（通常成功）は Cookie が発行済みなので success を返す。
   */
  async loginWithChallenge(request: LoginRequest): Promise<LoginOutcome> {
    const config: PublicSafeRequestConfig = { skipAuthRedirect: true };
    const response = await apiClient.post(AUTH.login, request, config);
    const data = response.data as { challenge?: string; session?: string };
    if (data?.challenge === 'NEW_PASSWORD_REQUIRED' && data.session) {
      return { kind: 'new_password_required', session: data.session };
    }
    return { kind: 'success' };
  }

  /**
   * NEW_PASSWORD_REQUIRED チャレンジに新パスワードで応答する（初回ログインの本人設定）。
   * 成功で認証 Cookie が発行される。session 失効は 401、パスワードポリシー違反は 400。
   */
  async submitNewPassword(request: NewPasswordRequest): Promise<void> {
    const config: PublicSafeRequestConfig = { skipAuthRedirect: true };
    await apiClient.post(AUTH.newPassword, request, config);
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
    // 認可コードの交換失敗(401)も正常な応答として呼び出し側で扱う(login と同じ理由)。
    const config: PublicSafeRequestConfig = { skipAuthRedirect: true };
    const response = await apiClient.post(AUTH.callback, body, config);
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
