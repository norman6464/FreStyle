/**
 * ユーザー（user）entity のドメイン型。
 */

/**
 * Profile は Go backend `domain.ProfileView` と 1:1 で対応する。
 * users.display_name と profiles を合成した「プロフィール表示」用 DTO。
 */
export interface Profile {
  userId: number;
  displayName: string;
  /** ログインユーザのメールアドレス。 sidebar のユーザーメニューで表示する。 */
  email: string;
  bio: string;
  avatarUrl: string;
  status: string;
  updatedAt: string;
}

/**
 * User は Go backend `domain.User` と 1:1 で対応する。
 * 認証フローおよび admin 操作で利用する。
 */
export interface User {
  id: number;
  email: string;
  displayName: string;
  role: string;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string | null;
}

/** 認証ステート */
export interface AuthState {
  isAuthenticated: boolean;
  loading: boolean;
  isAdmin: boolean;
  /**
   * 現在ユーザーの role（'super_admin' / 'company_admin' / 'trainee'）。
   * メニュー出し分け（super_admin は管理機能のみ）と Protected の trainee 用ルート保護に使う。
   * 未認証 / 未確定は null。
   */
  role: string | null;
}

/**
 * users.role の取りうる値。
 * ルート側の認可ゲート（RequireRole）が「通過を許す role」の許可リストに使う。
 */
export type UserRole = 'super_admin' | 'company_admin' | 'trainee';
