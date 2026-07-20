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


/**
 * MasterExercise は Go backend `domain.MasterExercise` と 1:1。
 * 言語非依存の運営マスタ演習問題（旧 `PhpExercise` を汎用化）。
 *
 * `language` で 'php' / 'sql' / 'go' / 'javascript' を区別。`chapterId` は将来
 * PR-H1 で導入する chapters テーブルに紐付く章末演習で使う任意 FK。
 */
export interface MasterExercise {
  id: number;
  slug: string;
  language: string;
  orderIndex: number;
  category: string;
  title: string;
  description: string;
  starterCode: string;
  hintText: string;
  expectedOutput: string;
  /**
   * 採点モード。 'execute' (default) はサンドボックスでコードを実行し stdout を比較、
   * 'qa' はコード実行をせず提出文字列と expectedOutput を直接比較する。
   * docker / kubernetes など サンドボックス実行が困難な題材を Q&A 形式で扱うために導入。
   */
  mode: 'execute' | 'qa';
  /**
   * QA モードで 正解後に表示される markdown 解説。 execute モードでは未使用 (空文字)。
   */
  explanation: string;
  difficulty: number;
  isPublished: boolean;
  chapterId?: number | null;
  createdAt: string;
  updatedAt: string;
}

/** コード実行結果（Go backend `usecase.ExecuteCodeOutput` と 1:1）*/
export interface CodeExecutionResult {
  stdout: string;
  stderr: string;
  exitCode: number;
}

/**
 * 演習問題に紐付く入出力例（テストケース）の 1 ペア。
 * 詳細ページに「入力例 1 / 入力例 2 / ...」として描画され、
 * 採点 usecase（PR-W）では同じ全件がテストケースとして実行される。
 */
export interface MasterExerciseExample {
  id: number;
  exerciseId: number;
  orderIndex: number;
  inputText: string;
  expectedOutput: string;
  createdAt: string;
  updatedAt: string;
}

/** 詳細 API (`GET /api/v2/exercises/:slug`) のレスポンス。 */
export interface MasterExerciseDetail {
  exercise: MasterExercise;
  examples: MasterExerciseExample[];
}

/**
 * Course は教材を束ねる「コース（プロジェクト）」。 backend `domain.Course` と 1:1。
 *
 * 階層: Company 1 ── * Course 1 ── * TeachingMaterial
 *
 * - company_admin: 自社の draft 含む全件 list / 編集 / 削除可
 * - trainee: 自社の `isPublished=true` コースのみ閲覧可
 */
export interface Course {
  id: number;
  companyId: number;
  createdByUserId: number;
  title: string;
  description: string;
  /** 学習領域カテゴリ（空 = 未分類。constants/courseCategories の key と対応） */
  category: string;
  /** 主に扱う言語・技術（例: 'go' / 'docker'。空 = 言語が主題でない → バッジ非表示） */
  language: string;
  sortOrder: number;
  isPublished: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * CourseWithProgress はコース一覧 API (`GET /api/v2/courses`) の要素。
 * backend `usecase.CourseWithProgress` と 1:1(Course に章数と完了章数を合成したフラット JSON)。
 */
export interface CourseWithProgress extends Course {
  /** コース内の教材(章)数。trainee は published のみ、admin 系は下書き込み。 */
  materialCount: number;
  /** current user が完了した章数(現存する published 章のみ。常に materialCount 以下)。 */
  completedCount: number;
}

/**
 * TeachingMaterial は Go backend `domain.TeachingMaterial` と 1:1。
 * 必ず 1 つの Course に所属する Markdown 教材。
 *
 * - company_admin: 自社の draft 含む全件 list / 編集 / 削除可
 * - trainee: 自社の `isPublished=true` 教材かつ所属コース published のみ閲覧可
 */
export interface TeachingMaterial {
  id: number;
  companyId: number;
  courseId: number;
  createdByUserId: number;
  title: string;
  content: string;
  orderInCourse: number;
  isPublished: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * UserLessonProgress は trainee 自身の教材（レッスン）完了記録。 backend `domain.UserLessonProgress` と 1:1。
 * current user 固定で、 他人の進捗は取得・操作できない。
 */
export interface UserLessonProgress {
  id: number;
  userId: number;
  teachingMaterialId: number;
  courseId: number;
  completedAt: string;
  createdAt: string;
}

/** 一覧 API (`GET /api/v2/exercises`) で返る集計値（Go backend `repository.ExerciseSubmissionStats`）。 */
export interface ExerciseSubmissionStats {
  totalSubmissions: number;
  solvedUsers: number;
}

/**
 * 一覧 API のレスポンス 1 行。詳細 API (MasterExercise) より軽量で
 * description / starterCode / hintText / expectedOutput / explanation は含まない。
 * status: "solved" | "in_progress" | "" （未提出）。
 */
export interface MasterExerciseWithStatus {
  id: number;
  slug: string;
  language: string;
  orderIndex: number;
  category: string;
  title: string;
  difficulty: number;
  mode: 'execute' | 'qa';
  isPublished: boolean;
  status: '' | 'solved' | 'in_progress';
  stats: ExerciseSubmissionStats;
}

/** 演習問題一覧 API のページネーションレスポンス。 */
export interface ExercisePage {
  items: MasterExerciseWithStatus[];
  hasNext: boolean;
  offset: number;
  limit: number;
}

/**
 * 言語別の演習集計（`GET /api/v2/exercises/summary`）。
 * 言語選択カードの進捗表示に使う（FRESTYLE-152）。solved は current user が正解済みの問題数。
 */
export interface ExerciseLanguageSummary {
  language: string;
  total: number;
  solved: number;
}

/** 提出 API (`POST /api/v2/exercises/:slug/submit`) のレスポンス 1 件あたりの採点結果。 */
export interface ExerciseTestCaseResult {
  orderIndex: number;
  input: string;
  expectedOutput: string;
  actualOutput: string;
  stderr: string;
  passed: boolean;
}

/** 提出 API のレスポンス。 */
export interface ExerciseSubmitResult {
  submissionId: number;
  isCorrect: boolean;
  results: ExerciseTestCaseResult[];
}

/** 提出履歴 1 行（Go backend `domain.ExerciseSubmission`）。 */
export interface ExerciseSubmission {
  id: number;
  userId: number;
  exerciseKind: 'master' | 'company';
  exerciseId: number;
  submittedCode: string;
  stdout: string;
  stderr: string;
  exitCode: number;
  isCorrect: boolean;
  submittedAt: string;
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

/** 章閲覧記録（`GET /api/v2/me/dashboard` の recentChapterViews 要素）。 */
export interface UserChapterView {
  userId: number;
  teachingMaterialId: number;
  courseId: number;
  firstViewedAt: string;
  lastViewedAt: string;
  viewCount: number;
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
