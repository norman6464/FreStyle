import { Link } from 'react-router-dom';
import { CodeBracketIcon } from '@heroicons/react/24/outline';
import LanguageIcon from '@/components/LanguageIcon';
import { useExerciseLanguageSummary, type ExerciseLanguageCard } from '@/hooks/useExerciseLanguageSummary';

/**
 * ExerciseLanguageSelectPage — `/code-editor` の入口（FRESTYLE-152）。
 *
 * 以前は全言語の問題を 1 画面に縦積みしていて見通しが悪かったため、
 * 「まず言語を選ぶ → その言語の問題一覧へ」の 2 段構成にした。
 * カードには言語ロゴ（Devicon）・問題数・完了数の進捗バーを出す。
 */
export default function ExerciseLanguageSelectPage() {
  const { cards, loading, error } = useExerciseLanguageSummary();

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-5xl mx-auto space-y-6">
      <header className="space-y-2">
        <div className="flex items-center gap-2 text-[var(--color-text-muted)]">
          <CodeBracketIcon className="w-5 h-5" />
          <span className="text-xs uppercase tracking-wider">Code Practice</span>
        </div>
        <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">コード学習</h1>
        <p className="text-sm text-[var(--color-text-tertiary)]">
          学びたい言語・ツールを選んでください。問題を解くと採点結果が履歴に残ります。
        </p>
      </header>

      {loading && <p className="text-sm text-[var(--color-text-muted)]">読み込み中...</p>}
      {error && <p className="text-sm text-red-500">{error}</p>}

      {!loading && !error && cards.length === 0 && (
        <p className="text-sm text-[var(--color-text-muted)]">公開されている問題がまだありません。</p>
      )}

      {!loading && !error && cards.length > 0 && (
        <ul className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {cards.map((card) => (
            <li key={card.language}>
              <LanguageCard card={card} />
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function LanguageCard({ card }: { card: ExerciseLanguageCard }) {
  const { language, label, total, solved } = card;
  const percent = total > 0 ? Math.round((solved / total) * 100) : 0;
  const completed = total > 0 && solved >= total;
  // 1 問でも解いていれば「続きから」。完了済みは「もう一度解く」。
  const actionLabel = completed ? 'もう一度解く' : solved > 0 ? '続きからはじめる' : 'はじめる';

  return (
    <Link
      to={`/code-editor/lang/${encodeURIComponent(language)}`}
      aria-label={`${label} の問題一覧へ（${total} 問中 ${solved} 問完了）`}
      className="group flex h-full flex-col gap-4 rounded-lg border border-surface-3 bg-surface-1 p-5 shadow-sm transition-colors hover:border-taupe-500/50 hover:bg-surface-2"
    >
      <div className="flex items-center gap-3">
        <LanguageIcon language={language} className="w-11 h-11 flex-shrink-0" />
        <div className="min-w-0">
          <h2 className="text-base font-semibold text-[var(--color-text-primary)] truncate">
            {label}
          </h2>
          <p className="text-xs text-[var(--color-text-muted)]">{total} 問</p>
        </div>
      </div>

      <div className="space-y-1.5">
        <div className="flex items-center justify-between text-xs">
          <span className="text-[var(--color-text-muted)]">
            {solved}/{total} 問完了
          </span>
          {completed && (
            <span className="font-semibold text-emerald-600">すべて完了</span>
          )}
        </div>
        {/* 進捗バー。完了済みは緑、進行中はブランド色で「今どこまで来たか」を一目で示す。 */}
        <div
          className="h-1.5 w-full overflow-hidden rounded-full bg-surface-3"
          role="progressbar"
          aria-valuenow={percent}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-label={`${label} の進捗`}
        >
          <div
            className={`h-full rounded-full transition-[width] ${
              completed ? 'bg-emerald-500' : 'bg-brand-500'
            }`}
            style={{ width: `${percent}%` }}
          />
        </div>
      </div>

      <span className="mt-auto inline-flex items-center justify-center rounded-lg border border-surface-3 px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition-colors group-hover:border-taupe-500/50 group-hover:bg-surface-1">
        {actionLabel}
      </span>
    </Link>
  );
}
