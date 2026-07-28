import { Link } from 'react-router-dom';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { CheckCircleIcon as CheckCircleSolidIcon } from '@heroicons/react/24/solid';
import { CourseProgressBar } from '@/entities/course';
import type { Course, TeachingMaterial } from '@/entities/course';

/**
 * 閲覧ビューの右サイドバーに出す章一覧(FRESTYLE-118)。
 * コース一覧への戻り + コース名 + 進捗バー + 章リスト(完了はチェック・現在章はハイライト)。
 */
export default function ChapterNav({
  course,
  materials,
  selectedId,
  completedIds,
  completedCount,
  onSelect,
}: {
  course: Course;
  materials: TeachingMaterial[];
  selectedId: number;
  completedIds: Set<number>;
  completedCount: number;
  onSelect: (id: number) => void;
}) {
  return (
    // 親カードが高さを制限したとき、見出し(コース名・進捗)は固定して章リストだけを内側でスクロールさせる(FRESTYLE-144)。
    <div className="flex min-h-0 flex-col">
      <div className="flex-shrink-0">
        <Link
          to="/courses"
          className="inline-flex items-center gap-1 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
        >
          <ArrowLeftIcon className="w-3.5 h-3.5" />
          コース一覧
        </Link>
        <p className="mt-2 text-sm font-semibold text-[var(--color-text-primary)]">
          {course.title || '無題のコース'}
          <span className="ml-2 text-xs font-normal text-[var(--color-text-muted)]">
            {materials.length} 章
          </span>
        </p>
        {materials.length > 0 && (
          <div className="mt-2">
            <CourseProgressBar completed={completedCount} total={materials.length} />
          </div>
        )}
      </div>
      <nav aria-label="章一覧" className="mt-3 min-h-0 flex-1 space-y-0.5 overflow-y-auto pr-1">
        {materials.map((material, index) => {
          const isCurrent = material.id === selectedId;
          return (
            <button
              key={material.id}
              type="button"
              onClick={() => onSelect(material.id)}
              aria-current={isCurrent ? 'page' : undefined}
              className={`w-full flex items-start gap-2 rounded-md px-2 py-1.5 text-left text-[13px] leading-snug transition-colors ${
                isCurrent
                  ? 'bg-brand-500/10 font-medium text-[var(--color-text-primary)]'
                  : 'text-[var(--color-text-secondary)] hover:bg-surface-2'
              }`}
            >
              {completedIds.has(material.id) ? (
                <CheckCircleSolidIcon className="w-4 h-4 mt-0.5 text-emerald-500 flex-shrink-0" aria-label="完了" />
              ) : (
                <span className="w-4 h-4 mt-0.5 flex items-center justify-center text-[10px] rounded-full border border-surface-3 text-[var(--color-text-muted)] flex-shrink-0">
                  {index + 1}
                </span>
              )}
              <span className="line-clamp-2">{material.title || '無題の教材'}</span>
            </button>
          );
        })}
      </nav>
    </div>
  );
}
