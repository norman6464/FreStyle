import { Link } from 'react-router-dom';
import { CheckCircleIcon, ClockIcon, CodeBracketIcon } from '@heroicons/react/24/outline';
import { FilterChip } from '../components/ui';
import { EXERCISE_LANGUAGES } from '../constants/exerciseLanguages';
import { useExerciseList } from '../hooks/useExerciseList';
import { MasterExerciseWithStatus } from '../types';

/**
 * ExerciseListPage — `/code-editor` のリスト画面。
 *
 * - カード形式で演習問題を一覧表示
 * - スクロール型ページネーション（IntersectionObserver）で次ページを自動取得
 * - カテゴリ見出しは蓄積されたリストから動的に生成する
 */
export default function ExerciseListPage() {
  const {
    language,
    setLanguage,
    exercises,
    categories,
    loading,
    loadingMore,
    error,
    sentinelRef,
  } = useExerciseList();

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

      {/* 言語の絞り込みはコース一覧のカテゴリチップと同じ操作感(常時見える一覧 +
          アクティブチップの再クリックで「すべて」に戻る)にする(FRESTYLE-101)。 */}
      <div className="flex items-center gap-2 flex-wrap" role="group" aria-label="言語で絞り込み">
        <FilterChip label="すべて" active={language === ''} onClick={() => setLanguage('')} />
        {EXERCISE_LANGUAGES.map((l) => (
          <FilterChip
            key={l.key}
            label={l.label}
            active={language === l.key}
            onClick={() => setLanguage(language === l.key ? '' : l.key)}
          />
        ))}
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

      {/* IntersectionObserver のターゲット。viewport に入ると次ページを自動取得する。 */}
      <div ref={sentinelRef} className="h-1" aria-hidden />

      {loadingMore && (
        <p className="text-sm text-center text-[var(--color-text-muted)] py-4">読み込み中...</p>
      )}
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
      className="group block rounded-lg border border-surface-3 bg-surface-1 hover:bg-surface-2 hover:border-taupe-500/50 transition-colors p-4 space-y-3"
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
          <h3 className="mt-1 text-base font-semibold text-[var(--color-text-primary)] group-hover:text-taupe-400 transition-colors truncate">
            {ex.title}
          </h3>
        </div>
        <StatusBadge status={ex.status} />
      </div>

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
  // 未着手はデフォルト状態なのでバッジを出さない。全カードに付く「未着手」チップは
  // 視覚ノイズになり、意味のある状態(解いた/取り組み中)の視認性を下げるため(FRESTYLE-64)。
  return null;
}

function DifficultyBadge({ level }: { level: number }) {
  const stars = '★'.repeat(Math.max(1, Math.min(5, level)));
  return <span className="text-yellow-500 tracking-tighter" aria-label={`難易度 ${level}`}>{stars}</span>;
}
