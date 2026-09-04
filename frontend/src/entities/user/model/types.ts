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
 * 認証フローで利用する。
 */
export interface User {
  id: number;
  email: string;
  displayName: string;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string | null;
}

/** 認証ステート */
export interface AuthState {
  isAuthenticated: boolean;
  loading: boolean;
}
