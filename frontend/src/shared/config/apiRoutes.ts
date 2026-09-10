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

/**
 * 認証（Bearer の ID トークン検証）
 *
 * backend はセッション用の Cookie を発行しない。login は Authorization: Bearer で
 * 渡した ID トークンを検証して users 行と個人ワークスペースを作る（自己サインアップ、
 * 既存ユーザーなら実質 no-op）。ログアウト・更新はどちらも発行者
 * （GCIP のクライアント SDK / ローカルの Dex）側だけで完結し、backend には対応する
 * エンドポイントが無い。
 */
export const AUTH = {
  login: `${API_V2}/auth/login`,
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

/** 画像アップロード（current user 名義の S3 PUT 署名 URL。リッチテキストエディタ全般で共有）*/
export const IMAGES = {
  /** POST /api/v2/rich-text/images/upload-url — {contentType} → {url, key, publicUrl} */
  uploadUrl: `${API_V2}/rich-text/images/upload-url`,
} as const;

export const RANKING = `${API_V2}/ranking` as const;

/** 練習モード（シナリオ / セッション / ブックマーク / 共有セッション） */
export const PRACTICE = {
  scenarios: `${API_V2}/practice/scenarios`,
  scenario: (scenarioId: number | string) => `${API_V2}/practice/scenarios/${scenarioId}`,
  sessions: `${API_V2}/practice/sessions`,
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

/** 設定（リマインダー・週次チャレンジ） */
export const REMINDER = `${API_V2}/reminder` as const;

export const WEEKLY_CHALLENGE = {
  current: `${API_V2}/weekly-challenge`,
  progress: `${API_V2}/weekly-challenge/progress`,
} as const;

/** 管理者ダッシュボード（会社 / 招待 / シナリオ） */
export const ADMIN = {
  members: `${API_V2}/admin/members`,
  /** PATCH /api/v2/admin/members/:userId/active — 従業員アカウントの有効/無効 */
  memberActive: (userId: number | string) => `${API_V2}/admin/members/${userId}/active`,
  /** DELETE /api/v2/admin/members/:userId — 従業員の論理削除 */
  member: (userId: number | string) => `${API_V2}/admin/members/${userId}`,
  invitations: `${API_V2}/admin/invitations`,
  invitationById: (id: number | string) => `${API_V2}/admin/invitations/${id}`,
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

/**
 * ナレッジ（workspaces → spaces → pages の木）。
 *
 * 旧リッチ文書（/api/v2/documents）の後継。あちらは所有者スコープの平らな一覧だったが、
 * 退役済み（移送なしで撤去）。こちらは付与（grant）だけで解決する木（打ち消す層は持たない）。
 *
 * ワークスペースは URL の slug で指す（内部 UUID は外に出さない）。slug から所属を確定する
 * middleware を backend 側の group が通しているので、slug を含まないパスは一覧と作成だけ。
 */
export const KB_API = {
  /** GET(所属一覧) / POST(作成) — /api/v2/kb/workspaces */
  workspaces: `${API_V2}/kb/workspaces`,
  /** DELETE(削除) — /api/v2/kb/workspaces/:slug。配下ごと消える。会社のものは消せない */
  workspace: (workspaceSlug: string) => `${API_V2}/kb/workspaces/${workspaceSlug}`,
  /** GET(一覧) / POST(作成) — /api/v2/kb/workspaces/:slug/spaces。一覧は見えるものだけ返る */
  spaces: (workspaceSlug: string) => `${API_V2}/kb/workspaces/${workspaceSlug}/spaces`,
  /**
   * GET(ツリー) / POST(作成) — /api/v2/kb/workspaces/:slug/spaces/:spaceId/pages
   *
   * ツリーは閲覧できるページだけを返し、見えない親の配下は現れない。
   * 代わりに各段の hasHiddenChildren に「見えないページが在るか」の有無だけが入る
   * （枚数も題名も返らない）。
   */
  pages: (workspaceSlug: string, spaceId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/pages`,
  /** GET(本文込み) / PATCH(改名) — /api/v2/kb/workspaces/:slug/pages/:pageId */
  page: (workspaceSlug: string, pageId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}`,
  /** PUT(設定) / DELETE(解除) — /api/v2/kb/workspaces/:slug/pages/:pageId/icon */
  pageIcon: (workspaceSlug: string, pageId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/icon`,
  /**
   * POST — /api/v2/kb/workspaces/:slug/pages/:pageId/images/upload-url
   *
   * ページ本文・カバー画像向けの S3 PUT 署名 URL を発行する（current user 名義）。
   * body は {contentType, size}。durable な保存形式は応答の key そのもの
   * （"kb/<workspaceId>/<pageId>/<epochNs>.bin"）で、doc にはこの key を保存する
   * （presigned URL は期限があるので doc に書き込まない）。
   */
  pageImageUploadUrl: (workspaceSlug: string, pageId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/images/upload-url`,
  /**
   * GET — /api/v2/kb/workspaces/:slug/pages/:pageId/images/download-url?key=…
   *
   * doc に保存された key を表示用の期限付き URL へ解決する。存在しない/参照されていない
   * key は 404 になり得る。
   */
  pageImageDownloadUrl: (workspaceSlug: string, pageId: string, key: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/images/download-url?key=${encodeURIComponent(key)}`,
  /** PUT(設定) / DELETE(解除) — /api/v2/kb/workspaces/:slug/pages/:pageId/cover */
  pageCover: (workspaceSlug: string, pageId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/cover`,
  /**
   * GET — /api/v2/kb/pages/:pageId
   *
   * /kb/{pageId} の URL からの解決。URL にワークスペースを出さないための口で、
   * 応答の workspaceSlug を以降の呼び出し（木・保存）に使う。
   */
  resolvePage: (pageId: string) => `${API_V2}/kb/pages/${pageId}`,
  /** PUT(本文の置き換え) — /api/v2/kb/workspaces/:slug/pages/:pageId/content */
  pageContent: (workspaceSlug: string, pageId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/content`,
  /** PATCH(表示名の変更) — /api/v2/kb/workspaces/:slug/spaces/:spaceId。key は変えられない */
  space: (workspaceSlug: string, spaceId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}`,
  /**
   * GET — /api/v2/kb/workspaces/:slug/search?q=
   *
   * ワークスペース全体の題名検索。返るのは閲覧できる現役ページだけで、
   * 判定はツリーと同じ規則をサーバーが持つ（検索だけ別の判定にしない）。
   */
  search: (workspaceSlug: string) => `${API_V2}/kb/workspaces/${workspaceSlug}/search`,
  /**
   * GET — /api/v2/kb/workspaces/:slug/pages/:pageId/backlinks
   *
   * このページを参照しているページの一覧（逆リンク）。応答は KbPage[] と同じ形
   * （追加フィールドなし）。見える範囲の判定は木・検索と同じ規則をサーバーが持つ。
   */
  pageBacklinks: (workspaceSlug: string, pageId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/backlinks`,
  /**
   * GET(一覧) — /api/v2/kb/workspaces/:slug/pages/:pageId/grants
   *
   * **返るのはそのページ自身に張った行だけ**で、上の段（ワークスペース / スペース /
   * 祖先のページ）から届いている相手は含まない。空 = 誰も見られない、ではない。
   */
  pageGrants: (workspaceSlug: string, pageId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/grants`,
  /** PUT(付与) / DELETE(取り消し) — 同じ 1 行を指す（DB の主キーと同じ形） */
  pageGrant: (workspaceSlug: string, pageId: string, principalId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/grants/${principalId}`,
  /**
   * GET — /api/v2/kb/workspaces/:slug/pages/:pageId/principals
   *
   * 権限を張れる相手を表示名つきで返す（相手選び用）。中身はワークスペース全体だが、
   * 呼べるかはページ単位で決まる。
   */
  pagePrincipals: (workspaceSlug: string, pageId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/principals`,
  /**
   * GET(一覧) / POST(作成) — /api/v2/kb/workspaces/:slug/pages/:pageId/comment-threads
   *
   * 一覧は作成日時昇順で、各スレッドは comments 配列を持つ。作成（POST）の応答も
   * 同じ形（comments に作った最初の 1 件が入ったスレッド 1 件）。
   */
  commentThreads: (workspaceSlug: string, pageId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/comment-threads`,
  /** POST(返信の追加) — /api/v2/kb/workspaces/:slug/pages/:pageId/comment-threads/:threadId/comments */
  comments: (workspaceSlug: string, pageId: string, threadId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/comment-threads/${threadId}/comments`,
  /** POST(解決) — .../comment-threads/:threadId/resolve。更新後のスレッドを返す */
  resolveCommentThread: (workspaceSlug: string, pageId: string, threadId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/comment-threads/${threadId}/resolve`,
  /** POST(再開) — .../comment-threads/:threadId/reopen。更新後のスレッドを返す */
  reopenCommentThread: (workspaceSlug: string, pageId: string, threadId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/comment-threads/${threadId}/reopen`,
  /**
   * GET(一覧・新しい順) / POST(明示的な版の作成) — /api/v2/kb/workspaces/:slug/pages/:pageId/versions
   *
   * 一覧・単体取得は閲覧できれば誰でもできる（canView）。作成（版を残す）は編集権限が要る。
   */
  pageVersions: (workspaceSlug: string, pageId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/versions`,
  /** GET(1件・doc込み) — /api/v2/kb/workspaces/:slug/pages/:pageId/versions/:seq */
  pageVersion: (workspaceSlug: string, pageId: string, seq: number) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/versions/${seq}`,
  /**
   * POST(復元・body無し) — .../versions/:seq/restore。編集権限が要る。
   * 応答は本文保存（PUT .../content）と同じ形（KbPageContentSaveResult）。
   */
  restorePageVersion: (workspaceSlug: string, pageId: string, seq: number) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/versions/${seq}/restore`,
  /**
   * GET(一覧・?spaceId= は任意) — /api/v2/kb/workspaces/:slug/templates
   *
   * 一覧・使用はワークスペース所属者なら誰でもできる。spaceId を渡すと、そのスペース専用の
   * テンプレート + ワークスペース全体のテンプレートの両方が返る想定（backend 未実装の段階の
   * 想定であり確定ではない — 要すり合わせ）。
   */
  templates: (workspaceSlug: string) => `${API_V2}/kb/workspaces/${workspaceSlug}/templates`,
  /**
   * POST(作成) — /api/v2/kb/workspaces/:slug/pages/:pageId/templates
   *
   * 今のページの内容からテンプレートを作る。ワークスペースの編集者（editor）以上が要る。
   */
  pageTemplates: (workspaceSlug: string, pageId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/templates`,
  /** DELETE(削除) — /api/v2/kb/workspaces/:slug/templates/:templateId。編集者以上が要る。 */
  template: (workspaceSlug: string, templateId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/templates/${templateId}`,
  /**
   * POST(テンプレートからページを作成) — /api/v2/kb/workspaces/:slug/spaces/:spaceId/pages/from-template
   *
   * 応答は通常のページ作成（POST .../pages）と同じ形（KbPage）。
   */
  pageFromTemplate: (workspaceSlug: string, spaceId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/pages/from-template`,
  /**
   * GET(openな一覧) / POST(作成) — /api/v2/kb/workspaces/:slug/pages/:pageId/suggestions
   *
   * 作成は CanComment、一覧の閲覧は CanView。一覧は open な提案だけを返す
   * （採用・却下が済んだものは含まない）。
   */
  pageSuggestions: (workspaceSlug: string, pageId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/suggestions`,
  /**
   * POST(採用・body無し) — .../suggestions/:suggestionId/accept。CanEdit が要る。
   * 本文へ反映し版を1つ切る。応答は反映後の提案そのもの（doc が反映後の本文と同じ）。
   */
  acceptPageSuggestion: (workspaceSlug: string, pageId: string, suggestionId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/suggestions/${suggestionId}/accept`,
  /** POST(却下・body無し) — .../suggestions/:suggestionId/reject。CanEdit が要る。本文は一切変えない。 */
  rejectPageSuggestion: (workspaceSlug: string, pageId: string, suggestionId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/pages/${pageId}/suggestions/${suggestionId}/reject`,
} as const;

// WebSocket は SSE への置換で廃止 (PR-D, 2026-05-07)。

/**
 * チケット・バックログ（既存の spaces に属する。routes_ticket.go 参照）。
 *
 * KB_API と同じくワークスペースは URL の slug で指す。チケットはページのような個票の
 * grant を持たず、実効権限は常にスペース単位（設計 Ⅳ-H）なので、grants / principals 系の
 * ルートは無い（担当者候補の名前解決は KB_API.pagePrincipals を流用する。Ⅳ-G 参照）。
 */
export const TICKET_API = {
  /** POST — /api/v2/kb/workspaces/:slug/spaces/:spaceId/tickets/enable。body は省略可 */
  enable: (workspaceSlug: string, spaceId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/tickets/enable`,
  /**
   * GET(一覧) / POST(作成) — /api/v2/kb/workspaces/:slug/spaces/:spaceId/tickets
   *
   * 一覧のクエリは statusId / typeId / assigneePrincipalId（いずれも省略可）と
   * archived（'true' でアーカイブだけを返す。省略時は現役だけ。「込み」は取れない）。
   */
  tickets: (workspaceSlug: string, spaceId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/tickets`,
  /** GET — /api/v2/kb/workspaces/:slug/tickets/by-key/:key（例 FRESTYLE-12） */
  ticketByKey: (workspaceSlug: string, key: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/tickets/by-key/${encodeURIComponent(key)}`,
  /** GET(取得) / PUT(全置換) — /api/v2/kb/workspaces/:slug/tickets/:ticketId */
  ticket: (workspaceSlug: string, ticketId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/tickets/${ticketId}`,
  /**
   * GET — /api/v2/kb/tickets/:ticketId
   *
   * /kb/tickets/{ticketId} の URL からの解決。ワークスペースを出さないための口で、
   * 応答の workspaceSlug を以降の呼び出しに使う（KB_API.resolvePage と同じ役割）。
   */
  resolveTicket: (ticketId: string) => `${API_V2}/kb/tickets/${ticketId}`,
  /** POST(並び替え・204) — .../tickets/:ticketId/move。anchorTicketId 省略で末尾へ */
  moveTicket: (workspaceSlug: string, ticketId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/tickets/${ticketId}/move`,
  /** POST(アーカイブ) — .../tickets/:ticketId/archive */
  archiveTicket: (workspaceSlug: string, ticketId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/tickets/${ticketId}/archive`,
  /** POST(復元) — .../tickets/:ticketId/restore */
  restoreTicket: (workspaceSlug: string, ticketId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/tickets/${ticketId}/restore`,
  /** POST — .../tickets/:ticketId/status。resolution は category=done のときだけ意味を持つ */
  changeTicketStatus: (workspaceSlug: string, ticketId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/tickets/${ticketId}/status`,
  /** PUT — .../tickets/:ticketId/parent。parentId 省略でトップレベルへ */
  changeTicketParent: (workspaceSlug: string, ticketId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/tickets/${ticketId}/parent`,
  /** PUT(設定) / DELETE(解除・204) — .../tickets/:ticketId/assignee */
  ticketAssignee: (workspaceSlug: string, ticketId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/tickets/${ticketId}/assignee`,
  /** GET — .../tickets/:ticketId/history（変更履歴・新しい順） */
  ticketHistory: (workspaceSlug: string, ticketId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/tickets/${ticketId}/history`,
  /** GET(一覧・古い順) / POST(投稿) — .../tickets/:ticketId/comments */
  ticketComments: (workspaceSlug: string, ticketId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/tickets/${ticketId}/comments`,
  /** PUT(本文の置換) / DELETE(削除・204) — .../comments/:commentId。投稿者本人以外は 403 */
  ticketComment: (workspaceSlug: string, ticketId: string, commentId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/tickets/${ticketId}/comments/${commentId}`,
  /** GET — .../comments/:commentId/edits（編集前の本文・新しい順） */
  ticketCommentEdits: (workspaceSlug: string, ticketId: string, commentId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/tickets/${ticketId}/comments/${commentId}/edits`,
  /**
   * PUT(付ける) / DELETE(外す) — .../comments/:commentId/reactions/:emoji。どちらも 204・冪等。
   *
   * 絵文字は URL の一部なので必ず encodeURIComponent を通す（生のままだと多バイト文字で経路が壊れる）。
   */
  ticketCommentReaction: (workspaceSlug: string, ticketId: string, commentId: string, emoji: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/tickets/${ticketId}/comments/${commentId}/reactions/${encodeURIComponent(emoji)}`,
  /**
   * GET(一覧) / POST(作成) — /api/v2/kb/workspaces/:slug/spaces/:spaceId/ticket-statuses
   *
   * 一覧の各行は activeTicketCount（現役チケットでの使用数）を持つ（管理画面の「使用中 N 件」）。
   */
  ticketStatuses: (workspaceSlug: string, spaceId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/ticket-statuses`,
  /** PUT — .../ticket-statuses/:statusId（name / color / category をまとめて置換） */
  ticketStatus: (workspaceSlug: string, spaceId: string, statusId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/ticket-statuses/${statusId}`,
  /** POST(body 無し) — .../ticket-statuses/:statusId/set-initial */
  setInitialTicketStatus: (workspaceSlug: string, spaceId: string, statusId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/ticket-statuses/${statusId}/set-initial`,
  /** POST(body 無し) — .../ticket-statuses/:statusId/archive。使用中は 409 status_in_use */
  archiveTicketStatus: (workspaceSlug: string, spaceId: string, statusId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/ticket-statuses/${statusId}/archive`,
  /** POST(body 無し) — .../ticket-statuses/:statusId/restore */
  restoreTicketStatus: (workspaceSlug: string, spaceId: string, statusId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/ticket-statuses/${statusId}/restore`,
  /** GET(一覧) / POST(作成) — /api/v2/kb/workspaces/:slug/spaces/:spaceId/ticket-types */
  ticketTypes: (workspaceSlug: string, spaceId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/ticket-types`,
  /** PUT — .../ticket-types/:typeId */
  ticketType: (workspaceSlug: string, spaceId: string, typeId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/ticket-types/${typeId}`,
  /** POST(body 無し) — .../ticket-types/:typeId/set-default */
  setDefaultTicketType: (workspaceSlug: string, spaceId: string, typeId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/ticket-types/${typeId}/set-default`,
  /** POST(body 無し) — .../ticket-types/:typeId/archive。使用中は 409 type_in_use */
  archiveTicketType: (workspaceSlug: string, spaceId: string, typeId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/ticket-types/${typeId}/archive`,
  /** POST(body 無し) — .../ticket-types/:typeId/restore */
  restoreTicketType: (workspaceSlug: string, spaceId: string, typeId: string) =>
    `${API_V2}/kb/workspaces/${workspaceSlug}/spaces/${spaceId}/ticket-types/${typeId}/restore`,
} as const;
