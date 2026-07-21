import { useMemo, useState } from 'react';
import { useAppSelector } from '@/shared/lib/store';
import { useNavigate, useParams, Link } from 'react-router-dom';
import { PlusIcon, ArrowLeftIcon } from '@heroicons/react/24/outline';
import EmptyState from '@/shared/ui/EmptyState';
import FaviconIcon from '@/shared/ui/icons/FaviconIcon';
import Loading from '@/shared/ui/Loading';
import ConfirmModal from '@/shared/ui/ConfirmModal';
import { COURSE_CATEGORIES, findCourseCategory } from '@/entities/course';
import type { Course } from '@/entities/course';
import { useCourses } from '../model/useCourses';
import { useToast } from '@/shared/lib/hooks/useToast';
import CourseCard from './CourseCard';
import CourseFormModal from './CourseFormModal';
import CategoryIcon from './CategoryIcon';

/** 未分類('')や未知の値を未分類バケットの key('') に正規化する。 */
function normalizeCategoryKey(category: string): string {
  return findCourseCategory(category) ? category : '';
}

/** URL の :category（'uncategorized' は未分類の '' に対応）。 */
const UNCATEGORIZED_SLUG = 'uncategorized';

/**
 * CoursesListPage — `/courses/category/:category` の領域スコープ一覧（FRESTYLE-177）。
 *
 * 入口は CourseCategorySelectPage（`/courses`）で、そこで選んだ 1 領域のコースだけを
 * カードグリッドで出す。以前は全領域を 1 画面に縦積みしていた。
 * company_admin / super_admin は作成 / 編集 / 削除でき、作成時はこの領域を初期選択にする。
 */
export default function CoursesListPage() {
  const role = useAppSelector((s) => s.auth.role);
  const canManage = role === 'company_admin' || role === 'super_admin';

  const { showToast } = useToast();
  const navigate = useNavigate();
  const { category: categoryParam } = useParams<{ category: string }>();

  // URL の :category を内部 key に正規化。'uncategorized' → ''、未知の値も '' 扱い。
  const categoryKey =
    categoryParam === UNCATEGORIZED_SLUG
      ? ''
      : COURSE_CATEGORIES.some((c) => c.key === categoryParam)
        ? (categoryParam as string)
        : '';
  const categoryDef = findCourseCategory(categoryKey);
  const categoryLabel = categoryDef ? categoryDef.label : '未分類';

  const { courses, loading, error, searchQuery, setSearchQuery, create, update, remove } =
    useCourses();

  const [editTarget, setEditTarget] = useState<Course | 'new' | null>(null);
  const [deleteTargetId, setDeleteTargetId] = useState<number | null>(null);

  // この領域のコースだけに絞り、検索(タイトル/説明)を掛ける。
  const visible = useMemo(() => {
    const inCat = courses.filter((c) => normalizeCategoryKey(c.category) === categoryKey);
    const q = searchQuery.trim().toLowerCase();
    if (!q) return inCat;
    return inCat.filter(
      (c) => c.title.toLowerCase().includes(q) || c.description.toLowerCase().includes(q),
    );
  }, [courses, categoryKey, searchQuery]);

  const handleConfirmDelete = async () => {
    if (deleteTargetId == null) return;
    await remove(deleteTargetId);
    setDeleteTargetId(null);
    showToast('success', 'コースを削除しました');
  };

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-6xl mx-auto space-y-6">
      <div>
        <Link
          to="/courses"
          className="inline-flex items-center gap-1 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
        >
          <ArrowLeftIcon className="w-3.5 h-3.5" />
          コース一覧に戻る
        </Link>
      </div>

      <header className="flex items-start justify-between gap-4 flex-wrap">
        <div className="flex items-center gap-3 min-w-0">
          <span className={`flex-shrink-0 ${categoryDef ? categoryDef.accentClass : 'text-[var(--color-text-muted)]'}`}>
            <CategoryIcon categoryKey={categoryKey} className="w-8 h-8" />
          </span>
          <div className="min-w-0">
            <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">{categoryLabel}</h1>
            <p className="text-sm text-[var(--color-text-tertiary)] mt-0.5">
              この学習領域のコース一覧。 各コースをクリックすると配下の教材を閲覧できます。
            </p>
          </div>
        </div>
        {canManage && (
          <button
            onClick={() => setEditTarget('new')}
            className="bg-brand-500 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-brand-600 transition-colors flex items-center gap-2"
          >
            <PlusIcon className="w-4 h-4" />
            新しいコース
          </button>
        )}
      </header>

      <div className="relative max-w-md">
        <input
          type="text"
          placeholder="この領域のコースを検索..."
          aria-label="コースを検索"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full px-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:outline-none focus:border-brand-400 transition-colors"
        />
      </div>

      {error && <p className="text-sm text-red-500">{error}</p>}

      {loading && courses.length === 0 ? (
        <Loading className="py-12" />
      ) : visible.length === 0 ? (
        <EmptyState
          icon={FaviconIcon}
          title={searchQuery ? '該当するコースがありません' : 'この領域のコースはまだありません'}
          description={
            searchQuery
              ? '検索条件を変えてみてください'
              : canManage
                ? 'この領域に最初のコースを作成しましょう'
                : '管理者がコースを公開すると、 ここに表示されます'
          }
          action={
            canManage && !searchQuery
              ? { label: '新しいコース', onClick: () => setEditTarget('new') }
              : undefined
          }
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {visible.map((c) => (
            <CourseCard
              key={c.id}
              course={c}
              canManage={canManage}
              onOpen={() => navigate(`/courses/${c.id}`)}
              onEdit={() => setEditTarget(c)}
              onDelete={() => setDeleteTargetId(c.id)}
            />
          ))}
        </div>
      )}

      {editTarget !== null && (
        <CourseFormModal
          initial={editTarget === 'new' ? null : editTarget}
          defaultCategory={editTarget === 'new' ? categoryKey : ''}
          onClose={() => setEditTarget(null)}
          onSubmit={async (payload) => {
            if (editTarget === 'new') {
              const created = await create(payload);
              if (created) {
                showToast('success', 'コースを作成しました');
                setEditTarget(null);
              }
            } else {
              const updated = await update(editTarget.id, payload);
              if (updated) {
                showToast('success', 'コースを更新しました');
                setEditTarget(null);
              }
            }
          }}
        />
      )}

      <ConfirmModal
        isOpen={deleteTargetId !== null}
        message="このコースを削除しますか？ 配下の教材もすべて削除されます。 この操作は元に戻せません。"
        onConfirm={handleConfirmDelete}
        onCancel={() => setDeleteTargetId(null)}
        isDanger={true}
      />
    </div>
  );
}
