import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useNotes } from '../useNotes';
import NoteRepository from '../../repositories/NoteRepository';

vi.mock('../../repositories/NoteRepository');

const mockNotes = [
  { noteId: 'note-1', userId: 1, title: 'ノート1', content: '内容1', isPinned: true, createdAt: 1000, updatedAt: 3000 },
  { noteId: 'note-2', userId: 1, title: 'ノート2', content: '内容2', isPinned: false, createdAt: 2000, updatedAt: 2000 },
];

describe('useNotes', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(NoteRepository.fetchNotes).mockResolvedValue(mockNotes);
  });

  it('初期状態でnotesが空配列である', () => {
    const { result } = renderHook(() => useNotes());
    expect(result.current.notes).toEqual([]);
  });

  it('fetchNotesでノート一覧を取得する', async () => {
    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    expect(result.current.notes).toHaveLength(2);
    expect(result.current.notes[0].title).toBe('ノート1');
  });

  it('createNoteで新しいノートを作成する', async () => {
    const newNote = { noteId: 'note-3', userId: 1, title: '新規', content: '', isPinned: false, createdAt: 4000, updatedAt: 4000 };
    vi.mocked(NoteRepository.createNote).mockResolvedValue(newNote);
    vi.mocked(NoteRepository.fetchNotes).mockResolvedValue([...mockNotes, newNote]);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.createNote('新規');
    });

    expect(NoteRepository.createNote).toHaveBeenCalledWith('新規');
  });

  it('deleteNoteでノートを削除する', async () => {
    vi.mocked(NoteRepository.deleteNote).mockResolvedValue(undefined);
    vi.mocked(NoteRepository.fetchNotes)
      .mockResolvedValueOnce(mockNotes)
      .mockResolvedValueOnce([mockNotes[1]]);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    await act(async () => {
      await result.current.deleteNote('note-1');
    });

    expect(NoteRepository.deleteNote).toHaveBeenCalledWith('note-1');
  });

  it('selectedNoteIdの初期値がnullである', () => {
    const { result } = renderHook(() => useNotes());
    expect(result.current.selectedNoteId).toBeNull();
  });

  it('selectNoteでノートを選択できる', async () => {
    const { result } = renderHook(() => useNotes());

    await act(async () => {
      result.current.selectNote('note-1');
    });

    expect(result.current.selectedNoteId).toBe('note-1');
  });
});
