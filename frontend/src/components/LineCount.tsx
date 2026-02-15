interface LineCountProps {
  lineCount: number;
}

export default function LineCount({ lineCount }: LineCountProps) {
  return (
    <span className="text-[10px] text-[var(--color-text-faint)]">
      {lineCount}行
    </span>
  );
}
