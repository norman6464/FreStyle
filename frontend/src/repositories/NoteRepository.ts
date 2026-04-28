import apiClient from '../lib/axios';
import type { Note } from '../types';

// Go バックエンドの Note レスポンスは { id, userId, title, content, isPublic, createdAt, updatedAt }。
// フロント Note 型は { noteId: string, ..., isPinned: boolean, createdAt: number(ms), updatedAt: number(ms) } を期待するため
// Repository 層で正規化する（CLAUDE.md の Mapper 集約方針に準拠）。
interface ApiNote {
  id: number;
  userId: number;
  title?: string;
  content?: string;
  isPublic?: boolean;
  isPinned?: boolean;
  createdAt?: string | number;
  updatedAt?: string | number;
}

function toMillis(value: string | number | undefined): number {
  if (value == null) return Date.now();
  if (typeof value === 'number') return value;
  const parsed = Date.parse(value);
  return Number.isNaN(parsed) ? Date.now() : parsed;
}

function toNote(api: ApiNote): Note {
  return {
    noteId: String(api.id),
    userId: api.userId,
    title: api.title ?? '',
    content: api.content ?? '',
    isPinned: api.isPinned ?? api.isPublic ?? false,
    createdAt: toMillis(api.createdAt),
    updatedAt: toMillis(api.updatedAt),
  };
}

const NoteRepository = {
  async fetchNotes(): Promise<Note[]> {
    const res = await apiClient.get<ApiNote[]>('/api/v2/notes');
    const list = Array.isArray(res.data) ? res.data : [];
    return list.map(toNote);
  },

  async createNote(title: string): Promise<Note> {
    const res = await apiClient.post<ApiNote>('/api/v2/notes', { title });
    return toNote(res.data);
  },

  async updateNote(noteId: string, data: { title: string; content: string; isPinned: boolean }): Promise<void> {
    // backend の domain は isPublic フィールドを持つ。互換のため両方送る。
    await apiClient.put(`/api/v2/notes/${noteId}`, { ...data, isPublic: data.isPinned });
  },

  async deleteNote(noteId: string): Promise<void> {
    await apiClient.delete(`/api/v2/notes/${noteId}`);
  },
};

export default NoteRepository;
