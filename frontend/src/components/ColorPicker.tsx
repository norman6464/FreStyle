interface ColorPickerProps {
  onSelectColor: (color: string) => void;
}

const COLORS = [
  { color: '#ef4444', label: '赤' },
  { color: '#f97316', label: 'オレンジ' },
  { color: '#eab308', label: '黄' },
  { color: '#22c55e', label: '緑' },
  { color: '#3b82f6', label: '青' },
  { color: '#a855f7', label: '紫' },
];

export default function ColorPicker({ onSelectColor }: ColorPickerProps) {
  return (
    <div className="flex items-center gap-1">
      {COLORS.map(({ color, label }) => (
        <button
          key={color}
          type="button"
          aria-label={label}
          className="w-5 h-5 rounded-full border border-[var(--color-surface-3)] hover:scale-110 transition-transform"
          style={{ backgroundColor: color }}
          onClick={() => onSelectColor(color)}
        />
      ))}
      <button
        type="button"
        aria-label="色をリセット"
        className="w-5 h-5 rounded-full border border-[var(--color-surface-3)] hover:scale-110 transition-transform flex items-center justify-center bg-[var(--color-surface-1)]"
        onClick={() => onSelectColor('')}
      >
        <svg width="10" height="10" viewBox="0 0 10 10" className="text-[var(--color-text-secondary)]">
          <line x1="2" y1="2" x2="8" y2="8" stroke="currentColor" strokeWidth="1.5" />
          <line x1="8" y1="2" x2="2" y2="8" stroke="currentColor" strokeWidth="1.5" />
        </svg>
      </button>
    </div>
  );
}
