export type SortOption = 'default' | 'difficulty-asc' | 'difficulty-desc' | 'name';

interface SortSelectorProps {
  selected: SortOption;
  onChange: (sort: SortOption) => void;
}

const SORT_OPTIONS: { value: SortOption; label: string }[] = [
  { value: 'default', label: 'デフォルト' },
  { value: 'difficulty-asc', label: '難易度↑' },
  { value: 'difficulty-desc', label: '難易度↓' },
  { value: 'name', label: '名前順' },
];

export default function SortSelector({ selected, onChange }: SortSelectorProps) {
  return (
    <div className="flex gap-1.5">
      {SORT_OPTIONS.map(({ value, label }) => {
        const isActive = selected === value;
        return (
          <button
            key={value}
            onClick={() => onChange(value)}
            className={`text-[10px] font-medium px-2.5 py-1 rounded-full transition-colors ${
              isActive
                ? 'text-[var(--color-text-secondary)] bg-surface-2'
                : 'text-[var(--color-text-muted)] bg-transparent hover:bg-surface-2'
            }`}
          >
            {label}
          </button>
        );
      })}
    </div>
  );
}
