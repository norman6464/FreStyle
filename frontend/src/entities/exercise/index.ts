/*
 * entities/exercise の Public API。
 *
 * 外から使ってよいものだけを名前付きで re-export する（FSD 公式仕様。`export *` は使わない）。
 * この Slice の内部ファイルを直接 import してはいけない。
 */

export { default as ExerciseRepository } from './api/exerciseRepository';

export type {
  MasterExercise,
  MasterExerciseExample,
  MasterExerciseDetail,
  MasterExerciseWithStatus,
  ExerciseSubmissionStats,
  ExercisePage,
  ExerciseLanguageSummary,
  ExerciseTestCaseResult,
  ExerciseSubmitResult,
  ExerciseSubmission,
  CodeExecutionResult,
} from './model/types';

export { default as BackLink } from './ui/BackLink';
export { default as ExampleBlock } from './ui/ExampleBlock';
export { default as ExecutionResultTable } from './ui/ExecutionResultTable';
export { default as ExerciseHeader } from './ui/ExerciseHeader';
export { default as QaExerciseView } from './ui/QaExerciseView';
export { default as ResultBadge } from './ui/ResultBadge';
export { default as SubmissionRow } from './ui/SubmissionRow';
export { default as SubmitResultPanel } from './ui/SubmitResultPanel';

export { monacoLanguageOf, pad } from './lib/exerciseFormat';
export { parseErrorLines } from './lib/executionErrors';
export type { ExecutionErrorMarker } from './lib/executionErrors';

export { EXERCISE_LANGUAGES } from './config/exerciseLanguages';
export type { ExerciseLanguageDef } from './config/exerciseLanguages';
