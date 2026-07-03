import { useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import {
  PlusIcon,
  PencilSquareIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';
import EmptyState from '../components/EmptyState';
import FaviconIcon from '../components/icons/FaviconIcon';
import Loading from '../components/Loading';
import ConfirmModal from '../components/ConfirmModal';
import { useCourses } from '../hooks/useCourses';
import { useToast } from '../hooks/useToast';
import { COURSE_CATEGORIES, findCourseCategory } from '../constants/courseCategories';
import type { RootState } from '../store';
import type { Course } from '../types';

/**
 * CoursesListPage — `/courses` コース一覧ページ。
 *
 * - company_admin / super_admin: コースを作成 / 編集 / 削除可
 * - trainee: published コースを閲覧のみ（クリックすると `/courses/:id` 詳細へ）
 *
 * カードグリッドで表示し、 各カードに教材数 / 公開状態 / 説明文を出す。
 */
export default function CoursesListPage() {
  const role = useSelector((s: RootState) => s.auth.role);
  const canManage = role === 'company_admin' || role === 'super_admin';

  const { showToast } = useToast();
  const navigate = useNavigate();

  const { filtered, courses, loading, error, searchQuery, setSearchQuery, create, update, remove } =
    useCourses();

  const [editTarget, setEditTarget] = useState<Course | 'new' | null>(null);
  const [deleteTargetId, setDeleteTargetId] = useState<number | null>(null);

  const handleConfirmDelete = async () => {
    if (deleteTargetId == null) return;
    await remove(deleteTargetId);
    setDeleteTargetId(null);
    showToast('success', 'コースを削除しました');
  };

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-6xl mx-auto space-y-6">
      <header className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">コース</h1>
          <p className="text-sm text-[var(--color-text-tertiary)] mt-1">
            学習コースの一覧。 各コースをクリックすると配下の教材を閲覧できます。
          </p>
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
          placeholder="コースを検索..."
          aria-label="コースを検索"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full px-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:outline-none focus:border-brand-400 transition-colors"
        />
      </div>

      {error && <p className="text-sm text-red-500">{error}</p>}

      {loading && courses.length === 0 ? (
        <Loading className="py-12" />
      ) : filtered.length === 0 ? (
        <EmptyState
          icon={FaviconIcon}
          title={searchQuery ? '該当するコースがありません' : 'コースがありません'}
          description={
            searchQuery
              ? '検索条件を変更してみてください'
              : canManage
                ? '最初のコースを作成しましょう'
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
          {filtered.map((c) => (
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

function CourseCard({
  course,
  canManage,
  onOpen,
  onEdit,
  onDelete,
}: {
  course: Course;
  canManage: boolean;
  onOpen: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  // 「色＝学習領域」の連想(FRESTYLE-67)。左の色帯 + カテゴリ名バッジで表現し、
  // 色のみに依存しない(未分類は無色 = 従来表示)。
  const category = findCourseCategory(course.category);
  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onOpen}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onOpen();
        }
      }}
      className={`group bg-surface-1 border border-surface-3 border-l-4 ${
        category ? category.barClass : 'border-l-surface-3'
      } rounded-lg p-4 cursor-pointer hover:border-taupe-500/50 hover:shadow-md transition-all`}
    >
      <div className="flex items-start justify-between gap-2 mb-2">
        <div className="flex items-center gap-2 min-w-0">
          <h2 className="text-base font-semibold text-[var(--color-text-primary)] truncate">
            {course.title || '無題のコース'}
          </h2>
        </div>
        {canManage && (
          <div className="opacity-0 group-hover:opacity-100 flex items-center gap-1 transition-opacity">
            <button
              onClick={(e) => {
                e.stopPropagation();
                onEdit();
              }}
              className="p-1 rounded hover:bg-surface-3 text-[var(--color-text-muted)]"
              aria-label="コースを編集"
            >
              <PencilSquareIcon className="w-4 h-4" />
            </button>
            <button
              onClick={(e) => {
                e.stopPropagation();
                onDelete();
              }}
              className="p-1 rounded hover:bg-red-900/30 text-[var(--color-text-muted)] hover:text-red-400"
              aria-label="コースを削除"
            >
              <TrashIcon className="w-4 h-4" />
            </button>
          </div>
        )}
      </div>
      <div className="flex items-center gap-2 mb-3">
        {category && (
          <span className={`text-[11px] px-2 py-0.5 rounded-full flex-shrink-0 ${category.badgeClass}`}>
            {category.label}
          </span>
        )}
        <p className="text-xs text-[var(--color-text-muted)]">
          {course.isPublished ? '公開中' : '下書き'} ・ 更新 {formatDate(course.updatedAt)}
        </p>
      </div>
      <p className="text-sm text-[var(--color-text-secondary)] line-clamp-3 min-h-[3.6em]">
        {course.description || 'コース説明が未設定です'}
      </p>
    </div>
  );
}

interface CourseFormProps {
  initial: Course | null;
  onClose: () => void;
  onSubmit: (payload: {
    title: string;
    description: string;
    category: string;
    sortOrder: number;
    isPublished: boolean;
  }) => Promise<void>;
}

function CourseFormModal({ initial, onClose, onSubmit }: CourseFormProps) {
  const [title, setTitle] = useState(initial?.title ?? '');
  const [description, setDescription] = useState(initial?.description ?? '');
  const [category, setCategory] = useState(initial?.category ?? '');
  const [sortOrder, setSortOrder] = useState(initial?.sortOrder ?? 100);
  const [isPublished, setIsPublished] = useState(initial?.isPublished ?? false);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    setSubmitting(true);
    try {
      await onSubmit({ title: title.trim(), description: description.trim(), category, sortOrder, isPublished });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <div
        className="bg-surface-1 rounded-lg shadow-xl border border-surface-3 max-w-lg w-full p-6 space-y-4"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-bold text-[var(--color-text-primary)]">
          {initial ? 'コースを編集' : '新しいコース'}
        </h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <label className="block">
            <span className="block text-sm text-[var(--color-text-secondary)] mb-1">タイトル *</span>
            <input
              required
              maxLength={200}
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full px-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm focus:outline-none focus:border-brand-400"
            />
          </label>
          <label className="block">
            <span className="block text-sm text-[var(--color-text-secondary)] mb-1">説明</span>
            <textarea
              rows={3}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="このコースで学べる内容の概要"
              className="w-full px-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm focus:outline-none focus:border-brand-400 resize-y"
            />
          </label>
          <label className="block">
            <span className="block text-sm text-[var(--color-text-secondary)] mb-1">カテゴリ（学習領域）</span>
            <select
              value={category}
              onChange={(e) => setCategory(e.target.value)}
              className="w-full px-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm focus:outline-none focus:border-brand-400"
            >
              <option value="">未分類</option>
              {COURSE_CATEGORIES.map((c) => (
                <option key={c.key} value={c.key}>
                  {c.label}
                </option>
              ))}
            </select>
            <span className="block text-xs text-[var(--color-text-muted)] mt-1">
              一覧カードの色分けに使われます（色＝学習領域）
            </span>
          </label>
          <label className="block">
            <span className="block text-sm text-[var(--color-text-secondary)] mb-1">並び順 (昇順)</span>
            <input
              type="number"
              value={sortOrder}
              onChange={(e) => setSortOrder(Number(e.target.value) || 0)}
              className="w-32 px-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm focus:outline-none focus:border-brand-400"
            />
            <span className="block text-xs text-[var(--color-text-muted)] mt-1">
              小さい値が上に来ます。 既定: 100
            </span>
          </label>
          <label className="flex items-center gap-2 text-sm text-[var(--color-text-secondary)]">
            <input
              type="checkbox"
              checked={isPublished}
              onChange={(e) => setIsPublished(e.target.checked)}
            />
            trainee に公開
          </label>
          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="px-3 py-1.5 text-sm text-[var(--color-text-muted)] hover:bg-surface-2 rounded"
            >
              キャンセル
            </button>
            <button
              type="submit"
              disabled={!title.trim() || submitting}
              className="bg-brand-500 text-white px-4 py-1.5 rounded text-sm font-medium hover:bg-brand-600 transition-colors disabled:opacity-50"
            >
              {submitting ? '保存中...' : initial ? '更新' : '作成'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function formatDate(iso: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  return `${d.getFullYear()}/${String(d.getMonth() + 1).padStart(2, '0')}/${String(d.getDate()).padStart(2, '0')}`;
}
