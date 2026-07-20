import { Link, useParams } from 'react-router-dom';
import { ArrowLeftIcon, CheckCircleIcon, ClockIcon } from '@heroicons/react/24/outline';
import { EXERCISE_LANGUAGES } from '@/entities/exercise';
import LanguageBadge from '@/shared/ui/LanguageBadge';
import LanguageIcon from '@/shared/ui/LanguageIcon';
import { useExerciseList } from '@/hooks/useExerciseList';
import type { MasterExerciseWithStatus } from '@/entities/exercise';

/**
 * ExerciseListPage — `/code-editor/lang/:language` の問題一覧画面。
 *
 * - 言語は URL から決まる（言語選択カード `/code-editor` から遷移してくる。FRESTYLE-152）
 * - カード形式で演習問題を一覧表示
 * - スクロール型ページネーション（IntersectionObserver）で次ページを自動取得
 * - カテゴリ見出しは蓄積されたリストから動的に生成する
 */
export default function ExerciseListPage() {
  const { language: routeLanguage = '' } = useParams<{ language: string }>();
  const {
    exercises,
    categories,
    loading,
    loadingMore,
    error,
    sentinelRef,
  } = useExerciseList(routeLanguage);

  const label =
    EXERCISE_LANGUAGES.find((l) => l.key === routeLanguage)?.label ?? routeLanguage;

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-5xl mx-auto space-y-6">
      <header className="space-y-3">
        <Link
          to="/code-editor"
          className="inline-flex items-center gap-1 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
        >
          <ArrowLeftIcon className="w-3.5 h-3.5" />
          言語を選びなおす
        </Link>
        <div className="flex items-center gap-3">
          <LanguageIcon language={routeLanguage} className="w-9 h-9 flex-shrink-0" />
          <div>
            <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">{label}</h1>
            <p className="text-sm text-[var(--color-text-tertiary)]">
              解きたい問題を選んでください。提出すると採点結果が履歴に残ります。
            </p>
          </div>
        </div>
      </header>

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
          {/* カテゴリ見出しは無色のテキスト(FRESTYLE-112 の色付けはユーザー要望で撤回)。 */}
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
            <LanguageBadge language={ex.language} />
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
  // 淡色だと白背景で薄く「見えにくい」ため、濃色塗り + 白文字ではっきり区別する(FRESTYLE-112)。
  if (status === 'solved') {
    return (
      <span className="flex items-center gap-1 text-xs font-semibold px-2 py-0.5 rounded-full bg-emerald-600 text-white flex-shrink-0">
        <CheckCircleIcon className="w-3.5 h-3.5" />
        解いた
      </span>
    );
  }
  if (status === 'in_progress') {
    return (
      <span className="flex items-center gap-1 text-xs font-semibold px-2 py-0.5 rounded-full bg-amber-600 text-white flex-shrink-0">
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
