/*
 * 未移行の型の置き場（FSD 移行中の暫定ファイル）。
 * Phase 5b-2 で残りの entity へ振り分けて、このファイルは削除する。
 */
import type { UserChapterView } from '@/entities/course';

/**
 * AiSession は AskAi 画面表示用の view 型。userId / updatedAt は不要なケースが多く、
 * scene のような UI 由来のフィールドを足してある。後段で AiChatSession へ統一予定。
 */
export interface AiSession {
  id: number;
  title?: string;
  scene?: string;
  sessionType?: string;
  scenarioId?: number;
  createdAt?: string;
}

/** AiAttachmentKind は添付が画像かドキュメントかの区分（表示の出し分けに使う）。 */
export type AiAttachmentKind = 'image' | 'document';

/**
 * AiAttachmentFormat は AWS Bedrock Converse の image/document format に渡す短い文字列。
 * バリデーションは backend 側 (`usecase.AllowedAttachmentContentTypes`) が一次情報。
 */
export type AiAttachmentFormat =
  | 'png'
  | 'jpeg'
  | 'gif'
  | 'webp'
  | 'pdf'
  | 'csv'
  | 'txt';

/**
 * AiAttachment は AI チャットメッセージに添付された画像 / ドキュメントの参照。
 * 実体は S3 にあり `key` で一意特定する。`previewUrl` は送信前のローカルプレビュー
 * （Object URL）を保持する用途で、バックエンドへは送らない。
 */
export interface AiAttachment {
  key: string;
  filename: string;
  contentType: string;
  kind: AiAttachmentKind;
  format: AiAttachmentFormat;
  sizeBytes: number;
  /** ローカル状態用: ブラウザ内 preview URL（送信完了後は破棄） */
  previewUrl?: string;
}

/**
 * AiMessage は AskAi 画面表示用の view 型。
 * `id` (string) を画面側で扱いやすい一意キーとして使う。
 * `createdAt` は ISO 8601 文字列（backend の SSE / REST 双方が ISO で返す）。
 */
export interface AiMessage {
  id: string;
  sessionId: number;
  content: string;
  role: 'user' | 'assistant';
  attachments?: AiAttachment[];
  createdAt?: string;
  isSender?: boolean;
  isDeleted?: boolean;
  /**
   * ストリーミング placeholder 由来のクライアント側 ID(FRESTYLE-146)。
   * done で id がサーバ確定値に差し替わっても React の key をこれで安定させ、
   * バブルの remount(ペーシングの残り放出が全文ジャンプになる)を防ぐ。
   */
  clientId?: string;
}

/** フォームメッセージ */
export interface FormMessage {
  type: 'success' | 'error';
  text: string;
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

/** セッションメモ */
export interface SessionNote {
  sessionId: number;
  note: string;
  updatedAt: string;
}

/** ノート */
/**
 * Note は Go backend `domain.Note` と 1:1 で対応する。
 * - id / userId は number
 * - createdAt / updatedAt は RFC3339 string（ISO）
 * - isPublic = 公開フラグ、isPinned = ピン留めフラグ（独立した属性）
 */
export interface Note {
  id: number;
  userId: number;
  title: string;
  content: string;
  isPublic: boolean;
  isPinned: boolean;
  createdAt: string;
  updatedAt: string;
}

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

/** 学習レポート。Go backend `domain.LearningReport` と 1:1。
 *  非同期生成のため status (`pending` / `ready` / `failed`) と s3Key を持つ。 */
export interface LearningReport {
  id: number;
  userId: number;
  /** 対象期間の開始（月初、ISO 文字列）。表示の対象月はここから導出する。 */
  periodFrom: string;
  /** 対象期間の終了（翌月初、半開区間）。 */
  periodTo: string;
  /** 生成状態。pending = 作成中 / ready = 完成 / failed = 失敗。 */
  status: 'pending' | 'ready' | 'failed';
  s3Key?: string;
  createdAt: string;
}

/** 通知（フロント表示用 view）。
 *  backend 1:1 は `NotificationDto` を参照すること。 */
export interface Notification {
  id: number;
  type: string;
  title: string;
  message: string;
  isRead: boolean;
  relatedId?: number;
  createdAt: string;
}


// ─── ダッシュボード ──────────────────────────────────────────────────────────

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
