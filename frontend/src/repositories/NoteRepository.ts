import apiClient from '../lib/axios';
import { NOTES } from '../constants/apiRoutes';
import type { Note } from '../types';

/**
 * Note 関連 API の薄いラッパ。
 * フロント Note 型は backend domain.Note と 1:1 なのでマッパーは不要（DOP 整合性方針）。
 */
const NoteRepository = {
  async fetchNotes(): Promise<Note[]> {
    const res = await apiClient.get<Note[]>(NOTES.list);
    return Array.isArray(res.data) ? res.data : [];
  },

  async createNote(title: string): Promise<Note> {
    const res = await apiClient.post<Note>(NOTES.list, { title });
    return res.data;
  },

  async updateNote(noteId: number, data: { title: string; content: string; isPinned: boolean }): Promise<void> {
    await apiClient.put(NOTES.byId(noteId), data);
  },

  async deleteNote(noteId: number): Promise<void> {
    await apiClient.delete(NOTES.byId(noteId));
  },
};

export default NoteRepository;
