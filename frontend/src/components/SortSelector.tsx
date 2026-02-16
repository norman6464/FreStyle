import { SORT_OPTIONS } from '../constants/sortOptions';
import type { SortOption } from '../constants/sortOptions';

interface SortSelectorProps {
  options?: { value: string; label: string }[];
  selected: string;
  onChange: (sort: string) => void;
}

export type { SortOption };

export default function SortSelector({ options = SORT_OPTIONS, selected, onChange }: SortSelectorProps) {
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
