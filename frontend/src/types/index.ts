/**
 * ChatRoom は Go backend `domain.ChatRoom` と 1:1。
 */
export interface ChatRoom {
  id: number;
  name: string;
  isGroup: boolean;
  createdAt: string;
}

/**
 * ChatUser はチャット一覧画面（ChatListPage）専用の view モデル。
 * backend には対応する単一の domain は無く、複数 domain (ChatRoom + last ChatMessage + User)
 * を合成した形でフロントが扱う。
 */
export interface ChatUser {
  roomId: number;
  userId: number;
  name: string;
  email?: string;
  lastMessage?: string;
  lastMessageAt?: number;
  lastMessageSenderId?: number;
  unreadCount: number;
  profileImage?: string;
}

/**
 * ChatMessageDto は Go backend `domain.ChatMessage` と 1:1。
 * 新規実装はこの型を「真のソース」として参照する。Repository / hook で UI 表示用に
 * `ChatMessage` view へ写像する。
 */
export interface ChatMessageDto {
  roomId: number;
  messageId: string;
  senderId: number;
  content: string;
  createdAt: string;
}

/**
 * ChatMessage はチャット画面で表示するメッセージ view。
 * UI 用フィールド (`isSender`, `isDeleted`, `senderName`) を持ち、Hook 側で算出される。
 * 既存コードベースに合わせて `id: string` / `createdAt?: number` も残しているが、
 * 将来的に DTO へ統一する。
 */
export interface ChatMessage {
  id: string;
  roomId: number;
  senderId: number;
  senderName?: string;
  content: string;
  createdAt?: number;
  isSender: boolean;
  isDeleted?: boolean;
}

/** メンバーユーザー（ユーザー検索用） */
export interface MemberUser {
  id: number;
  name: string;
  email: string;
  roomId?: number;
}

/**
 * AiChatSession は Go backend `domain.AiChatSession` と 1:1。
 * 新規実装はこちらを参照すること。
 */
export interface AiChatSession {
  id: number;
  userId: number;
  title: string;
  sessionType: string;
  scenarioId?: number | null;
  createdAt: string;
  updatedAt: string;
}

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

/** 練習シナリオ */
export interface PracticeScenario {
  id: number;
  name: string;
  description: string;
  category: string;
  roleName: string;
  difficulty: string;
}

/**
 * AiChatMessageDto は Go backend `domain.AiChatMessage` と 1:1。
 * `messageId` (DynamoDB key) と `createdAt` (RFC3339 string) を持つ。
 */
export interface AiChatMessageDto {
  sessionId: number;
  messageId: string;
  role: 'user' | 'assistant';
  content: string;
  createdAt: string;
}

/**
 * AiMessage は AskAi 画面表示用の view 型。
 * `id` (string) を画面側で扱いやすい一意キーとして使う。
 */
export interface AiMessage {
  id: string;
  sessionId: number;
  content: string;
  role: 'user' | 'assistant';
  createdAt?: number;
  isSender?: boolean;
  isDeleted?: boolean;
}

/** 未読数更新通知 */
export interface UnreadUpdate {
  type: 'unread_update';
  roomId: number;
  increment: number;
}

/** フラッシュメッセージ */
export interface FlashMessage {
  type: 'success' | 'error';
  text: string;
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
}

/** 言い換え提案結果 */
export interface RephraseResult {
  formal: string;
  soft: string;
  concise: string;
  questioning: string;
  proposal: string;
}

/** 評価軸スコア */
export interface AxisScore {
  axis: string;
  score: number;
  comment: string;
}

/** スコアカード */
export interface ScoreCard {
  sessionId: number;
  scores: AxisScore[];
  overallScore: number;
}

/** SNSプロバイダー */
export type SnsProvider = 'google' | 'facebook' | 'x';

/** お気に入りフレーズ */
export interface FavoritePhrase {
  id: string;
  originalText: string;
  rephrasedText: string;
  pattern: string;
  createdAt: string;
}

/** 日次学習目標 */
export interface DailyGoal {
  date: string;
  target: number;
  completed: number;
}

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

/** スコア履歴 */
export interface ScoreHistory {
  sessionId: number;
  sessionTitle: string;
  overallScore: number;
  createdAt: string;
}

/** スコア履歴アイテム（スコア詳細付き） */
export interface ScoreHistoryItem {
  sessionId: number;
  sessionTitle: string;
  scenarioId: number | null;
  overallScore: number;
  scores: AxisScore[];
  createdAt: string;
}

/** 学習レポート */
export interface LearningReport {
  id: number;
  year: number;
  month: number;
  totalSessions: number;
  averageScore: number;
  previousAverageScore?: number;
  scoreChange?: number;
  bestAxis?: string;
  worstAxis?: string;
  practiceDays: number;
  createdAt?: string;
}

/** 通知 */
export interface Notification {
  id: number;
  type: string;
  title: string;
  message: string;
  isRead: boolean;
  relatedId?: number;
  createdAt: string;
}

export interface FriendshipUser {
  id: number;
  userId: number;
  username: string;
  iconUrl: string | null;
  bio: string | null;
  mutual: boolean;
  createdAt: string;
  status: string | null;
}

export interface FollowStatus {
  isFollowing: boolean;
  isFollowedBy: boolean;
  isMutual: boolean;
}

/** ランキングエントリー */
export interface RankingEntry {
  rank: number;
  userId: number;
  username: string;
  iconUrl: string | null;
  averageScore: number;
  sessionCount: number;
}

/** ランキング */
export interface Ranking {
  entries: RankingEntry[];
  myRanking: RankingEntry | null;
}

/** 会話テンプレート */
export interface ConversationTemplate {
  id: number;
  title: string;
  description: string;
  category: string;
  openingMessage: string;
  difficulty: string;
}

/** リマインダー設定 */
export interface ReminderSetting {
  enabled: boolean;
  reminderTime: string;
  daysOfWeek: string;
}

/** 共有セッション */
export interface SharedSession {
  id: number;
  sessionId: number;
  sessionTitle: string;
  userId: number;
  username: string;
  userIconUrl: string | null;
  description: string | null;
  createdAt: string;
}

/** ウィークリーチャレンジ */
export interface WeeklyChallenge {
  id: number;
  title: string;
  description: string;
  category: string;
  targetSessions: number;
  completedSessions: number;
  isCompleted: boolean;
  weekStart: string;
  weekEnd: string;
}
