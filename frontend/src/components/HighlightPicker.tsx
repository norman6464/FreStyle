interface HighlightPickerProps {
  onSelectHighlight: (color: string) => void;
}

const HIGHLIGHTS = [
  { color: '#fecaca', label: '赤ハイライト' },
  { color: '#fed7aa', label: 'オレンジハイライト' },
  { color: '#fef08a', label: '黄ハイライト' },
  { color: '#bbf7d0', label: '緑ハイライト' },
  { color: '#bfdbfe', label: '青ハイライト' },
  { color: '#e9d5ff', label: '紫ハイライト' },
];

export default function HighlightPicker({ onSelectHighlight }: HighlightPickerProps) {
  return (
    <div className="flex items-center gap-1">
      {HIGHLIGHTS.map(({ color, label }) => (
        <button
          key={color}
          type="button"
          aria-label={label}
          className="w-5 h-5 rounded border border-[var(--color-surface-3)] hover:scale-110 transition-transform"
          style={{ backgroundColor: color }}
          onClick={() => onSelectHighlight(color)}
        />
      ))}
      <button
        type="button"
        aria-label="ハイライトをリセット"
        className="w-5 h-5 rounded border border-[var(--color-surface-3)] hover:scale-110 transition-transform flex items-center justify-center bg-[var(--color-surface-1)]"
        onClick={() => onSelectHighlight('')}
      >
        <svg width="10" height="10" viewBox="0 0 10 10" className="text-[var(--color-text-secondary)]">
          <line x1="2" y1="2" x2="8" y2="8" stroke="currentColor" strokeWidth="1.5" />
          <line x1="8" y1="2" x2="2" y2="8" stroke="currentColor" strokeWidth="1.5" />
        </svg>
      </button>
    </div>
  );
}
