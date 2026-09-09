export interface BacklogReorderBarProps {
  /** 選択中チケットの表示キー（例 FRESTYLE-457）。未選択なら null。 */
  selectedKey: string | null;
  /** 選択中チケットが一覧の先頭か（「1つ上へ」を disable する）。 */
  isFirst: boolean;
  /** 選択中チケットが一覧の末尾か（「1つ下へ」「末尾へ」を disable する）。 */
  isLast: boolean;
  onMoveUp: () => void;
  onMoveDown: () => void;
  onMoveLast: () => void;
}

const ICON_PROPS = {
  width: 12,
  height: 12,
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 2.4,
  strokeLinecap: 'round' as const,
  'aria-hidden': true,
};

/**
 * 一覧の下の帯。「選択中 X を 1つ上へ / 1つ下へ / 末尾へ」（設計 Ⅳ-F・見本どおり）。
 * ドラッグ&ドロップは段2（キーボードだけで完結し、依存を増やさないボタン案を採用）。
 */
export default function BacklogReorderBar({
  selectedKey,
  isFirst,
  isLast,
  onMoveUp,
  onMoveDown,
  onMoveLast,
}: BacklogReorderBarProps) {
  const hasSelection = selectedKey !== null;
  return (
    <div className="flex items-center gap-2 border-t border-surface-3 bg-surface-1 px-3 py-2 text-xs">
      <span className="text-[var(--color-text-muted)]">
        {hasSelection ? (
          <>
            選択中 <b className="text-[var(--color-text-primary)]">{selectedKey}</b> を
          </>
        ) : (
          '行を選ぶと並び替えられます'
        )}
      </span>
      <button
        type="button"
        onClick={onMoveUp}
        disabled={!hasSelection || isFirst}
        className="inline-flex items-center gap-1 rounded border border-surface-3 px-2 py-1 font-medium text-[var(--color-text-secondary)] hover:bg-surface-2 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
      >
        <svg {...ICON_PROPS}>
          <path d="m5 12 7-7 7 7" />
          <path d="M12 19V5" />
        </svg>
        1 つ上へ
      </button>
      <button
        type="button"
        onClick={onMoveDown}
        disabled={!hasSelection || isLast}
        className="inline-flex items-center gap-1 rounded border border-surface-3 px-2 py-1 font-medium text-[var(--color-text-secondary)] hover:bg-surface-2 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
      >
        <svg {...ICON_PROPS}>
          <path d="M12 5v14" />
          <path d="m19 12-7 7-7-7" />
        </svg>
        1 つ下へ
      </button>
      <button
        type="button"
        onClick={onMoveLast}
        disabled={!hasSelection || isLast}
        className="inline-flex items-center gap-1 rounded border border-surface-3 px-2 py-1 font-medium text-[var(--color-text-secondary)] hover:bg-surface-2 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
      >
        <svg {...ICON_PROPS}>
          <path d="m7 6 5 5 5-5" />
          <path d="m7 13 5 5 5-5" />
        </svg>
        末尾へ
      </button>
      <span className="ml-auto text-[var(--color-text-muted)]">
        アーカイブでは出さない（並び替えは現役の兄弟に限る）
      </span>
    </div>
  );
}
