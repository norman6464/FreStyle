interface DifficultyFilterProps {
  selected: string | null;
  onChange: (difficulty: string | null) => void;
}

const DIFFICULTIES = [
  { value: null, label: '全レベル', style: 'text-[var(--color-text-secondary)] bg-surface-2' },
  { value: '初級', label: '初級', style: 'text-emerald-400 bg-emerald-900/30' },
  { value: '中級', label: '中級', style: 'text-amber-400 bg-amber-900/30' },
  { value: '上級', label: '上級', style: 'text-rose-400 bg-rose-900/30' },
] as const;

export default function DifficultyFilter({ selected, onChange }: DifficultyFilterProps) {
  return (
    <div className="flex gap-1.5">
      {DIFFICULTIES.map(({ value, label, style }) => {
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
