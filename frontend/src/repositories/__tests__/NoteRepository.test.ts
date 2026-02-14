import { describe, it, expect, vi, beforeEach } from 'vitest';
import NoteRepository from '../NoteRepository';
import apiClient from '../../lib/axios';

vi.mock('../../lib/axios');

describe('NoteRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchNotesがGETリクエストを送信する', async () => {
    const mockNotes = [
      { noteId: 'note-1', userId: 1, title: 'テスト', content: '', isPinned: false, createdAt: 1000, updatedAt: 2000 },
    ];
    vi.mocked(apiClient.get).mockResolvedValue({ data: mockNotes });

    const result = await NoteRepository.fetchNotes();

    expect(apiClient.get).toHaveBeenCalledWith('/api/notes');
    expect(result).toEqual(mockNotes);
  });

  it('createNoteがPOSTリクエストを送信する', async () => {
    const mockNote = { noteId: 'note-new', userId: 1, title: '新しいノート', content: '', isPinned: false, createdAt: 1000, updatedAt: 1000 };
    vi.mocked(apiClient.post).mockResolvedValue({ data: mockNote });

    const result = await NoteRepository.createNote('新しいノート');

    expect(apiClient.post).toHaveBeenCalledWith('/api/notes', { title: '新しいノート' });
    expect(result.title).toBe('新しいノート');
  });

  it('updateNoteがPUTリクエストを送信する', async () => {
    vi.mocked(apiClient.put).mockResolvedValue({ data: undefined });

    await NoteRepository.updateNote('note-1', { title: '更新', content: '内容', isPinned: false });

    expect(apiClient.put).toHaveBeenCalledWith('/api/notes/note-1', { title: '更新', content: '内容', isPinned: false });
  });

  it('deleteNoteがDELETEリクエストを送信する', async () => {
    vi.mocked(apiClient.delete).mockResolvedValue({ data: undefined });

    await NoteRepository.deleteNote('note-1');

    expect(apiClient.delete).toHaveBeenCalledWith('/api/notes/note-1');
  });
});
