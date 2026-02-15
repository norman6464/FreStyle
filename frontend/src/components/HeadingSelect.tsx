interface HeadingSelectProps {
  onHeading: (level: number) => void;
}

export default function HeadingSelect({ onHeading }: HeadingSelectProps) {
  return (
    <select
      aria-label="見出し"
      className="text-xs bg-transparent text-[var(--color-text-faint)] border border-[var(--color-surface-3)] rounded px-1 py-0.5 focus:outline-none focus:border-primary-500"
      defaultValue="0"
      onChange={(e) => onHeading(Number(e.target.value))}
    >
      <option value="0">標準</option>
      <option value="1">見出し1</option>
      <option value="2">見出し2</option>
      <option value="3">見出し3</option>
    </select>
  );
}
