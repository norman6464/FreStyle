import { describe, it, expect, vi, beforeEach } from 'vitest';
import NoteRepository from '../NoteRepository';
import apiClient from '../../lib/axios';

vi.mock('../../lib/axios');

describe('NoteRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchNotes は backend のレスポンス配列をそのまま返す', async () => {
    const apiResponse = [
      { id: 21, userId: 1, title: 'テスト', content: '', isPublic: false, isPinned: false, createdAt: '2026-04-28T08:00:00Z', updatedAt: '2026-04-28T09:00:00Z' },
    ];
    vi.mocked(apiClient.get).mockResolvedValue({ data: apiResponse });

    const result = await NoteRepository.fetchNotes();
    expect(apiClient.get).toHaveBeenCalledWith('/api/v2/notes');
    expect(result).toEqual(apiResponse);
  });

  it('fetchNotes は配列以外のレスポンスを空配列にフォールバックする', async () => {
    vi.mocked(apiClient.get).mockResolvedValue({ data: null });
    expect(await NoteRepository.fetchNotes()).toEqual([]);
  });

  it('createNote は POST し backend レスポンスをそのまま返す', async () => {
    const apiResponse = { id: 99, userId: 1, title: '新規', content: '', isPublic: false, isPinned: false, createdAt: '2026-04-28T00:00:00Z', updatedAt: '2026-04-28T00:00:00Z' };
    vi.mocked(apiClient.post).mockResolvedValue({ data: apiResponse });

    const result = await NoteRepository.createNote('新規');
    expect(apiClient.post).toHaveBeenCalledWith('/api/v2/notes', { title: '新規' });
    expect(result).toEqual(apiResponse);
  });

  it('updateNote は noteId(number) を URL に埋め込んで PUT する', async () => {
    vi.mocked(apiClient.put).mockResolvedValue({ data: undefined });

    await NoteRepository.updateNote(21, { title: '更新', content: '内容', isPinned: true });

    expect(apiClient.put).toHaveBeenCalledWith('/api/v2/notes/21', { title: '更新', content: '内容', isPinned: true });
  });

  it('deleteNote は noteId(number) を URL に埋め込んで DELETE する', async () => {
    vi.mocked(apiClient.delete).mockResolvedValue({ data: undefined });

    await NoteRepository.deleteNote(21);
    expect(apiClient.delete).toHaveBeenCalledWith('/api/v2/notes/21');
  });
});
