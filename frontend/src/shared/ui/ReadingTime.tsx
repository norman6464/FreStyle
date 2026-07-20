interface ReadingTimeProps {
  charCount: number;
}

const CHARS_PER_MINUTE = 500;

export default function ReadingTime({ charCount }: ReadingTimeProps) {
  // 「約0分」は情報量ゼロの表示になるため、本文が無いときは描画しない(FRESTYLE-64)。
  if (charCount === 0) return null;

  const minutes = Math.max(1, Math.ceil(charCount / CHARS_PER_MINUTE));

  return (
    <span className="text-[10px] text-[var(--color-text-faint)]">
      約{minutes}分
    </span>
  );
}
