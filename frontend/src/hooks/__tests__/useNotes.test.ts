import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useNotes } from '../useNotes';
import NoteRepository from '../../repositories/NoteRepository';
import type { Note } from '../../types';

vi.mock('../../repositories/NoteRepository');

// backend `domain.Note` と 1:1。createdAt / updatedAt は RFC3339 string。
const mockNotes: Note[] = [
  {
    id: 1,
    userId: 1,
    title: 'ノート1',
    content: '内容1',
    isPublic: false,
    isPinned: true,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-03T00:00:00Z',
  },
  {
    id: 2,
    userId: 1,
    title: 'ノート2',
    content: '内容2',
    isPublic: false,
    isPinned: false,
    createdAt: '2026-01-02T00:00:00Z',
    updatedAt: '2026-01-02T00:00:00Z',
  },
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
    const newNote: Note = {
      id: 3,
      userId: 1,
      title: '新規',
      content: '',
      isPublic: false,
      isPinned: false,
      createdAt: '2026-01-04T00:00:00Z',
      updatedAt: '2026-01-04T00:00:00Z',
    };
    vi.mocked(NoteRepository.createNote).mockResolvedValue(newNote);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.createNote('新規');
    });

    expect(NoteRepository.createNote).toHaveBeenCalledWith('新規');
    expect(result.current.notes).toHaveLength(1);
    expect(result.current.notes[0].id).toBe(3);
  });

  it('deleteNoteでノートを削除しローカル除外する', async () => {
    vi.mocked(NoteRepository.deleteNote).mockResolvedValue(undefined);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });
    expect(result.current.notes).toHaveLength(2);

    await act(async () => {
      await result.current.deleteNote(1);
    });

    expect(NoteRepository.deleteNote).toHaveBeenCalledWith(1);
    expect(result.current.notes).toHaveLength(1);
    expect(result.current.notes[0].id).toBe(2);
  });

  it('selectedNoteIdの初期値がnullである', () => {
    const { result } = renderHook(() => useNotes());
    expect(result.current.selectedNoteId).toBeNull();
  });

  it('selectNoteでノートを選択できる', () => {
    const { result } = renderHook(() => useNotes());

    act(() => {
      result.current.selectNote(1);
    });

    expect(result.current.selectedNoteId).toBe(1);
  });

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
    const newNote: Note = {
      id: 99,
      userId: 1,
      title: 'テスト',
      content: '',
      isPublic: false,
      isPinned: false,
      createdAt: '2026-01-05T00:00:00Z',
      updatedAt: '2026-01-05T00:00:00Z',
    };
    vi.mocked(NoteRepository.createNote).mockResolvedValue(newNote);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.createNote('テスト');
    });

    expect(result.current.selectedNoteId).toBe(99);
  });

  it('updateNoteでnotes配列がローカル更新される', async () => {
    vi.mocked(NoteRepository.updateNote).mockResolvedValue(undefined);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    await act(async () => {
      await result.current.updateNote(1, { title: '更新済み', content: '新内容', isPinned: false });
    });

    expect(result.current.notes.find((n) => n.id === 1)?.title).toBe('更新済み');
    expect(result.current.notes.find((n) => n.id === 1)?.content).toBe('新内容');
  });

  it('選択中のノートを削除するとselectedNoteIdがnullになる', async () => {
    vi.mocked(NoteRepository.deleteNote).mockResolvedValue(undefined);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    act(() => {
      result.current.selectNote(1);
    });

    await act(async () => {
      await result.current.deleteNote(1);
    });

    expect(result.current.selectedNoteId).toBeNull();
    expect(result.current.notes).toHaveLength(1);
  });

  it('selectNoteにnullを渡すと選択解除される', () => {
    const { result } = renderHook(() => useNotes());

    act(() => {
      result.current.selectNote(1);
    });
    expect(result.current.selectedNoteId).toBe(1);

    act(() => {
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

  it('searchQueryの初期値が空文字である', () => {
    const { result } = renderHook(() => useNotes());
    expect(result.current.searchQuery).toBe('');
  });

  it('setSearchQueryで検索クエリを変更できる', () => {
    const { result } = renderHook(() => useNotes());

    act(() => {
      result.current.setSearchQuery('テスト');
    });

    expect(result.current.searchQuery).toBe('テスト');
  });

  it('filteredNotesがピン留め優先でソートされる', async () => {
    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    expect(result.current.filteredNotes[0].id).toBe(1);
    expect(result.current.filteredNotes[0].isPinned).toBe(true);
  });

  it('filteredNotesが検索クエリでタイトルフィルタリングされる', async () => {
    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    act(() => {
      result.current.setSearchQuery('ノート1');
    });

    expect(result.current.filteredNotes).toHaveLength(1);
    expect(result.current.filteredNotes[0].title).toBe('ノート1');
  });

  it('filteredNotesが検索クエリで内容もフィルタリングする', async () => {
    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    act(() => {
      result.current.setSearchQuery('内容2');
    });

    expect(result.current.filteredNotes).toHaveLength(1);
    expect(result.current.filteredNotes[0].id).toBe(2);
  });

  it('togglePinでノートのピン留め状態をトグルする', async () => {
    vi.mocked(NoteRepository.updateNote).mockResolvedValue(undefined);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    await act(async () => {
      await result.current.togglePin(2);
    });

    const note2 = result.current.notes.find((n) => n.id === 2);
    expect(note2?.isPinned).toBe(true);
    expect(NoteRepository.updateNote).toHaveBeenCalledWith(2, {
      title: 'ノート2',
      content: '内容2',
      isPinned: true,
    });
  });

  it('filteredNotesはピン留め内でも更新日時降順でソートされる', async () => {
    const threeNotes: Note[] = [
      { id: 1, userId: 1, title: 'A', content: '', isPublic: false, isPinned: true, createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z' },
      { id: 2, userId: 1, title: 'B', content: '', isPublic: false, isPinned: true, createdAt: '2026-01-02T00:00:00Z', updatedAt: '2026-01-03T00:00:00Z' },
      { id: 3, userId: 1, title: 'C', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-03T00:00:00Z', updatedAt: '2026-01-05T00:00:00Z' },
    ];
    vi.mocked(NoteRepository.fetchNotes).mockResolvedValue(threeNotes);

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    expect(result.current.filteredNotes[0].id).toBe(2);
    expect(result.current.filteredNotes[1].id).toBe(1);
    expect(result.current.filteredNotes[2].id).toBe(3);
  });

  it('検索クエリが大文字小文字を区別しない', async () => {
    const mixedNotes: Note[] = [
      { id: 1, userId: 1, title: 'Hello World', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-02T00:00:00Z' },
    ];
    vi.mocked(NoteRepository.fetchNotes).mockResolvedValue(mixedNotes);

    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    act(() => { result.current.setSearchQuery('hello'); });
    expect(result.current.filteredNotes).toHaveLength(1);

    act(() => { result.current.setSearchQuery('HELLO'); });
    expect(result.current.filteredNotes).toHaveLength(1);
  });

  it('togglePinエラー時にnotes状態が変更されない', async () => {
    vi.mocked(NoteRepository.updateNote).mockRejectedValue(new Error('更新エラー'));

    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    const originalPinState = result.current.notes.find((n) => n.id === 2)?.isPinned;

    await act(async () => { await result.current.togglePin(2); });

    expect(result.current.notes.find((n) => n.id === 2)?.isPinned).toBe(originalPinState);
  });

  it('togglePinで存在しないIDの場合何も起こらない', async () => {
    vi.mocked(NoteRepository.updateNote).mockResolvedValue(undefined);

    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    await act(async () => { await result.current.togglePin(9999); });

    expect(NoteRepository.updateNote).not.toHaveBeenCalled();
    expect(result.current.notes).toHaveLength(2);
  });

  it('検索クエリが空文字の場合全ノートが表示される', async () => {
    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    act(() => { result.current.setSearchQuery('ノート1'); });
    expect(result.current.filteredNotes).toHaveLength(1);

    act(() => { result.current.setSearchQuery(''); });
    expect(result.current.filteredNotes).toHaveLength(2);
  });

  it('検索でヒットしない場合は空配列を返す', async () => {
    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    act(() => { result.current.setSearchQuery('存在しないクエリ'); });
    expect(result.current.filteredNotes).toHaveLength(0);
  });

  it('selectedNoteが選択中のノートオブジェクトを返す', async () => {
    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    act(() => { result.current.selectNote(1); });
    expect(result.current.selectedNote?.id).toBe(1);
    expect(result.current.selectedNote?.title).toBe('ノート1');
  });

  it('selectedNoteが未選択時にnullを返す', async () => {
    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    expect(result.current.selectedNote).toBeNull();
  });

  it('deleteTargetIdの初期値がnullである', () => {
    const { result } = renderHook(() => useNotes());
    expect(result.current.deleteTargetId).toBeNull();
  });

  it('requestDeleteでdeleteTargetIdが設定される', () => {
    const { result } = renderHook(() => useNotes());

    act(() => { result.current.requestDelete(1); });
    expect(result.current.deleteTargetId).toBe(1);
  });

  it('confirmDeleteでノートが削除されdeleteTargetIdがリセットされる', async () => {
    vi.mocked(NoteRepository.deleteNote).mockResolvedValue(undefined);

    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    act(() => { result.current.requestDelete(1); });
    expect(result.current.deleteTargetId).toBe(1);

    await act(async () => { await result.current.confirmDelete(); });

    expect(result.current.deleteTargetId).toBeNull();
    expect(NoteRepository.deleteNote).toHaveBeenCalledWith(1);
    expect(result.current.notes).toHaveLength(1);
  });

  it('cancelDeleteでdeleteTargetIdがリセットされる', () => {
    const { result } = renderHook(() => useNotes());

    act(() => { result.current.requestDelete(1); });
    expect(result.current.deleteTargetId).toBe(1);

    act(() => { result.current.cancelDelete(); });
    expect(result.current.deleteTargetId).toBeNull();
  });

  it('confirmDeleteで選択中ノート削除後に次のノートが自動選択される', async () => {
    const threeNotes: Note[] = [
      { id: 11, userId: 1, title: 'A', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-03T00:00:00Z' },
      { id: 12, userId: 1, title: 'B', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-02T00:00:00Z', updatedAt: '2026-01-02T00:00:00Z' },
      { id: 13, userId: 1, title: 'C', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-03T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z' },
    ];
    vi.mocked(NoteRepository.fetchNotes).mockResolvedValue(threeNotes);
    vi.mocked(NoteRepository.deleteNote).mockResolvedValue(undefined);

    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    act(() => { result.current.selectNote(11); });
    act(() => { result.current.requestDelete(11); });

    await act(async () => { await result.current.confirmDelete(); });

    expect(result.current.selectedNoteId).toBe(12);
  });

  it('confirmDeleteで最後のノート削除後に前のノートが自動選択される', async () => {
    const threeNotes: Note[] = [
      { id: 11, userId: 1, title: 'A', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-03T00:00:00Z' },
      { id: 12, userId: 1, title: 'B', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-02T00:00:00Z', updatedAt: '2026-01-02T00:00:00Z' },
      { id: 13, userId: 1, title: 'C', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-03T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z' },
    ];
    vi.mocked(NoteRepository.fetchNotes).mockResolvedValue(threeNotes);
    vi.mocked(NoteRepository.deleteNote).mockResolvedValue(undefined);

    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    act(() => { result.current.selectNote(13); });
    act(() => { result.current.requestDelete(13); });

    await act(async () => { await result.current.confirmDelete(); });

    expect(result.current.selectedNoteId).toBe(12);
  });

  it('deleteTargetIdがnullのときconfirmDeleteは何もしない', async () => {
    const { result } = renderHook(() => useNotes());

    await act(async () => { await result.current.confirmDelete(); });
    expect(NoteRepository.deleteNote).not.toHaveBeenCalled();
  });

  it('noteSortの初期値がdefaultである', () => {
    const { result } = renderHook(() => useNotes());
    expect(result.current.noteSort).toBe('default');
  });

  it('setNoteSortでソート順を変更できる', () => {
    const { result } = renderHook(() => useNotes());

    act(() => { result.current.setNoteSort('updated-asc'); });
    expect(result.current.noteSort).toBe('updated-asc');
  });

  it('noteSortがupdated-ascの場合、更新日時昇順でソートされる', async () => {
    const notes: Note[] = [
      { id: 1, userId: 1, title: 'A', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-03T00:00:00Z' },
      { id: 2, userId: 1, title: 'B', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-02T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z' },
      { id: 3, userId: 1, title: 'C', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-03T00:00:00Z', updatedAt: '2026-01-02T00:00:00Z' },
    ];
    vi.mocked(NoteRepository.fetchNotes).mockResolvedValue(notes);

    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    act(() => { result.current.setNoteSort('updated-asc'); });

    expect(result.current.filteredNotes[0].id).toBe(2);
    expect(result.current.filteredNotes[1].id).toBe(3);
    expect(result.current.filteredNotes[2].id).toBe(1);
  });

  it('noteSortがtitleの場合、タイトル順でソートされる', async () => {
    const notes: Note[] = [
      { id: 1, userId: 1, title: 'バナナ', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z' },
      { id: 2, userId: 1, title: 'アップル', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-02T00:00:00Z', updatedAt: '2026-01-02T00:00:00Z' },
      { id: 3, userId: 1, title: 'チェリー', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-03T00:00:00Z', updatedAt: '2026-01-03T00:00:00Z' },
    ];
    vi.mocked(NoteRepository.fetchNotes).mockResolvedValue(notes);

    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    act(() => { result.current.setNoteSort('title'); });

    expect(result.current.filteredNotes[0].title).toBe('アップル');
    expect(result.current.filteredNotes[1].title).toBe('チェリー');
    expect(result.current.filteredNotes[2].title).toBe('バナナ');
  });

  it('noteSortがcreated-descの場合、作成日時降順でソートされる', async () => {
    const notes: Note[] = [
      { id: 1, userId: 1, title: 'A', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-03T00:00:00Z' },
      { id: 2, userId: 1, title: 'B', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-03T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z' },
      { id: 3, userId: 1, title: 'C', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-02T00:00:00Z', updatedAt: '2026-01-02T00:00:00Z' },
    ];
    vi.mocked(NoteRepository.fetchNotes).mockResolvedValue(notes);

    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    act(() => { result.current.setNoteSort('created-desc'); });

    expect(result.current.filteredNotes[0].id).toBe(2);
    expect(result.current.filteredNotes[1].id).toBe(3);
    expect(result.current.filteredNotes[2].id).toBe(1);
  });

  it('noteSortを変更してもピン留めされたノートが常に先頭に来る', async () => {
    const notes: Note[] = [
      { id: 1, userId: 1, title: 'バナナ', content: '', isPublic: false, isPinned: true, createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-04T00:00:00Z' },
      { id: 2, userId: 1, title: 'アップル', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-02T00:00:00Z', updatedAt: '2026-01-03T00:00:00Z' },
      { id: 3, userId: 1, title: 'チェリー', content: '', isPublic: false, isPinned: true, createdAt: '2026-01-03T00:00:00Z', updatedAt: '2026-01-02T00:00:00Z' },
      { id: 4, userId: 1, title: 'オレンジ', content: '', isPublic: false, isPinned: false, createdAt: '2026-01-04T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z' },
    ];
    vi.mocked(NoteRepository.fetchNotes).mockResolvedValue(notes);

    const { result } = renderHook(() => useNotes());
    await act(async () => { await result.current.fetchNotes(); });

    const assertPinnedFirst = () => {
      const filtered = result.current.filteredNotes;
      const firstUnpinnedIndex = filtered.findIndex((note) => !note.isPinned);
      if (firstUnpinnedIndex === -1) return;
      const allAfterUnpinned = filtered.slice(firstUnpinnedIndex).every((note) => !note.isPinned);
      expect(allAfterUnpinned).toBe(true);
    };

    act(() => { result.current.setNoteSort('updated-asc'); });
    assertPinnedFirst();

    act(() => { result.current.setNoteSort('title'); });
    assertPinnedFirst();

    act(() => { result.current.setNoteSort('created-desc'); });
    assertPinnedFirst();
  });

  it('fetchNotes失敗時にerrorが設定される', async () => {
    vi.mocked(NoteRepository.fetchNotes).mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    expect(result.current.error).toBe('ノートの取得に失敗しました');
    expect(result.current.notes).toEqual([]);
  });

  it('updateNote失敗時にerrorが設定される', async () => {
    vi.mocked(NoteRepository.updateNote).mockRejectedValue(new Error('Update failed'));

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    await act(async () => {
      await result.current.updateNote(1, { title: '更新', content: '内容', isPinned: false });
    });

    expect(result.current.error).toBe('ノートの更新に失敗しました');
  });

  it('deleteNote失敗時にerrorが設定される', async () => {
    vi.mocked(NoteRepository.deleteNote).mockRejectedValue(new Error('Delete failed'));

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });

    await act(async () => {
      await result.current.deleteNote(1);
    });

    expect(result.current.error).toBe('ノートの削除に失敗しました');
    expect(result.current.notes).toHaveLength(2);
  });

  it('成功した操作でerrorがクリアされる', async () => {
    vi.mocked(NoteRepository.fetchNotes).mockRejectedValueOnce(new Error('fail'));

    const { result } = renderHook(() => useNotes());

    await act(async () => {
      await result.current.fetchNotes();
    });
    expect(result.current.error).toBe('ノートの取得に失敗しました');

    vi.mocked(NoteRepository.fetchNotes).mockResolvedValue(mockNotes);

    await act(async () => {
      await result.current.fetchNotes();
    });
    expect(result.current.error).toBeNull();
  });
});
