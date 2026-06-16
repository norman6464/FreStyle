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
  // OAuth Hosted UI ログイン(認可コード→token 交換)/ logout / refresh は
  // provider 非依存の REST パスに統一(/auth/login, /auth/logout, /auth/refresh)。
  callback: `${API_V2}/auth/login`,
  forgotPassword: `${API_V2}/auth/cognito/forgot-password`,
  confirmForgotPassword: `${API_V2}/auth/cognito/confirm-forgot-password`,
  logout: `${API_V2}/auth/logout`,
  refreshToken: `${API_V2}/auth/refresh`,
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

/** AI チャット（セッション CRUD + SSE ストリーミング + 添付ファイル） */
export const AI_CHAT = {
  sessions: `${API_V2}/ai-chat/sessions`,
  session: (sessionId: number | string) => `${API_V2}/ai-chat/sessions/${sessionId}`,
  sessionMessages: (sessionId: number | string) =>
    `${API_V2}/ai-chat/sessions/${sessionId}/messages`,
  /** POST /ai-chat/stream — SSE ストリーミング送信 */
  stream: `${API_V2}/ai-chat/stream`,
  /** POST /ai-chat/attachments/upload-url — 添付ファイル PUT 用 presigned URL */
  attachmentUploadUrl: `${API_V2}/ai-chat/attachments/upload-url`,
} as const;

/** Note CRUD と画像 presigned upload */
export const NOTES = {
  list: `${API_V2}/notes`,
  byId: (noteId: number | string) => `${API_V2}/notes/${noteId}`,
  imagesPresignedUrl: (noteId: number | string) =>
    `${API_V2}/notes/${noteId}/images/presigned-url`,
} as const;

/** 画像アップロード（current user 名義の S3 PUT 署名 URL。ノート/教材で共有）*/
export const IMAGES = {
  /** POST /api/v2/notes/images/upload-url — {contentType} → {url, key, publicUrl} */
  uploadUrl: `${API_V2}/notes/images/upload-url`,
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

/** 管理者ダッシュボード（会社 / 招待 / シナリオ） */
export const ADMIN = {
  companies: `${API_V2}/admin/companies`,
  /** GET /api/v2/admin/companies/stats — 各社のメンバー集計付き会社横断ビュー（super_admin） */
  companiesStats: `${API_V2}/admin/companies/stats`,
  /** PATCH /api/v2/admin/companies/:id/active — 会社アカウントの有効/無効（super_admin） */
  companyActive: (id: number | string) => `${API_V2}/admin/companies/${id}/active`,
  members: `${API_V2}/admin/members`,
  memberAiAccess: (userId: number | string) => `${API_V2}/admin/members/${userId}/ai-access`,
  /** PATCH /api/v2/admin/members/:userId/active — 従業員アカウントの有効/無効 */
  memberActive: (userId: number | string) => `${API_V2}/admin/members/${userId}/active`,
  /** DELETE /api/v2/admin/members/:userId — 従業員の論理削除 */
  member: (userId: number | string) => `${API_V2}/admin/members/${userId}`,
  invitations: `${API_V2}/admin/invitations`,
  invitationById: (id: number | string) => `${API_V2}/admin/invitations/${id}`,
  /** GET /api/v2/admin/audit-events — 監査ログ一覧（super_admin） */
  auditEvents: `${API_V2}/admin/audit-events`,
  scenarios: `${API_V2}/admin/scenarios`,
  scenarioById: (id: number | string) => `${API_V2}/admin/scenarios/${id}`,
} as const;

/** 招待マジックリンク受諾フロー（認証不要） */
export const INVITATIONS = {
  validateToken: (token: string) =>
    `${API_V2}/invitations/accept/${encodeURIComponent(token)}`,
} as const;

/** 外部 URL の OGP / oEmbed メタ情報を取得するプロキシ */
export const EMBEDS = {
  oembed: `${API_V2}/embeds/oembed`,
} as const;

/** マスタ演習問題（旧 PHP 専用 API を言語非依存に汎用化）+ コード実行 */
export const EXERCISES = {
  list: `${API_V2}/exercises`,
  bySlug: (slug: string) => `${API_V2}/exercises/${encodeURIComponent(slug)}`,
  submit: (slug: string) => `${API_V2}/exercises/${encodeURIComponent(slug)}/submit`,
  submissions: (slug: string) => `${API_V2}/exercises/${encodeURIComponent(slug)}/submissions`,
} as const;

export const CODE = {
  execute: `${API_V2}/code/execute`,
  warmup: `${API_V2}/code/warmup`,
} as const;

/** コース（company_admin が作成、 自社 trainee + 同社 admin が閲覧）*/
export const COURSES = {
  list: `${API_V2}/courses`,
  byId: (id: number | string) => `${API_V2}/courses/${id}`,
  /** GET /api/v2/courses/:id/materials — コース内教材一覧 */
  materials: (id: number | string) => `${API_V2}/courses/${id}/materials`,
} as const;

/** 教材 個別 CRUD（コース配下）*/
export const TEACHING_MATERIALS = {
  byId: (id: number | string) => `${API_V2}/teaching-materials/${id}`,
  /** POST /api/v2/teaching-materials — body の courseId 必須 */
  create: `${API_V2}/teaching-materials`,
} as const;

/** 学習進捗（trainee 自身の教材完了状態。 current user 固定で userId は受け取らない）*/
export const LESSON_PROGRESS = {
  /** GET /api/v2/lesson-progress — 自分の完了レッスン一覧 */
  list: `${API_V2}/lesson-progress`,
  /** POST /api/v2/lesson-progress — body の teachingMaterialId を完了として記録 */
  complete: `${API_V2}/lesson-progress`,
  /** DELETE /api/v2/lesson-progress/:teachingMaterialId — 完了を取り消す */
  incomplete: (id: number | string) => `${API_V2}/lesson-progress/${id}`,
} as const;

/** 企業利用申請（公開フォーム → super_admin 通知）*/
export const COMPANY_APPLICATIONS = {
  /** POST /api/v2/company-applications — 認証不要の申請作成 */
  create: `${API_V2}/company-applications`,
  /** GET /api/v2/admin/company-applications — super_admin 専用一覧 */
  adminList: `${API_V2}/admin/company-applications`,
  /** PATCH /api/v2/admin/company-applications/:id/status — status 更新 */
  adminUpdateStatus: (id: number | string) =>
    `${API_V2}/admin/company-applications/${id}/status`,
} as const;

/** 会社設定（company_admin / super_admin が自社の設定を取得・更新）*/
export const COMPANY_SETTINGS = {
  /** GET / PUT /api/v2/company/settings */
  base: `${API_V2}/company/settings`,
} as const;

// WebSocket は SSE (AI_CHAT.stream) への置換で廃止 (PR-D, 2026-05-07)。
