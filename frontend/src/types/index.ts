/** チャットユーザー（チャットリスト表示用） */
export interface ChatUser {
  roomId: number;
  userId: number;
  name: string;
  email?: string;
  lastMessage?: string;
  lastMessageAt?: string;
  lastMessageSenderId?: number;
  unreadCount: number;
  profileImage?: string;
}

/** チャットメッセージ */
export interface ChatMessage {
  id: number;
  roomId: number;
  senderId: number;
  senderName?: string;
  content: string;
  createdAt?: string;
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

/** AIセッション */
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

/** AIメッセージ */
export interface AiMessage {
  id: number;
  sessionId: number;
  content: string;
  role: 'user' | 'assistant';
  createdAt?: string;
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

/** ユーザープロフィール（パーソナリティ設定） */
export interface UserProfile {
  displayName: string;
  selfIntroduction: string;
  communicationStyle: string;
  personalityTraits: string[];
  goals: string;
  concerns: string;
  preferredFeedbackStyle: string;
}

/** 認証ステート */
export interface AuthState {
  isAuthenticated: boolean;
  loading: boolean;
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
export interface Note {
  noteId: string;
  userId: number;
  title: string;
  content: string;
  isPinned: boolean;
  createdAt: number;
  updatedAt: number;
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
  overallScore: number;
  scores: AxisScore[];
  createdAt: string;
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
