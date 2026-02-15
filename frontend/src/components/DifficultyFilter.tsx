import { DIFFICULTY_OPTIONS } from '../constants/difficultyStyles';

interface DifficultyFilterProps {
  selected: string | null;
  onChange: (difficulty: string | null) => void;
}

export default function DifficultyFilter({ selected, onChange }: DifficultyFilterProps) {
  return (
    <div className="flex gap-1.5">
      {DIFFICULTY_OPTIONS.map(({ value, label, style }) => {
        const isActive = selected === value;
        return (
          <button
            key={label}
            onClick={() => onChange(value)}
            className={`text-[10px] font-medium px-2.5 py-1 rounded-full transition-colors ${
              isActive
                ? style
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
