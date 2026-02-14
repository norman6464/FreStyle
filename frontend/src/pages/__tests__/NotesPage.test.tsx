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
    expect(screen.getByDisplayValue('選択内容')).toBeInTheDocument();
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
        { noteId: 'n1', userId: 1, title: '元タイトル', content: '元内容', isPinned: false, createdAt: 1000, updatedAt: 2000 },
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

    expect(mockUseNotes.updateNote).toHaveBeenCalledWith('n1', {
      title: '新タイトル',
      content: '元内容',
      isPinned: false,
    });
  });

  it('内容変更で自動保存が800ms後に呼ばれる', async () => {
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: [
        { noteId: 'n1', userId: 1, title: 'タイトル', content: '元内容', isPinned: false, createdAt: 1000, updatedAt: 2000 },
      ],
      selectedNoteId: 'n1',
    });
    render(<NotesPage />);

    const textarea = screen.getByDisplayValue('元内容');
    fireEvent.change(textarea, { target: { value: '新内容' } });

    act(() => {
      vi.advanceTimersByTime(800);
    });

    expect(mockUseNotes.updateNote).toHaveBeenCalledWith('n1', {
      title: 'タイトル',
      content: '新内容',
      isPinned: false,
    });
  });

  it('連続入力で自動保存がデバウンスされる', async () => {
    vi.mocked(useNotes).mockReturnValue({
      ...mockUseNotes,
      notes: [
        { noteId: 'n1', userId: 1, title: 'タイトル', content: '', isPinned: false, createdAt: 1000, updatedAt: 2000 },
      ],
      selectedNoteId: 'n1',
    });
    render(<NotesPage />);

    const titleInput = screen.getByDisplayValue('タイトル');
    fireEvent.change(titleInput, { target: { value: 'A' } });
    act(() => { vi.advanceTimersByTime(400); });

    fireEvent.change(titleInput, { target: { value: 'AB' } });
    act(() => { vi.advanceTimersByTime(400); });

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
});
