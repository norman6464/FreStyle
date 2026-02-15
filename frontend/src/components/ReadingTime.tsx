interface ReadingTimeProps {
  charCount: number;
}

const CHARS_PER_MINUTE = 500;

export default function ReadingTime({ charCount }: ReadingTimeProps) {
  const minutes = charCount === 0 ? 0 : Math.max(1, Math.ceil(charCount / CHARS_PER_MINUTE));

  return (
    <span className="text-[10px] text-[var(--color-text-faint)]">
      約{minutes}分
    </span>
  );
}
