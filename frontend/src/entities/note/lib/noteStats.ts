export interface NoteStats {
  charCount: number;
  lineCount: number;
}

/**
 * Markdown 文字列から表示用の統計を計算する。
 * 文字カウントは空白 / 改行を除いた純粋な文字数。
 */
export function getNoteStats(content: string): NoteStats {
  const cleaned = content.replace(/[\s\n]/g, '');
  const charCount = cleaned.length;
  const lineCount = content === '' ? 0 : content.split('\n').length;

  return { charCount, lineCount };
}
