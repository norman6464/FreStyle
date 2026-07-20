/*
 * entities/note の Public API。
 *
 * 外から使ってよいものだけを名前付きで re-export する（FSD 公式仕様。`export *` は使わない）。
 * この Slice の内部ファイルを直接 import してはいけない。
 */

export { default as NoteRepository } from './api/noteRepository';
export { SessionNoteRepository } from './api/sessionNoteRepository';

export type {
  Note,
  SessionNote,
} from './model/types';

export { default as NoteListItem } from './ui/NoteListItem';

export { getNoteStats } from './lib/noteStats';
