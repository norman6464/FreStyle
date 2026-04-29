/**
 * API ルート定義の単一ソース。
 *
 * 設計方針:
 * - フロント側 repository から呼び出す Go backend のエンドポイント URL を
 *   1 ファイルに集約する（旧実装は 25 repository に 166 箇所ハードコード）
 * - パラメータを取るルートは pure な関数 (`(id: number) => string`) として export
 * - パラメータ無しのルートは `as const` の string literal
 * - prefix `/api/v2` は `API_V2` として共通化し、Go backend 移行に伴う
 *   v2 → v3 のような大規模変更があった場合に 1 行で切り替え可能にする
 *
 * 追加ルール:
 * - 新規エンドポイントは backend `routes_*.go` への追加と同時にここに足す
 * - フロント実装は repository 経由で参照し、page / hook / component から
 *   直接 API パスを書かない
 *
 * Go backend 側との対応は `backend/internal/handler/router.go` 系を参照。
 */

const API_V2 = '/api/v2' as const;

/** 認証 (Cognito Hosted UI / SRP / 自己情報取得) */
export const AUTH = {
  login: `${API_V2}/auth/cognito/login`,
  signup: `${API_V2}/auth/cognito/signup`,
  confirm: `${API_V2}/auth/cognito/confirm`,
  callback: `${API_V2}/auth/cognito/callback`,
  forgotPassword: `${API_V2}/auth/cognito/forgot-password`,
  confirmForgotPassword: `${API_V2}/auth/cognito/confirm-forgot-password`,
  logout: `${API_V2}/auth/cognito/logout`,
  refreshToken: `${API_V2}/auth/cognito/refresh-token`,
  me: `${API_V2}/auth/me`,
} as const;

/** プロフィール / アイコン画像 / 統計 */
export const PROFILE = {
  me: `${API_V2}/profile/me`,
  meUpdate: `${API_V2}/profile/me/update`,
  meImagePresignedUrl: `${API_V2}/profile/me/image/presigned-url`,
  /** GET /users/me/stats — 自分の使い方統計 */
  meStats: `${API_V2}/users/me/stats`,
} as const;

/** AI チャット（セッション・メッセージ） */
export const AI_CHAT = {
  sessions: `${API_V2}/ai-chat/sessions`,
  session: (sessionId: number | string) => `${API_V2}/ai-chat/sessions/${sessionId}`,
  sessionMessages: (sessionId: number | string) =>
    `${API_V2}/ai-chat/sessions/${sessionId}/messages`,
} as const;

/** ユーザー間チャット */
export const CHAT = {
  rooms: `${API_V2}/chat/rooms`,
  roomRead: (roomId: number | string) => `${API_V2}/chat/rooms/${roomId}/read`,
  users: `${API_V2}/chat/users`,
  userCreate: (userId: number | string) => `${API_V2}/chat/users/${userId}/create`,
  userHistory: (roomId: number | string) => `${API_V2}/chat/users/${roomId}/history`,
  aiRephrase: `${API_V2}/chat/ai/rephrase`,
} as const;

/** Note CRUD と画像 presigned upload */
export const NOTES = {
  list: `${API_V2}/notes`,
  byId: (noteId: number | string) => `${API_V2}/notes/${noteId}`,
  imagesPresignedUrl: (noteId: number | string) =>
    `${API_V2}/notes/${noteId}/images/presigned-url`,
} as const;

/** SessionNote (セッション固有ノート) */
export const SESSION_NOTES = {
  byId: (sessionId: number | string) => `${API_V2}/session-notes/${sessionId}`,
} as const;

/** スコア / トレンド / ランキング / ゴール / カード / 学習レポート */
export const SCORES = {
  sessionScore: (sessionId: number | string) => `${API_V2}/scores/sessions/${sessionId}`,
  cards: `${API_V2}/score-cards`,
  goals: `${API_V2}/score-goals`,
} as const;

export const RANKING = `${API_V2}/ranking` as const;

export const LEARNING_REPORTS = {
  list: `${API_V2}/learning-reports`,
  generate: `${API_V2}/learning-reports/generate`,
  yearMonth: (year: number, month: number) =>
    `${API_V2}/learning-reports/${year}/${month}`,
} as const;

/** 練習モード（シナリオ / セッション / ブックマーク / 共有セッション） */
export const PRACTICE = {
  scenarios: `${API_V2}/practice/scenarios`,
  scenario: (scenarioId: number | string) => `${API_V2}/practice/scenarios/${scenarioId}`,
  sessions: `${API_V2}/practice/sessions`,
} as const;

export const SCENARIO_BOOKMARKS = {
  list: `${API_V2}/scenario-bookmarks`,
  byScenario: (scenarioId: number | string) =>
    `${API_V2}/scenario-bookmarks/${scenarioId}`,
} as const;

export const SHARED_SESSIONS = {
  list: `${API_V2}/shared-sessions`,
  byId: (sessionId: number | string) => `${API_V2}/shared-sessions/${sessionId}`,
} as const;

/** 会話テンプレート / お気に入りフレーズ */
export const TEMPLATES = {
  list: `${API_V2}/templates`,
  byId: (id: number | string) => `${API_V2}/templates/${id}`,
} as const;

export const FAVORITE_PHRASES = {
  list: `${API_V2}/favorite-phrases`,
  byId: (id: number | string) => `${API_V2}/favorite-phrases/${id}`,
} as const;

/** Friendship + フォロー */
export const FRIENDSHIPS = {
  following: `${API_V2}/friendships/following`,
  followers: `${API_V2}/friendships/followers`,
  follow: (userId: number | string) => `${API_V2}/friendships/${userId}/follow`,
  status: (userId: number | string) => `${API_V2}/friendships/${userId}/status`,
} as const;

/** 通知 */
export const NOTIFICATIONS = {
  list: `${API_V2}/notifications`,
  unreadCount: `${API_V2}/notifications/unread-count`,
  read: (notificationId: number | string) =>
    `${API_V2}/notifications/${notificationId}/read`,
  readAll: `${API_V2}/notifications/read-all`,
} as const;

/** 設定（リマインダー・日次目標・週次チャレンジ） */
export const REMINDER = `${API_V2}/reminder` as const;

export const DAILY_GOALS = {
  today: `${API_V2}/daily-goals/today`,
  target: `${API_V2}/daily-goals/target`,
  increment: `${API_V2}/daily-goals/increment`,
  streak: `${API_V2}/daily-goals/streak`,
} as const;

export const WEEKLY_CHALLENGE = {
  current: `${API_V2}/weekly-challenge`,
  progress: `${API_V2}/weekly-challenge/progress`,
} as const;

/** 管理者ダッシュボード（招待 / シナリオ） */
export const ADMIN = {
  invitations: `${API_V2}/admin/invitations`,
  invitationById: (id: number | string) => `${API_V2}/admin/invitations/${id}`,
  scenarios: `${API_V2}/admin/scenarios`,
  scenarioById: (id: number | string) => `${API_V2}/admin/scenarios/${id}`,
} as const;

/** 外部 URL の OGP / oEmbed メタ情報を取得するプロキシ */
export const EMBEDS = {
  oembed: `${API_V2}/embeds/oembed`,
} as const;

/** WebSocket（Cookie 認証 / 通常 ws:// or wss:// にフロント側で書き換え）*/
export const WS = {
  /** ルームごとブロードキャスト（ユーザー間チャット） */
  chatRoom: (roomId: number | string) => `${API_V2}/ws/chat/${roomId}`,
  /** AI チャット（Bedrock 連携） */
  aiChat: `${API_V2}/ws/ai-chat`,
} as const;
