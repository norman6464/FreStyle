import { DocumentTextIcon } from '@heroicons/react/24/outline';
import BackLink from './BackLink';
import ResultBadge from './ResultBadge';
import LanguageBadge from '@/shared/ui/LanguageBadge';
import { MasterExercise, ExerciseSubmitResult } from '../../types';

interface Props {
  exercise: MasterExercise;
  submitResult: ExerciseSubmitResult | null;
}

/**
 * 演習詳細 ヘッダー — BackLink + タイトル + 難易度 + 採点バッジ を 1 まとめ にする。
 *
 * 元 ExerciseDetailPage の 主ビュー と QaExerciseView で 同じ markup が 重複 して いた のを 統合。
 * バッジ は submitResult が ある とき のみ 表示。
 */
export default function ExerciseHeader({ exercise: ex, submitResult }: Props) {
  return (
    <header className="space-y-3">
      <BackLink />
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0 w-10 h-10 rounded-md bg-taupe-500/10 border border-taupe-500/30 flex items-center justify-center">
          <DocumentTextIcon className="w-5 h-5 text-taupe-400" />
        </div>
        <div className="min-w-0 flex-1">
          <h1 className="text-xl font-bold text-[var(--color-text-primary)] leading-tight">
            {ex.title}
          </h1>
          <div className="mt-1 flex items-center gap-2 text-xs text-[var(--color-text-muted)]">
            <LanguageBadge language={ex.language} />
            <span>難易度 {'★'.repeat(Math.max(1, Math.min(5, ex.difficulty)))}</span>
            <span>#{ex.orderIndex}</span>
          </div>
        </div>
        {submitResult && <ResultBadge isCorrect={submitResult.isCorrect} />}
      </div>
    </header>
  );
}
