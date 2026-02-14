const CHARS_PER_MINUTE = 400;

export interface NoteStats {
  charCount: number;
  readingTimeMin: number;
}

export function getNoteStats(content: string): NoteStats {
  const cleaned = content.replace(/[\s\n]/g, '');
  const charCount = cleaned.length;
  const readingTimeMin = charCount === 0 ? 0 : Math.max(1, Math.round(charCount / CHARS_PER_MINUTE));

  return { charCount, readingTimeMin };
}
