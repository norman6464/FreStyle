import { render, screen, fireEvent, act, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import NotesPage from '../NotesPage';
import { useNotes } from '../../hooks/useNotes';

vi.mock('../../hooks/useNotes');

const mockUseNotes = {
  notes: [],
  filteredNotes: [],
  selectedNoteId: null,
  loading: false,
  searchQuery: '',
  setSearchQuery: vi.fn(),
  fetchNotes: vi.fn(),
  createNote: vi.fn(),
  updateNote: vi.fn(),
  deleteNote: vi.fn(),
  selectNote: vi.fn(),
  togglePin: vi.fn(),
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
      { noteId: 'n1', userId: 1, title: 'テスト1', content: '内容1', isPinned: false, createdAt: 1000, updatedAt: 2000 },
      { noteId: 'n2', userId: 1, title: 'テスト2', content: '内容2', isPinned: true, createdAt: 1500, updatedAt: 3000 },
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
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: [
        { noteId: 'n1', userId: 1, title: '選択ノート', content: '選択内容', isPinned: false, createdAt: 1000, updatedAt: 2000 },
      ],
      selectedNoteId: 'n1',
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

  it('タイトル変更で自動保存が800ms後に呼ばれる', async () => {
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: [
        { noteId: 'n1', userId: 1, title: '元タイトル', content: '', isPinned: false, createdAt: 1000, updatedAt: 2000 },
      ],
      selectedNoteId: 'n1',
    });
    render(<NotesPage />);

    const titleInput = screen.getByDisplayValue('元タイトル');
    fireEvent.change(titleInput, { target: { value: '新タイトル' } });

    expect(mockUseNotes.updateNote).not.toHaveBeenCalled();

    act(() => {
      vi.advanceTimersByTime(800);
    });

    expect(mockUseNotes.updateNote).toHaveBeenCalledWith('n1', expect.objectContaining({
      title: '新タイトル',
      isPinned: false,
    }));
  });

  it('タイトル変更でデバウンス後にupdateNoteが呼ばれる', async () => {
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: [
        { noteId: 'n1', userId: 1, title: 'タイトル', content: '', isPinned: false, createdAt: 1000, updatedAt: 2000 },
      ],
      selectedNoteId: 'n1',
    });
    render(<NotesPage />);

    const titleInput = screen.getByDisplayValue('タイトル');
    fireEvent.change(titleInput, { target: { value: 'ABC' } });

    act(() => { vi.advanceTimersByTime(800); });

    expect(mockUseNotes.updateNote).toHaveBeenCalledTimes(1);
    expect(mockUseNotes.updateNote).toHaveBeenCalledWith('n1', {
      title: 'ABC',
      content: '',
      isPinned: false,
    });
  });

  it('モバイルでノート一覧ボタンが表示される', () => {
    render(<NotesPage />);
    expect(screen.getByLabelText('ノート一覧を開く')).toBeInTheDocument();
  });

  // 検索・ピン留め機能テスト

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
      { noteId: 'n1', userId: 1, title: 'テスト1', content: '内容1', isPinned: false, createdAt: 1000, updatedAt: 2000 },
      { noteId: 'n2', userId: 1, title: 'テスト2', content: '内容2', isPinned: true, createdAt: 1500, updatedAt: 3000 },
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

  it('ピン留めトグルでtogglePinが呼ばれる', async () => {
    const noteList = [
      { noteId: 'n1', userId: 1, title: 'ピンテスト', content: '内容', isPinned: false, createdAt: 1000, updatedAt: 2000 },
    ];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
    });
    render(<NotesPage />);
    const pinButtons = screen.getAllByLabelText('ピン留め');
    fireEvent.click(pinButtons[0]);
    expect(mockUseNotes.togglePin).toHaveBeenCalledWith('n1');
  });

  it('ノート削除時に確認ダイアログが表示される', () => {
    const noteList = [
      { noteId: 'n1', userId: 1, title: '削除テスト', content: '内容', isPinned: false, createdAt: 1000, updatedAt: 2000 },
    ];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
    });
    render(<NotesPage />);
    const deleteButtons = screen.getAllByLabelText('ノートを削除');
    fireEvent.click(deleteButtons[0]);
    expect(screen.getByText('このノートを削除しますか？')).toBeInTheDocument();
    expect(mockUseNotes.deleteNote).not.toHaveBeenCalled();
  });

  it('確認ダイアログで削除を実行するとdeleteNoteが呼ばれる', async () => {
    const noteList = [
      { noteId: 'n1', userId: 1, title: '削除テスト', content: '内容', isPinned: false, createdAt: 1000, updatedAt: 2000 },
    ];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
    });
    render(<NotesPage />);
    const deleteButtons = screen.getAllByLabelText('ノートを削除');
    fireEvent.click(deleteButtons[0]);
    await act(async () => {
      fireEvent.click(screen.getByText('削除'));
    });
    expect(mockUseNotes.deleteNote).toHaveBeenCalledWith('n1');
  });

  it('確認ダイアログでキャンセルするとdeleteNoteが呼ばれない', () => {
    const noteList = [
      { noteId: 'n1', userId: 1, title: '削除テスト', content: '内容', isPinned: false, createdAt: 1000, updatedAt: 2000 },
    ];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
    });
    render(<NotesPage />);
    const deleteButtons = screen.getAllByLabelText('ノートを削除');
    fireEvent.click(deleteButtons[0]);
    fireEvent.click(screen.getByText('キャンセル'));
    expect(mockUseNotes.deleteNote).not.toHaveBeenCalled();
  });

  it('削除確認ダイアログのキャンセル後にダイアログが閉じる', () => {
    const noteList = [
      { noteId: 'n1', userId: 1, title: '削除テスト', content: '内容', isPinned: false, createdAt: 1000, updatedAt: 2000 },
    ];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
    });
    render(<NotesPage />);
    const deleteButtons = screen.getAllByLabelText('ノートを削除');
    fireEvent.click(deleteButtons[0]);
    expect(screen.getByText('このノートを削除しますか？')).toBeInTheDocument();
    fireEvent.click(screen.getByText('キャンセル'));
    expect(screen.queryByText('このノートを削除しますか？')).not.toBeInTheDocument();
  });

  it('複数ノートがある場合、正しいノートが削除される', async () => {
    const noteList = [
      { noteId: 'n1', userId: 1, title: 'ノート1', content: '内容1', isPinned: false, createdAt: 1000, updatedAt: 2000 },
      { noteId: 'n2', userId: 1, title: 'ノート2', content: '内容2', isPinned: false, createdAt: 1500, updatedAt: 3000 },
    ];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
    });
    render(<NotesPage />);
    const deleteButtons = screen.getAllByLabelText('ノートを削除');
    fireEvent.click(deleteButtons[1]);
    await act(async () => {
      fireEvent.click(screen.getByText('削除'));
    });
    expect(mockUseNotes.deleteNote).toHaveBeenCalledWith('n2');
  });

  it('削除ボタンを押してもdeleteNoteが直接呼ばれない（確認が必要）', () => {
    const noteList = [
      { noteId: 'n1', userId: 1, title: 'テスト', content: '内容', isPinned: false, createdAt: 1000, updatedAt: 2000 },
    ];
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: noteList,
      filteredNotes: noteList,
    });
    render(<NotesPage />);
    const deleteButtons = screen.getAllByLabelText('ノートを削除');
    fireEvent.click(deleteButtons[0]);
    expect(mockUseNotes.deleteNote).not.toHaveBeenCalled();
    expect(screen.getByText('このノートを削除しますか？')).toBeInTheDocument();
  });
});
