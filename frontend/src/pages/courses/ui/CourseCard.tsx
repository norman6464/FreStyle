import {
  CheckCircleIcon,
  PencilSquareIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';
import { CourseProgressBar } from '@/entities/course';
import LanguageBadge from '@/shared/ui/LanguageBadge';
import { findCourseCategory } from '@/entities/course';
import type { CourseWithProgress } from '@/entities/course';

/**
 * CourseCard — コース 1 件のカード。領域一覧ページ（CoursesListPage）で使う。
 *
 * 「色＝学習領域」の連想（FRESTYLE-67）でカードは左の色帯で領域を表す。
 * 受講者視点で全章完了したコースは緑アクセントで一目で分かるようにする（FRESTYLE-114）。
 */
export default function CourseCard({
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
  const category = findCourseCategory(course.category);
  const isCompleted =
    !canManage && course.materialCount > 0 && course.completedCount >= course.materialCount;
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
        isCompleted ? 'bg-emerald-500/5 border-emerald-500/40' : 'bg-surface-1 border-surface-3'
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
        {course.language && <LanguageBadge language={course.language} />}
        <span>
          {course.isPublished ? '公開中' : '下書き'} ・ 更新 {formatDate(course.updatedAt)}
        </span>
      </p>
      <p className="text-sm text-[var(--color-text-secondary)] line-clamp-3 min-h-[3.6em]">
        {course.description || 'コース説明が未設定です'}
      </p>
      {!canManage && course.materialCount > 0 && (
        <div className="mt-3">
          <CourseProgressBar completed={course.completedCount} total={course.materialCount} />
        </div>
      )}
    </div>
  );
}

function formatDate(iso: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  return `${d.getFullYear()}/${String(d.getMonth() + 1).padStart(2, '0')}/${String(d.getDate()).padStart(2, '0')}`;
}
