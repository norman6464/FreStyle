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

/** 練習シナリオ（フロント表示用 view）。
 *  backend 1:1 は `PracticeScenarioDto` を参照すること。 */
export interface PracticeScenario {
  id: number;
  name: string;
  description: string;
  category: string;
  roleName: string;
  difficulty: string;
}

/** PracticeScenarioDto は Go backend `domain.PracticeScenario` と 1:1。 */
export interface PracticeScenarioDto {
  id: number;
  title: string;
  description: string;
  category: string;
  difficultyLevel: number;
  isActive: boolean;
  createdAt: string;
}

/** ScenarioBookmark は Go backend `domain.ScenarioBookmark` と 1:1。 */
export interface ScenarioBookmark {
  id: number;
  userId: number;
  scenarioId: number;
  createdAt: string;
}

/**
 * AiChatMessageDto は Go backend `domain.AiChatMessage` と 1:1。
 * `messageId` (DynamoDB key) と `createdAt` (RFC3339 string) を持つ。
 * `attachments` はユーザー発話に紐付く画像 / ドキュメント（PR-G1: 画像のみ）。
 */
export interface AiChatMessageDto {
  sessionId: number;
  messageId: string;
  role: 'user' | 'assistant';
  content: string;
  attachments?: AiAttachment[];
  createdAt: string;
}

/**
 * AiAttachmentKind は backend `domain.AttachmentKind*` 定数 と 1:1。
 * PR-G1 では "image" のみ実用。"document" は PR-G2 で PDF / CSV 対応のため予約。
 */
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
  /**
   * 現在ユーザーの role（'super_admin' / 'company_admin' / 'trainee'）。
   * メニュー出し分け（super_admin は管理機能のみ）と Protected の trainee 用ルート保護に使う。
   * 未認証 / 未確定は null。
   */
  role: string | null;
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

/** スコアカード（フロント表示用 view）。
 *  backend 1:1 は `ScoreCardDto` を参照すること。 */
export interface ScoreCard {
  sessionId: number;
  scores: AxisScore[];
  overallScore: number;
}

/**
 * ScoreCardDto は Go backend `domain.ScoreCard` と 1:1。
 * 5 軸スコアは個別カラム（logicalScore / considerationScore / summaryScore /
 * proposalScore / listeningScore）として保持される。
 */
export interface ScoreCardDto {
  id: number;
  userId: number;
  sessionId: number;
  overallScore: number;
  logicalScore: number;
  considerationScore: number;
  summaryScore: number;
  proposalScore: number;
  listeningScore: number;
  feedback: string;
  createdAt: string;
}

/** ScoreGoalDto は Go backend `domain.ScoreGoal` と 1:1。 */
export interface ScoreGoalDto {
  userId: number;
  targetScore: number;
  updatedAt: string;
}

/** ScoreTrendPoint / ScoreTrend は Go backend `domain.ScoreTrend` と 1:1。 */
export interface ScoreTrendPoint {
  date: string;
  overallScore: number;
}
export interface ScoreTrend {
  userId: number;
  points: ScoreTrendPoint[];
}

/** RankingEntryDto は Go backend `domain.RankingEntry` と 1:1。
 *  既存 `RankingEntry` (UI view) は username / iconUrl / sessionCount を
 *  別 API から取得して合成しているため、当面 view 型は別に保持する。 */
export interface RankingEntryDto {
  userId: number;
  displayName: string;
  averageScore: number;
  rank: number;
}

/** SNSプロバイダー */
export type SnsProvider = 'google' | 'facebook' | 'x';

/** お気に入りフレーズ（フロント表示用 view）。
 *  backend 1:1 は `FavoritePhraseDto` を参照すること。 */
export interface FavoritePhrase {
  id: string;
  originalText: string;
  rephrasedText: string;
  pattern: string;
  createdAt: string;
}

/** FavoritePhraseDto は Go backend `domain.FavoritePhrase` と 1:1。 */
export interface FavoritePhraseDto {
  id: number;
  userId: number;
  phrase: string;
  note: string;
  createdAt: string;
}

/** 日次学習目標（フロント表示用 view）。
 *  backend 1:1 は `DailyGoalDto` を参照すること。 */
export interface DailyGoal {
  date: string;
  target: number;
  completed: number;
}

/** DailyGoalDto は Go backend `domain.DailyGoal` と 1:1。
 *  date は YYYY-MM-DD で、targetMinutes / actualMinutes / isAchieved を持つ。 */
export interface DailyGoalDto {
  id: number;
  userId: number;
  date: string;
  targetMinutes: number;
  actualMinutes: number;
  isAchieved: boolean;
  createdAt: string;
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

/** 学習レポート（フロント表示用 view）。
 *  backend 1:1 は `LearningReportDto` を参照すること。 */
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

/** LearningReportDto は Go backend `domain.LearningReport` と 1:1。
 *  非同期生成のため status (`pending` / `ready` / `failed`) と s3Key を持つ。 */
export interface LearningReportDto {
  id: number;
  userId: number;
  periodFrom: string;
  periodTo: string;
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

/** NotificationDto は Go backend `domain.Notification` と 1:1。
 *  backend は `body` カラム、フロント view は `message` を使う点が差。 */
export interface NotificationDto {
  id: number;
  userId: number;
  type: string;
  title: string;
  body: string;
  isRead: boolean;
  createdAt: string;
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

/** 会話テンプレート（フロント表示用 view）。
 *  backend 1:1 は `ConversationTemplateDto` を参照すること。 */
export interface ConversationTemplate {
  id: number;
  title: string;
  description: string;
  category: string;
  openingMessage: string;
  difficulty: string;
}

/** ConversationTemplateDto は Go backend `domain.ConversationTemplate` と 1:1。 */
export interface ConversationTemplateDto {
  id: number;
  title: string;
  body: string;
  category: string;
  isActive: boolean;
  createdAt: string;
}

/** リマインダー設定（フロント表示用 view）。 */
export interface ReminderSetting {
  enabled: boolean;
  reminderTime: string;
  daysOfWeek: string;
}

/** ReminderSettingDto は Go backend `domain.ReminderSetting` と 1:1（user 単位の通知設定）。 */
export interface ReminderSettingDto {
  userId: number;
  enabled: boolean;
  reminderTime: string;
  daysOfWeek: string;
  updatedAt: string;
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

/** ウィークリーチャレンジ（フロント表示用 view）。
 *  backend 1:1 は `WeeklyChallengeDto` / `WeeklyChallengeProgressDto` を参照。 */
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

/** WeeklyChallengeDto は Go backend `domain.WeeklyChallenge` と 1:1。 */
export interface WeeklyChallengeDto {
  id: number;
  weekStart: string;
  title: string;
  description: string;
  isActive: boolean;
  createdAt: string;
}

/** WeeklyChallengeProgressDto は Go backend `domain.WeeklyChallengeProgress` と 1:1。 */
export interface WeeklyChallengeProgressDto {
  userId: number;
  challengeId: number;
  completed: boolean;
  updatedAt: string;
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
  sortOrder: number;
  isPublished: boolean;
  createdAt: string;
  updatedAt: string;
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

/** 一覧 API (`GET /api/v2/exercises`) で返る集計値（Go backend `repository.ExerciseSubmissionStats`）。 */
export interface ExerciseSubmissionStats {
  totalSubmissions: number;
  solvedUsers: number;
}

/** 一覧 API のレスポンス 1 行。 MasterExercise + current user 状態 + 全体集計。
 *  status: "solved" | "in_progress" | "" （未提出）。 */
export interface MasterExerciseWithStatus extends MasterExercise {
  status: '' | 'solved' | 'in_progress';
  stats: ExerciseSubmissionStats;
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
