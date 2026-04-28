import { describe, it, expect, vi, beforeEach } from 'vitest';
import NoteRepository from '../NoteRepository';
import apiClient from '../../lib/axios';

vi.mock('../../lib/axios');

describe('NoteRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchNotes は backend の id (number) を noteId (string) に正規化する', async () => {
    const apiResponse = [
      {
        id: 21,
        userId: 1,
        title: 'テスト',
        content: '',
        isPublic: false,
        createdAt: '2026-04-28T08:00:00Z',
        updatedAt: '2026-04-28T09:00:00Z',
      },
    ];
    vi.mocked(apiClient.get).mockResolvedValue({ data: apiResponse });

    const result = await NoteRepository.fetchNotes();

    expect(apiClient.get).toHaveBeenCalledWith('/api/v2/notes');
    expect(result).toHaveLength(1);
    expect(result[0].noteId).toBe('21');
    expect(result[0].title).toBe('テスト');
    expect(result[0].isPinned).toBe(false);
    expect(typeof result[0].createdAt).toBe('number');
  });

  it('fetchNotes は配列以外のレスポンスを空配列にフォールバックする', async () => {
    vi.mocked(apiClient.get).mockResolvedValue({ data: null });
    const result = await NoteRepository.fetchNotes();
    expect(result).toEqual([]);
  });

  it('createNote は backend レスポンスの id を noteId に変換した Note を返す', async () => {
    const apiResponse = {
      id: 99,
      userId: 1,
      title: '新しいノート',
      content: '',
      isPublic: true,
      createdAt: '2026-04-28T00:00:00Z',
      updatedAt: '2026-04-28T00:00:00Z',
    };
    vi.mocked(apiClient.post).mockResolvedValue({ data: apiResponse });

    const result = await NoteRepository.createNote('新しいノート');

    expect(apiClient.post).toHaveBeenCalledWith('/api/v2/notes', { title: '新しいノート' });
    expect(result.noteId).toBe('99');
    expect(result.title).toBe('新しいノート');
    expect(result.isPinned).toBe(true);
  });

  it('updateNote は noteId を URL に埋め込み isPublic も互換で送信する', async () => {
    vi.mocked(apiClient.put).mockResolvedValue({ data: undefined });

    await NoteRepository.updateNote('21', { title: '更新', content: '内容', isPinned: true });

    expect(apiClient.put).toHaveBeenCalledWith('/api/v2/notes/21', {
      title: '更新',
      content: '内容',
      isPinned: true,
      isPublic: true,
    });
  });

  it('deleteNote は noteId を URL に埋め込んで DELETE を投げる', async () => {
    vi.mocked(apiClient.delete).mockResolvedValue({ data: undefined });

    await NoteRepository.deleteNote('21');

    expect(apiClient.delete).toHaveBeenCalledWith('/api/v2/notes/21');
  });
});
