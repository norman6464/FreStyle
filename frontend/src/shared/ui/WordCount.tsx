interface WordCountProps {
  charCount: number;
}

export default function WordCount({ charCount }: WordCountProps) {
  return (
    <div className="px-8 py-1 text-[10px] text-[var(--color-text-faint)] text-right">
      {charCount}文字
    </div>
  );
}
