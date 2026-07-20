/**
 * FilterChip — 一覧の絞り込みに使うピル型トグルボタン。
 *
 * コース一覧のカテゴリ絞り込み(FRESTYLE-68)と演習一覧の言語絞り込み(FRESTYLE-101)で共用。
 * active 時は activeClass(カテゴリ色などの badgeClass)、未指定なら brand 色。
 * 選択状態は aria-pressed で公開する。
 */
export interface FilterChipProps {
  label: string;
  active: boolean;
  /** active 時に適用する色クラス(例: カテゴリの badgeClass)。未指定なら brand 色。 */
  activeClass?: string;
  onClick: () => void;
}

export default function FilterChip({ label, active, activeClass, onClick }: FilterChipProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={`text-xs font-medium px-3 py-1 rounded-full border transition-colors ${
        active
          ? (activeClass ?? 'bg-brand-500/15 text-brand-600 border-brand-500/30')
          : 'bg-surface-1 text-[var(--color-text-muted)] border-surface-3 hover:bg-surface-2'
      }`}
    >
      {label}
    </button>
  );
}
