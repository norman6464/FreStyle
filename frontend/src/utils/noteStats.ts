import { tiptapToPlainText } from './tiptapToPlainText';

export interface NoteStats {
  charCount: number;
}

export function getNoteStats(content: string): NoteStats {
  const plainText = tiptapToPlainText(content);
  const cleaned = plainText.replace(/[\s\n]/g, '');
  const charCount = cleaned.length;

  return { charCount };
}
