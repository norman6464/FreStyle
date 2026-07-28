import { TrashIcon } from '@heroicons/react/24/outline';
import type { TeachingMaterial } from '@/entities/course';

/** 左パネル(教材リスト)の 1 行。クリック / Enter・Space で選択、管理ロールでは削除ボタンを出す。 */
export default function MaterialListItem({
  material,
  isActive,
  onSelect,
  onDelete,
}: {
  material: TeachingMaterial;
  isActive: boolean;
  onSelect: (id: number) => void;
  onDelete?: (id: number) => void;
}) {
  return (
    <div
      role="button"
      tabIndex={0}
      onClick={() => onSelect(material.id)}
      onKeyDown={(event) => {
        if (event.key === 'Enter' || event.key === ' ') {
          event.preventDefault();
          onSelect(material.id);
        }
      }}
      className={`group flex items-start gap-2.5 pl-4 pr-2 py-2.5 cursor-pointer transition-colors border-l-2 ${
        isActive
          ? 'border-brand-500 bg-[var(--color-nav-active)]'
          : 'border-transparent hover:bg-[var(--color-nav-hover)]'
      }`}
    >
      <p className={`flex-1 min-w-0 text-[13px] leading-snug line-clamp-2 ${
        isActive
          ? 'font-medium text-[var(--color-text-primary)]'
          : 'text-[var(--color-text-secondary)]'
      }`}>
        {material.title || '無題の教材'}
      </p>

      {onDelete && (
        <button
          onClick={(event) => {
            event.stopPropagation();
            onDelete(material.id);
          }}
          className="opacity-0 group-hover:opacity-100 p-1 hover:bg-red-900/30 rounded transition-opacity flex-shrink-0"
          aria-label="教材を削除"
        >
          <TrashIcon className="w-3.5 h-3.5 text-[var(--color-text-muted)] hover:text-red-400" />
        </button>
      )}
    </div>
  );
}
