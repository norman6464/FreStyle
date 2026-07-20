/**
 * コーディング演習（exercise）entity のドメイン型。
 */

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
