import { render, screen, fireEvent, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import NotesPage from '../NotesPage';
import { useNotes } from '../../hooks/useNotes';
import type { Note } from '../../types';

vi.mock('../../hooks/useNotes');
vi.mock('../../hooks/useBlockEditor', () => ({
  useBlockEditor: () => ({ editor: null }),
}));
const mockShowToast = vi.fn();
vi.mock('../../hooks/useToast', () => ({
  useToast: () => ({ showToast: mockShowToast, toasts: [], removeToast: vi.fn() }),
}));

function makeNote(partial: Partial<Note> & { id: number }): Note {
  return {
    id: partial.id,
    userId: partial.userId ?? 1,
    title: partial.title ?? `ノート${partial.id}`,
    content: partial.content ?? '',
    isPublic: partial.isPublic ?? false,
    isPinned: partial.isPinned ?? false,
    createdAt: partial.createdAt ?? '2026-01-01T00:00:00Z',
    updatedAt: partial.updatedAt ?? '2026-01-02T00:00:00Z',
  };
}

const mockUseNotes = {
  notes: [] as Note[],
  filteredNotes: [] as Note[],
  selectedNoteId: null as number | null,
  selectedNote: null as Note | null,
  loading: false,
  error: null as string | null,
  searchQuery: '',
  setSearchQuery: vi.fn(),
  fetchNotes: vi.fn(),
  createNote: vi.fn(),
  updateNote: vi.fn(),
  deleteNote: vi.fn(),
  selectNote: vi.fn(),
  togglePin: vi.fn(),
  deleteTargetId: null as number | null,
  requestDelete: vi.fn(),
  confirmDelete: vi.fn(),
  cancelDelete: vi.fn(),
  noteSort: 'default' as const,
  setNoteSort: vi.fn(),
};

describe('NotesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
    vi.mocked(useNotes).mockReturnValue({ ...mockUseNotes });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('マウント時にfetchNotesが呼ばれる', () => {
    render(<NotesPage />);
    expect(mockUseNotes.fetchNotes).toHaveBeenCalled();
  });

  it('ローディング中はスピナーを表示する', () => {
    vi.mocked(useNotes).mockReturnValue({ ...mockUseNotes, loading: true, notes: [] });
    const { container } = render(<NotesPage />);
    expect(container.querySelector('.animate-spin')).toBeInTheDocument();
  });

  it('ノートが空の場合は「ノートがありません」を表示する', () => {
    render(<NotesPage />);
    expect(screen.getAllByText('ノートがありません').length).toBeGreaterThanOrEqual(1);
  });

  it('ノート未選択時にEmptyStateを表示する', () => {
    render(<NotesPage />);
    expect(screen.getByText('ノートを選択してください')).toBeInTheDocument();
  });

  it('ノート一覧を表示する', () => {
    const noteList = [
      makeNote({ id: 1, title: 'テスト1', content: '内容1' }),
      makeNote({ id: 2, title: 'テスト2', content: '内容2', isPinned: true }),
    ];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
    });
    render(<NotesPage />);
    expect(screen.getAllByText('テスト1').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('テスト2').length).toBeGreaterThanOrEqual(1);
  });

  it('ノート選択時にエディタが表示される', () => {
    const note = makeNote({ id: 1, title: '選択ノート', content: '選択内容' });
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: [note],
      selectedNoteId: 1,
      selectedNote: note,
    });
    render(<NotesPage />);
    expect(screen.getByDisplayValue('選択ノート')).toBeInTheDocument();
    expect(screen.getByTestId('block-editor')).toBeInTheDocument();
  });

  it('新しいノートボタンでcreateNoteが呼ばれる', async () => {
    mockUseNotes.createNote.mockResolvedValue(null);
    render(<NotesPage />);

    const createButtons = screen.getAllByText('新しいノート');
    await act(async () => {
      fireEvent.click(createButtons[0]);
    });

    expect(mockUseNotes.createNote).toHaveBeenCalledWith('無題');
  });

  it('タイトル変更で自動保存が800ms後に呼ばれる', () => {
    const note = makeNote({ id: 1, title: '元タイトル' });
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: [note],
      selectedNoteId: 1,
      selectedNote: note,
    });
    render(<NotesPage />);

    const titleInput = screen.getByDisplayValue('元タイトル');
    fireEvent.change(titleInput, { target: { value: '新タイトル' } });

    expect(mockUseNotes.updateNote).not.toHaveBeenCalled();

    act(() => {
      vi.advanceTimersByTime(800);
    });

    expect(mockUseNotes.updateNote).toHaveBeenCalledWith(1, expect.objectContaining({
      title: '新タイトル',
      isPinned: false,
    }));
  });

  it('タイトル変更でデバウンス後にupdateNoteが呼ばれる', () => {
    const note = makeNote({ id: 1, title: 'タイトル' });
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: [note],
      selectedNoteId: 1,
      selectedNote: note,
    });
    render(<NotesPage />);

    const titleInput = screen.getByDisplayValue('タイトル');
    fireEvent.change(titleInput, { target: { value: 'ABC' } });

    act(() => { vi.advanceTimersByTime(800); });

    expect(mockUseNotes.updateNote).toHaveBeenCalledTimes(1);
    expect(mockUseNotes.updateNote).toHaveBeenCalledWith(1, {
      title: 'ABC',
      content: '',
      isPinned: false,
    });
  });

  it('モバイルでノート一覧ボタンが表示される', () => {
    render(<NotesPage />);
    expect(screen.getByLabelText('ノート一覧を開く')).toBeInTheDocument();
  });

  it('検索入力欄が表示される', () => {
    render(<NotesPage />);
    expect(screen.getAllByPlaceholderText('ノートを検索...').length).toBeGreaterThanOrEqual(1);
  });

  it('検索入力でsetSearchQueryが呼ばれる', () => {
    render(<NotesPage />);
    const searchInputs = screen.getAllByPlaceholderText('ノートを検索...');
    fireEvent.change(searchInputs[0], { target: { value: 'テスト' } });
    expect(mockUseNotes.setSearchQuery).toHaveBeenCalledWith('テスト');
  });

  it('filteredNotesを使ってノート一覧を表示する', () => {
    const allNotes = [
      makeNote({ id: 1, title: 'テスト1', content: '内容1' }),
      makeNote({ id: 2, title: 'テスト2', content: '内容2', isPinned: true }),
    ];
    const filtered = [allNotes[0]];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: allNotes,
      filteredNotes: filtered,
      searchQuery: 'テスト1',
    });
    render(<NotesPage />);
    expect(screen.getAllByText('テスト1').length).toBeGreaterThanOrEqual(1);
  });

  it('ピン留めトグルでtogglePinが呼ばれる', () => {
    const noteList = [makeNote({ id: 1, title: 'ピンテスト', content: '内容' })];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
    });
    render(<NotesPage />);
    const pinButtons = screen.getAllByLabelText('ピン留め');
    fireEvent.click(pinButtons[0]);
    expect(mockUseNotes.togglePin).toHaveBeenCalledWith(1);
  });

  it('削除ボタンクリックでrequestDeleteが呼ばれる', () => {
    const noteList = [makeNote({ id: 1, title: '削除テスト', content: '内容' })];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
    });
    render(<NotesPage />);
    const deleteButtons = screen.getAllByLabelText('ノートを削除');
    fireEvent.click(deleteButtons[0]);
    expect(mockUseNotes.requestDelete).toHaveBeenCalledWith(1);
  });

  it('deleteTargetIdがセットされると確認ダイアログが表示される', () => {
    const noteList = [makeNote({ id: 1, title: '削除テスト', content: '内容' })];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
      deleteTargetId: 1,
    });
    render(<NotesPage />);
    expect(screen.getByText('このノートを削除しますか？')).toBeInTheDocument();
  });

  it('確認ダイアログで削除を実行するとconfirmDeleteが呼ばれる', async () => {
    const noteList = [makeNote({ id: 1, title: '削除テスト', content: '内容' })];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
      deleteTargetId: 1,
    });
    render(<NotesPage />);
    await act(async () => {
      fireEvent.click(screen.getByText('削除'));
    });
    expect(mockUseNotes.confirmDelete).toHaveBeenCalled();
  });

  it('確認ダイアログでキャンセルするとcancelDeleteが呼ばれる', () => {
    const noteList = [makeNote({ id: 1, title: '削除テスト', content: '内容' })];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
      deleteTargetId: 1,
    });
    render(<NotesPage />);
    fireEvent.click(screen.getByText('キャンセル'));
    expect(mockUseNotes.cancelDelete).toHaveBeenCalled();
  });

  it('deleteTargetIdがnullのとき確認ダイアログが非表示', () => {
    const noteList = [makeNote({ id: 1, title: 'テスト', content: '内容' })];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
      deleteTargetId: null,
    });
    render(<NotesPage />);
    expect(screen.queryByText('このノートを削除しますか？')).not.toBeInTheDocument();
  });

  it('ノート作成成功時にトーストが表示される', async () => {
    mockUseNotes.createNote.mockResolvedValue(makeNote({ id: 1, title: '無題' }));
    render(<NotesPage />);

    const createButtons = screen.getAllByText('新しいノート');
    await act(async () => {
      fireEvent.click(createButtons[0]);
    });

    expect(mockShowToast).toHaveBeenCalledWith('success', 'ノートを作成しました');
  });

  it('ノート作成失敗時にエラートーストが表示される', async () => {
    mockUseNotes.createNote.mockResolvedValue(null);
    render(<NotesPage />);

    const createButtons = screen.getAllByText('新しいノート');
    await act(async () => {
      fireEvent.click(createButtons[0]);
    });

    expect(mockShowToast).toHaveBeenCalledWith('error', 'ノートの作成に失敗しました');
  });

  it('ノート削除確認時にトーストが表示される', async () => {
    mockUseNotes.confirmDelete.mockResolvedValue(undefined);
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      deleteTargetId: 1,
    });
    render(<NotesPage />);

    await act(async () => {
      fireEvent.click(screen.getByText('削除'));
    });

    expect(mockShowToast).toHaveBeenCalledWith('success', 'ノートを削除しました');
  });

  it('ノートがある場合、件数が表示される', () => {
    const noteList = [
      makeNote({ id: 1, title: 'テスト1', content: '内容1' }),
      makeNote({ id: 2, title: 'テスト2', content: '内容2' }),
      makeNote({ id: 3, title: 'テスト3', content: '内容3' }),
    ];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
    });
    render(<NotesPage />);
    expect(screen.getAllByText('3件').length).toBeGreaterThanOrEqual(1);
  });

  it('ノートが0件の場合、件数が表示される', () => {
    render(<NotesPage />);
    expect(screen.getAllByText('0件').length).toBeGreaterThanOrEqual(1);
  });

  it('複数ノートがある場合、正しいノートのrequestDeleteが呼ばれる', () => {
    const noteList = [
      makeNote({ id: 1, title: 'ノート1', content: '内容1' }),
      makeNote({ id: 2, title: 'ノート2', content: '内容2' }),
    ];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
    });
    render(<NotesPage />);
    const deleteButtons = screen.getAllByLabelText('ノートを削除');
    fireEvent.click(deleteButtons[1]);
    expect(mockUseNotes.requestDelete).toHaveBeenCalledWith(2);
  });
});
