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
  createdAt?: string;
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
