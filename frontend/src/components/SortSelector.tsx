import { SORT_OPTIONS } from '../constants/sortOptions';
import type { SortOption } from '../constants/sortOptions';

interface SortSelectorProps<T extends string = SortOption> {
  options?: { value: T; label: string }[];
  selected: T;
  onChange: (sort: T) => void;
}

export type { SortOption };

export default function SortSelector<T extends string = SortOption>({ options = SORT_OPTIONS as { value: T; label: string }[], selected, onChange }: SortSelectorProps<T>) {
  return (
    <div className="flex gap-1.5">
      {options.map(({ value, label }) => {
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
