import { tiptapToPlainText } from './tiptapToPlainText';

export interface NoteStats {
  charCount: number;
  lineCount: number;
}

export function getNoteStats(content: string): NoteStats {
  const plainText = tiptapToPlainText(content);
  const cleaned = plainText.replace(/[\s\n]/g, '');
  const charCount = cleaned.length;
  const lineCount = plainText === '' ? 0 : plainText.split('\n').length;

  return { charCount, lineCount };
}
