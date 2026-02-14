import apiClient from '../lib/axios';
import type { Note } from '../types';

const NoteRepository = {
  async fetchNotes(): Promise<Note[]> {
    const res = await apiClient.get('/api/notes');
    return res.data;
  },

  async createNote(title: string): Promise<Note> {
    const res = await apiClient.post('/api/notes', { title });
    return res.data;
  },

  async updateNote(noteId: string, data: { title: string; content: string; isPinned: boolean }): Promise<void> {
    await apiClient.put(`/api/notes/${noteId}`, data);
  },

  async deleteNote(noteId: string): Promise<void> {
    await apiClient.delete(`/api/notes/${noteId}`);
  },
};

export default NoteRepository;
