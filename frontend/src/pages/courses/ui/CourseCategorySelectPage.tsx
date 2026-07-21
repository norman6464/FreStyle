import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { AcademicCapIcon, PlusIcon, ChevronRightIcon } from '@heroicons/react/24/outline';
import { useAppSelector } from '@/shared/lib/store';
import Loading from '@/shared/ui/Loading';
import EmptyState from '@/shared/ui/EmptyState';
import FaviconIcon from '@/shared/ui/icons/FaviconIcon';
import { COURSE_CATEGORIES, findCourseCategory } from '@/entities/course';
import type { CourseCategoryDef, CourseWithProgress } from '@/entities/course';
import { useToast } from '@/shared/lib/hooks/useToast';
import { useCourses } from '../model/useCourses';
import CategoryIcon from './CategoryIcon';
import CourseFormModal from './CourseFormModal';

/** 未分類('')や未知の値を未分類バケットの key('') に正規化する。 */
function normalizeCategoryKey(category: string): string {
  return findCourseCategory(category) ? category : '';
}

/** URL で未分類を表す予約語。'' はパスに使えないため。 */
const UNCATEGORIZED_SLUG = 'uncategorized';

interface CategoryBucket {
  /** COURSE_CATEGORIES の定義。未分類は null。 */
  def: CourseCategoryDef | null;
  /** 遷移先の :category（未分類は 'uncategorized'）。 */
  slug: string;
  label: string;
  total: number;
  /** 受講者の進捗集計（章単位）。管理者は使わない。 */
  materials: number;
  completed: number;
}

/**
 * CourseCategorySelectPage — `/courses` の入口（FRESTYLE-177）。
 *
 * 演習（`/code-editor`）と同じく「まず学習領域を選ぶ → その領域のコース一覧へ」の
 * 2 段構成にする。全コースを 1 画面に縦積みしていたのを、領域の選択カードにした。
 * コースが 1 件でもある領域だけカードを出す（死んだ領域カードを出さない）。
 */
export default function CourseCategorySelectPage() {
  const role = useAppSelector((s) => s.auth.role);
  const canManage = role === 'company_admin' || role === 'super_admin';
  const { showToast } = useToast();
  const { courses, loading, error, create } = useCourses();
  const [creating, setCreating] = useState(false);

  const buckets = useMemo<CategoryBucket[]>(() => {
    const order: { def: CourseCategoryDef | null; key: string; slug: string; label: string }[] = [
      ...COURSE_CATEGORIES.map((def) => ({ def, key: def.key, slug: def.key, label: def.label })),
      { def: null, key: '', slug: UNCATEGORIZED_SLUG, label: '未分類' },
    ];
    return order
      .map(({ def, key, slug, label }) => {
        const inCat = courses.filter((c) => normalizeCategoryKey(c.category) === key);
        return {
          def,
          slug,
          label,
          total: inCat.length,
          materials: inCat.reduce((s, c) => s + c.materialCount, 0),
          completed: inCat.reduce((s, c) => s + c.completedCount, 0),
        };
      })
      .filter((b) => b.total > 0);
  }, [courses]);

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-5xl mx-auto space-y-6">
      <header className="flex items-start justify-between gap-4 flex-wrap">
        <div className="space-y-2">
          <div className="flex items-center gap-2 text-[var(--color-text-muted)]">
            <AcademicCapIcon className="w-5 h-5" />
            <span className="text-xs uppercase tracking-wider">Courses</span>
          </div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">コース</h1>
          <p className="text-sm text-[var(--color-text-tertiary)]">
            学びたい学習領域を選んでください。各コースを開くと配下の教材を閲覧できます。
          </p>
        </div>
        {canManage && (
          <button
            onClick={() => setCreating(true)}
            className="bg-brand-500 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-brand-600 transition-colors flex items-center gap-2"
          >
            <PlusIcon className="w-4 h-4" />
            新しいコース
          </button>
        )}
      </header>

      {error && <p className="text-sm text-red-500">{error}</p>}

      {loading && courses.length === 0 ? (
        <Loading className="py-12" />
      ) : buckets.length === 0 ? (
        <EmptyState
          icon={FaviconIcon}
          title="コースがありません"
          description={
            canManage
              ? '最初のコースを作成しましょう'
              : '管理者がコースを公開すると、 ここに表示されます'
          }
          action={canManage ? { label: '新しいコース', onClick: () => setCreating(true) } : undefined}
        />
      ) : (
        <ul className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {buckets.map((b) => (
            <li key={b.slug}>
              <CategorySelectCard bucket={b} showProgress={!canManage} />
            </li>
          ))}
        </ul>
      )}

      {creating && (
        <CourseFormModal
          initial={null}
          onClose={() => setCreating(false)}
          onSubmit={async (payload) => {
            const created = await create(payload);
            if (created) {
              showToast('success', 'コースを作成しました');
              setCreating(false);
            }
          }}
        />
      )}
    </div>
  );
}

function CategorySelectCard({
  bucket,
  showProgress,
}: {
  bucket: CategoryBucket;
  showProgress: boolean;
}) {
  const { def, slug, label, total, materials, completed } = bucket;
  const accent = def ? def.accentClass : 'text-[var(--color-text-muted)]';
  const percent = materials > 0 ? Math.round((completed / materials) * 100) : 0;
  const allDone = showProgress && materials > 0 && completed >= materials;

  return (
    <Link
      to={`/courses/category/${slug}`}
      aria-label={`${label} のコース一覧へ（${total} コース）`}
      className="group flex h-full flex-col gap-4 rounded-lg border border-surface-3 bg-surface-1 p-5 shadow-sm transition-colors hover:border-taupe-500/50 hover:bg-surface-2"
    >
      <div className="flex items-center gap-3">
        <span className={`flex-shrink-0 ${accent}`}>
          <CategoryIcon categoryKey={def ? def.key : ''} className="w-9 h-9" />
        </span>
        <div className="min-w-0">
          <h2 className="text-base font-semibold text-[var(--color-text-primary)] truncate">
            {label}
          </h2>
          <p className="text-xs text-[var(--color-text-muted)]">{total} コース</p>
        </div>
        <ChevronRightIcon className="ml-auto w-5 h-5 text-[var(--color-text-muted)] transition-transform group-hover:translate-x-0.5" />
      </div>

      {/* 受講者のみ、その領域全体の学習進捗（章単位の集計）を出す。 */}
      {showProgress && materials > 0 && (
        <div className="space-y-1.5">
          <div className="flex items-center justify-between text-xs">
            <span className="text-[var(--color-text-muted)]">
              {completed}/{materials} 章完了
            </span>
            {allDone && <span className="font-semibold text-emerald-600">すべて完了</span>}
          </div>
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
                allDone ? 'bg-emerald-500' : 'bg-brand-500'
              }`}
              style={{ width: `${percent}%` }}
            />
          </div>
        </div>
      )}
    </Link>
  );
}
