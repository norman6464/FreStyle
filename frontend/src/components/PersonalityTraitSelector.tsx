interface PersonalityTraitSelectorProps {
  options: string[];
  selected: string[];
  onToggle: (trait: string) => void;
  label?: string;
}

export default function PersonalityTraitSelector({
  options,
  selected,
  onToggle,
  label,
}: PersonalityTraitSelectorProps) {
  return (
    <div>
      {label && (
        <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
          {label}
        </label>
      )}
      <div className="flex flex-wrap gap-1.5">
        {options.map((trait) => (
          <button
            key={trait}
            type="button"
            onClick={() => onToggle(trait)}
            className={`px-3 py-1.5 rounded-full text-xs font-medium transition-colors ${
              selected.includes(trait)
                ? 'bg-primary-500 text-white'
                : 'bg-surface-3 text-[var(--color-text-secondary)] hover:bg-surface-3'
            }`}
          >
            {trait}
          </button>
        ))}
      </div>
    </div>
  );
}
