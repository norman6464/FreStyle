import { Link } from 'react-router-dom';
import { CheckCircleIcon, ClockIcon, CodeBracketIcon } from '@heroicons/react/24/outline';
import { useExerciseList } from '../hooks/useExerciseList';
import { MasterExerciseWithStatus } from '../types';

/**
 * ExerciseListPage — `/code-editor` のリスト画面。
 *
 * - カード形式で全演習問題を一覧表示
 * - 各カードに current user のステータス（解いた / 取り組み中 / 未着手）と
 *   全体集計（提出数 / 正答ユーザ数）を表示
 * - クリックで詳細ページ `/code-editor/:slug` へ遷移
 *
 * 言語フィルタは UI 上で切替可能。 PHP のみ運用中の現時点では実用性は薄いが、
 * 多言語化を見据えて入口だけ用意する。
 */
export default function ExerciseListPage() {
  const { language, setLanguage, exercises, categories, loading, error } = useExerciseList();

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-5xl mx-auto space-y-6">
      <header className="space-y-2">
        <div className="flex items-center gap-2 text-[var(--color-text-muted)]">
          <CodeBracketIcon className="w-5 h-5" />
          <span className="text-xs uppercase tracking-wider">Code Practice</span>
        </div>
        <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">コード学習</h1>
        <p className="text-sm text-[var(--color-text-tertiary)]">
          解きたい問題を選んでください。提出すると採点結果が履歴に残ります。
        </p>
      </header>

      <div className="flex items-center gap-2 text-sm">
        <label className="text-[var(--color-text-muted)]">言語:</label>
        <select
          value={language}
          onChange={(e) => setLanguage(e.target.value)}
          className="px-2 py-1 rounded-lg bg-surface-2 border border-surface-3 text-[var(--color-text-primary)] focus:outline-none focus:border-brand-400"
        >
          <option value="">すべて</option>
          <option value="php">PHP</option>
          <option value="go">Go</option>
          <option value="bash">Bash / Linux</option>
          <option value="docker">Docker</option>
        </select>
      </div>

      {loading && (
        <p className="text-sm text-[var(--color-text-muted)]">読み込み中...</p>
      )}
      {error && (
        <p className="text-sm text-red-500">{error}</p>
      )}

      {!loading && !error && exercises.length === 0 && (
        <p className="text-sm text-[var(--color-text-muted)]">該当する問題がありません。</p>
      )}

      {!loading && !error && categories.map((cat) => (
        <section key={cat} className="space-y-3">
          <h2 className="text-sm font-semibold text-[var(--color-text-secondary)] tracking-wide">
            {cat}
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {exercises
              .filter((e) => e.category === cat)
              .map((ex) => (
                <ExerciseCard key={ex.id} ex={ex} />
              ))}
          </div>
        </section>
      ))}
    </div>
  );
}

function ExerciseCard({ ex }: { ex: MasterExerciseWithStatus }) {
  const passRate = ex.stats.totalSubmissions > 0
    ? Math.round((ex.stats.solvedUsers / ex.stats.totalSubmissions) * 100)
    : null;

  return (
    <Link
      to={`/code-editor/${encodeURIComponent(ex.slug)}`}
      className="group block rounded-lg border border-surface-3 bg-surface-1 hover:bg-surface-2 hover:border-primary-500/50 transition-colors p-4 space-y-3"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2 text-xs text-[var(--color-text-muted)]">
            <span className="font-mono">#{ex.orderIndex}</span>
            <span className="px-1.5 py-0.5 rounded bg-surface-3 uppercase tracking-wide">
              {ex.language}
            </span>
            <DifficultyBadge level={ex.difficulty} />
          </div>
          <h3 className="mt-1 text-base font-semibold text-[var(--color-text-primary)] group-hover:text-primary-400 transition-colors truncate">
            {ex.title}
          </h3>
        </div>
        <StatusBadge status={ex.status} />
      </div>

      <p className="text-xs text-[var(--color-text-tertiary)] line-clamp-2">
        {ex.description}
      </p>

      <div className="flex items-center gap-4 text-xs text-[var(--color-text-muted)] pt-1 border-t border-surface-3">
        <span>提出 {ex.stats.totalSubmissions}</span>
        <span>正答ユーザ {ex.stats.solvedUsers}</span>
        {passRate !== null && <span>正答率 {passRate}%</span>}
      </div>
    </Link>
  );
}

function StatusBadge({ status }: { status: MasterExerciseWithStatus['status'] }) {
  if (status === 'solved') {
    return (
      <span className="flex items-center gap-1 text-xs px-2 py-0.5 rounded-full bg-green-500/15 text-green-400 border border-green-500/30 flex-shrink-0">
        <CheckCircleIcon className="w-3.5 h-3.5" />
        解いた
      </span>
    );
  }
  if (status === 'in_progress') {
    return (
      <span className="flex items-center gap-1 text-xs px-2 py-0.5 rounded-full bg-yellow-500/15 text-yellow-400 border border-yellow-500/30 flex-shrink-0">
        <ClockIcon className="w-3.5 h-3.5" />
        取り組み中
      </span>
    );
  }
  return (
    <span className="text-xs px-2 py-0.5 rounded-full bg-surface-3 text-[var(--color-text-muted)] flex-shrink-0">
      未着手
    </span>
  );
}

function DifficultyBadge({ level }: { level: number }) {
  const stars = '★'.repeat(Math.max(1, Math.min(5, level)));
  return <span className="text-yellow-500 tracking-tighter" aria-label={`難易度 ${level}`}>{stars}</span>;
}
