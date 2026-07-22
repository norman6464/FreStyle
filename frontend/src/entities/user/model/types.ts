/**
 * ユーザー（user）entity のドメイン型。
 *
 * UserDashboard / UserDailyActivity は独立した業務エンティティではなく
 * 「そのユーザーの学習集計」という read model なのでここに含める。
 */

import type { UserChapterView } from '@/entities/course/@x/user';

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
  cognitoSub: string;
  email: string;
  displayName: string;
  companyId?: number | null;
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
  /**
   * 自社が trainee の AI エージェント機能を有効にしているか（/auth/me 由来）。
   * trainee のサイドバー「AI」表示判定に使う。未取得 / 会社未所属 / 管理者は true。
   */
  aiChatEnabledForTrainees: boolean;
}

/** SNSプロバイダー */
export type SnsProvider = 'google' | 'facebook' | 'x';

/** 日次学習活動サマリー（`GET /api/v2/me/dashboard` の recentActivity 要素）。 */
export interface UserDailyActivity {
  userId: number;
  activityDate: string; // ISO 8601 date e.g. "2026-06-19T00:00:00Z"
  exerciseCount: number;
  correctCount: number;
  lessonCount: number;
  aiChatCount: number;
  noteCount: number;
}

/** `GET /api/v2/me/dashboard` のレスポンス全体。 */
export interface UserDashboard {
  streak: number;
  totalExercises: number;
  totalCorrect: number;
  totalLessons: number;
  recentActivity: UserDailyActivity[];
  recentChapterViews: UserChapterView[];
}
