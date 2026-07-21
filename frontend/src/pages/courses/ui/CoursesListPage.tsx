import { useMemo, useState } from 'react';
import { useAppSelector } from '@/shared/lib/store';

import { useNavigate } from 'react-router-dom';
import {
  CheckCircleIcon,
  ChevronDownIcon,
  PlusIcon,
  PencilSquareIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';
import EmptyState from '@/shared/ui/EmptyState';
import FaviconIcon from '@/shared/ui/icons/FaviconIcon';
import Loading from '@/shared/ui/Loading';
import ConfirmModal from '@/shared/ui/ConfirmModal';
import { CourseProgressBar } from '@/entities/course';
import LanguageBadge from '@/shared/ui/LanguageBadge';
import { FilterChip } from '@/shared/ui';
import { useCourses } from '../model/useCourses';
import { useToast } from '@/shared/lib/hooks/useToast';
import { COURSE_CATEGORIES, findCourseCategory } from '@/entities/course';
import { COURSE_LANGUAGES } from '@/entities/course';

import type { Course, CourseWithProgress } from '@/entities/course';

/** 未分類('')や未知の値を未分類バケットの key('') に正規化する。 */
function normalizeCategoryKey(category: string): string {
  return findCourseCategory(category) ? category : '';
}

/**
 * CoursesListPage — `/courses` コース一覧ページ。
 *
 * - company_admin / super_admin: コースを作成 / 編集 / 削除可
 * - trainee: published コースを閲覧のみ（クリックすると `/courses/:id` 詳細へ）
 *
 * カテゴリ（学習領域 = 色）ごとのセクションでカードグリッドを表示する(FRESTYLE-68)。
 * セクションは閉じ開きでき、上部のチップでカテゴリ絞り込みができる。
 */
export default function CoursesListPage() {
  const role = useAppSelector((s) => s.auth.role);
  const canManage = role === 'company_admin' || role === 'super_admin';

  const { showToast } = useToast();
  const navigate = useNavigate();

  const { filtered, courses, loading, error, searchQuery, setSearchQuery, create, update, remove } =
    useCourses();

  const [editTarget, setEditTarget] = useState<Course | 'new' | null>(null);
  const [deleteTargetId, setDeleteTargetId] = useState<number | null>(null);
  // カテゴリ絞り込み(null = すべて)と、セクションごとの折りたたみ状態。
  const [categoryFilter, setCategoryFilter] = useState<string | null>(null);
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});

  const toggleGroup = (key: string) =>
    setCollapsed((prev) => ({ ...prev, [key]: !prev[key] }));

  // 絞り込みチップに出すカテゴリ = 実際にコースが存在するものだけ(死にチップを出さない)。
  const chipCategories = useMemo(() => {
    const present = new Set(courses.map((c) => normalizeCategoryKey(c.category)));
    return {
      cats: COURSE_CATEGORIES.filter((c) => present.has(c.key)),
      hasUncategorized: present.has(''),
    };
  }, [courses]);

  // 検索(filtered) → カテゴリ絞り込み → 定義順のセクションにグルーピング。
  const visible = useMemo(
    () =>
      categoryFilter === null
        ? filtered
        : filtered.filter((c) => normalizeCategoryKey(c.category) === categoryFilter),
    [filtered, categoryFilter],
  );
  const groups = useMemo(() => {
    const order = [...COURSE_CATEGORIES.map((c) => c.key), ''];
    return order
      .map((key) => ({
        key,
        courses: visible.filter((c) => normalizeCategoryKey(c.category) === key),
      }))
      .filter((g) => g.courses.length > 0);
  }, [visible]);

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

      {/* カテゴリ絞り込みチップ(色＝学習領域で分別できるように。メンター/trainee 共通) */}
      {courses.length > 0 && (chipCategories.cats.length > 0 || chipCategories.hasUncategorized) && (
        <div className="flex items-center gap-2 flex-wrap" role="group" aria-label="カテゴリで絞り込み">
          <FilterChip
            label="すべて"
            active={categoryFilter === null}
            onClick={() => setCategoryFilter(null)}
          />
          {chipCategories.cats.map((c) => (
            <FilterChip
              key={c.key}
              label={c.label}
              active={categoryFilter === c.key}
              activeClass={c.badgeClass}
              onClick={() => setCategoryFilter((prev) => (prev === c.key ? null : c.key))}
            />
          ))}
          {chipCategories.hasUncategorized && (
            <FilterChip
              label="未分類"
              active={categoryFilter === ''}
              onClick={() => setCategoryFilter((prev) => (prev === '' ? null : ''))}
            />
          )}
        </div>
      )}

      {error && <p className="text-sm text-red-500">{error}</p>}

      {loading && courses.length === 0 ? (
        <Loading className="py-12" />
      ) : visible.length === 0 ? (
        <EmptyState
          icon={FaviconIcon}
          title={searchQuery || categoryFilter !== null ? '該当するコースがありません' : 'コースがありません'}
          description={
            searchQuery || categoryFilter !== null
              ? '検索条件やカテゴリの絞り込みを変更してみてください'
              : canManage
                ? '最初のコースを作成しましょう'
                : '管理者がコースを公開すると、 ここに表示されます'
          }
          action={
            canManage && !searchQuery && categoryFilter === null
              ? { label: '新しいコース', onClick: () => setEditTarget('new') }
              : undefined
          }
        />
      ) : (
        <div className="space-y-6">
          {groups.map((g) => {
            const def = findCourseCategory(g.key);
            const isCollapsed = !!collapsed[g.key];
            return (
              <section key={g.key || 'uncategorized'} aria-label={def ? def.label : '未分類'}>
                <button
                  type="button"
                  onClick={() => toggleGroup(g.key)}
                  aria-expanded={!isCollapsed}
                  className="flex items-center gap-2 mb-3 rounded-md px-1 py-0.5 hover:bg-surface-2 transition-colors"
                >
                  <ChevronDownIcon
                    aria-hidden="true"
                    className={`w-4 h-4 text-[var(--color-text-muted)] transition-transform ${
                      isCollapsed ? '-rotate-90' : ''
                    }`}
                  />
                  {def ? (
                    <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${def.badgeClass}`}>
                      {def.label}
                    </span>
                  ) : (
                    <span className="text-xs font-medium px-2 py-0.5 rounded-full bg-surface-3 text-[var(--color-text-muted)]">
                      未分類
                    </span>
                  )}
                  <span className="text-xs text-[var(--color-text-muted)]">{g.courses.length} 件</span>
                </button>
                {!isCollapsed && (
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {g.courses.map((c) => (
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
              </section>
            );
          })}
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
  course: CourseWithProgress;
  canManage: boolean;
  onOpen: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  // 「色＝学習領域」の連想(FRESTYLE-67)。カードは左の色帯のみで表現し、
  // カテゴリ名ラベルはセクション見出しが担う(同一セクション内での繰り返しを避ける。FRESTYLE-68)。
  const category = findCourseCategory(course.category);
  // 受講者視点で全章完了したコースは一目で分かる見た目にする(FRESTYLE-114)。
  const isCompleted = !canManage && course.materialCount > 0 && course.completedCount >= course.materialCount;
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
      className={`group border border-l-4 ${
        category ? category.barClass : 'border-l-surface-3'
      } ${
        isCompleted
          ? 'bg-emerald-500/5 border-emerald-500/40'
          : 'bg-surface-1 border-surface-3'
      } rounded-lg p-4 cursor-pointer hover:border-taupe-500/50 hover:shadow-md transition-all`}
    >
      <div className="flex items-start justify-between gap-2 mb-2">
        <div className="flex items-center gap-2 min-w-0">
          <h2 className="text-base font-semibold text-[var(--color-text-primary)] truncate">
            {course.title || '無題のコース'}
          </h2>
        </div>
        {isCompleted && (
          <span className="flex items-center gap-1 text-xs font-semibold px-2 py-0.5 rounded-full bg-emerald-600 text-white flex-shrink-0">
            <CheckCircleIcon className="w-3.5 h-3.5" />
            完了
          </span>
        )}
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
      <p className="flex items-center gap-2 text-xs text-[var(--color-text-muted)] mb-3">
        {/* パッと見てどんな言語のコースか分かるように言語バッジを出す(FRESTYLE-114)。 */}
        {course.language && <LanguageBadge language={course.language} />}
        <span>
          {course.isPublished ? '公開中' : '下書き'} ・ 更新 {formatDate(course.updatedAt)}
        </span>
      </p>
      <p className="text-sm text-[var(--color-text-secondary)] line-clamp-3 min-h-[3.6em]">
        {course.description || 'コース説明が未設定です'}
      </p>
      {/* 受講者のみ進捗を表示(FRESTYLE-98)。0 章のコースはバーを出さない。 */}
      {!canManage && course.materialCount > 0 && (
        <div className="mt-3">
          <CourseProgressBar completed={course.completedCount} total={course.materialCount} />
        </div>
      )}
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
    language: string;
    sortOrder: number;
    isPublished: boolean;
  }) => Promise<void>;
}

function CourseFormModal({ initial, onClose, onSubmit }: CourseFormProps) {
  const [title, setTitle] = useState(initial?.title ?? '');
  const [description, setDescription] = useState(initial?.description ?? '');
  const [category, setCategory] = useState(initial?.category ?? '');
  const [language, setLanguage] = useState(initial?.language ?? '');
  const [sortOrder, setSortOrder] = useState(initial?.sortOrder ?? 100);
  const [isPublished, setIsPublished] = useState(initial?.isPublished ?? false);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    setSubmitting(true);
    try {
      await onSubmit({
        title: title.trim(),
        description: description.trim(),
        category,
        language,
        sortOrder,
        isPublished,
      });
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
            <span className="block text-sm text-[var(--color-text-secondary)] mb-1">主に扱う言語・技術</span>
            <select
              value={language}
              onChange={(e) => setLanguage(e.target.value)}
              className="w-full px-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm focus:outline-none focus:border-brand-400"
            >
              <option value="">未設定（言語が主題でないコース）</option>
              {COURSE_LANGUAGES.map((l) => (
                <option key={l.key} value={l.key}>
                  {l.label}
                </option>
              ))}
            </select>
            <span className="block text-xs text-[var(--color-text-muted)] mt-1">
              一覧カードに言語バッジとして表示されます
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
