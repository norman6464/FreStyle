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

  it('createNoteで新しいノートを作成しローカル追加する', async () => {
    const newNote = { noteId: 'note-3', userId: 1, title: '新規', content: '', isPinned: false, createdAt: 4000, updatedAt: 4000 };
    vi.mocked(NoteRepository.createNote).mockResolvedValue(newNote);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.createNote('新規');
    });

    expect(NoteRepository.createNote).toHaveBeenCalledWith('新規');
    expect(result.current.notes).toHaveLength(1);
    expect(result.current.notes[0].noteId).toBe('note-3');
  });

  it('deleteNoteでノートを削除しローカル除外する', async () => {
    vi.mocked(NoteRepository.deleteNote).mockResolvedValue(undefined);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });
    expect(result.current.notes).toHaveLength(2);

    await act(async () => {
      await result.current.deleteNote('note-1');
    });

    expect(NoteRepository.deleteNote).toHaveBeenCalledWith('note-1');
    expect(result.current.notes).toHaveLength(1);
    expect(result.current.notes[0].noteId).toBe('note-2');
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

  // エッジケーステスト

  it('fetchNotesエラー時にnotesが空のまま', async () => {
    vi.mocked(NoteRepository.fetchNotes).mockRejectedValue(new Error('ネットワークエラー'));

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    expect(result.current.notes).toEqual([]);
    expect(result.current.loading).toBe(false);
  });

  it('createNoteエラー時にnullを返す', async () => {
    vi.mocked(NoteRepository.createNote).mockRejectedValue(new Error('作成エラー'));

    const { result } = renderHook(() => useNotes());
    let returnValue: unknown;

    await act(async () => {
      returnValue = await result.current.createNote('テスト');
    });

    expect(returnValue).toBeNull();
  });

  it('createNote成功後にselectedNoteIdが更新される', async () => {
    const newNote = { noteId: 'new-id', userId: 1, title: 'テスト', content: '', isPinned: false, createdAt: 5000, updatedAt: 5000 };
    vi.mocked(NoteRepository.createNote).mockResolvedValue(newNote);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.createNote('テスト');
    });

    expect(result.current.selectedNoteId).toBe('new-id');
  });

  it('updateNoteでnotes配列がローカル更新される', async () => {
    vi.mocked(NoteRepository.updateNote).mockResolvedValue(undefined);
    vi.mocked(NoteRepository.fetchNotes).mockResolvedValue(mockNotes);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    await act(async () => {
      await result.current.updateNote('note-1', { title: '更新済み', content: '新内容', isPinned: false });
    });

    expect(result.current.notes.find(n => n.noteId === 'note-1')?.title).toBe('更新済み');
    expect(result.current.notes.find(n => n.noteId === 'note-1')?.content).toBe('新内容');
  });

  it('選択中のノートを削除するとselectedNoteIdがnullになる', async () => {
    vi.mocked(NoteRepository.deleteNote).mockResolvedValue(undefined);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    await act(async () => {
      result.current.selectNote('note-1');
    });

    await act(async () => {
      await result.current.deleteNote('note-1');
    });

    expect(result.current.selectedNoteId).toBeNull();
    expect(result.current.notes).toHaveLength(1);
  });

  it('selectNoteにnullを渡すと選択解除される', async () => {
    const { result } = renderHook(() => useNotes());

    await act(async () => {
      result.current.selectNote('note-1');
    });
    expect(result.current.selectedNoteId).toBe('note-1');

    await act(async () => {
      result.current.selectNote(null);
    });
    expect(result.current.selectedNoteId).toBeNull();
  });

  it('loadingはfetchNotes中にtrueになりfalseに戻る', async () => {
    let resolveFn: (value: typeof mockNotes) => void;
    vi.mocked(NoteRepository.fetchNotes).mockReturnValue(
      new Promise((resolve) => { resolveFn = resolve; })
    );

    const { result } = renderHook(() => useNotes());
    expect(result.current.loading).toBe(false);

    let fetchPromise: Promise<void>;
    act(() => {
      fetchPromise = result.current.fetchNotes();
    });

    expect(result.current.loading).toBe(true);

    await act(async () => {
      resolveFn!(mockNotes);
      await fetchPromise!;
    });

    expect(result.current.loading).toBe(false);
  });
});
