import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import FavoritesPage from '../FavoritesPage';

const mockRemoveFavorite = vi.fn();
const mockSetSearchQuery = vi.fn();
const mockSetPatternFilter = vi.fn();

const allPhrases = [
  {
    id: '1',
    originalText: '確認お願いします',
    rephrasedText: 'ご確認いただけますでしょうか',
    pattern: 'フォーマル',
    createdAt: '2026-01-15T10:00:00.000Z',
  },
  {
    id: '2',
    originalText: 'これでいいですか',
    rephrasedText: 'こちらの内容でよろしいでしょうか',
    pattern: 'ソフト',
    createdAt: '2026-01-14T10:00:00.000Z',
  },
];

let mockReturnValue: any = {
  phrases: allPhrases,
  filteredPhrases: allPhrases,
  searchQuery: '',
  setSearchQuery: mockSetSearchQuery,
  patternFilter: 'すべて',
  setPatternFilter: mockSetPatternFilter,
  removeFavorite: mockRemoveFavorite,
  saveFavorite: vi.fn(),
  isFavorite: vi.fn(),
};

vi.mock('../../hooks/useFavoritePhrase', () => ({
  useFavoritePhrase: () => mockReturnValue,
}));

describe('FavoritesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockReturnValue = {
      phrases: allPhrases,
      filteredPhrases: allPhrases,
      searchQuery: '',
      setSearchQuery: mockSetSearchQuery,
      patternFilter: 'すべて',
      setPatternFilter: mockSetPatternFilter,
      removeFavorite: mockRemoveFavorite,
      saveFavorite: vi.fn(),
      isFavorite: vi.fn(),
    };
  });

  it('お気に入りフレーズ一覧が表示される', () => {
    render(<FavoritesPage />);

    expect(screen.getByRole('heading', { name: /お気に入りフレーズ/ })).toBeInTheDocument();
    expect(screen.getByText('ご確認いただけますでしょうか')).toBeInTheDocument();
    expect(screen.getByText('こちらの内容でよろしいでしょうか')).toBeInTheDocument();
  });

  it('パターンラベルが表示される', () => {
    render(<FavoritesPage />);

    expect(screen.getAllByText('フォーマル').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('ソフト').length).toBeGreaterThanOrEqual(1);
  });

  it('元のテキストが表示される', () => {
    render(<FavoritesPage />);

    expect(screen.getByText(/確認お願いします/)).toBeInTheDocument();
    expect(screen.getByText(/これでいいですか/)).toBeInTheDocument();
  });

  it('削除ボタンクリックで確認ダイアログが表示される', () => {
    render(<FavoritesPage />);

    const deleteButtons = screen.getAllByLabelText('お気に入りから削除');
    fireEvent.click(deleteButtons[0]);

    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText(/「ご確認いただけますでしょうか」を削除しますか/)).toBeInTheDocument();
  });

  it('確認ダイアログで削除を実行するとremoveFavoriteが呼ばれる', () => {
    render(<FavoritesPage />);

    const deleteButtons = screen.getAllByLabelText('お気に入りから削除');
    fireEvent.click(deleteButtons[0]);
    fireEvent.click(screen.getByText('削除'));

    expect(mockRemoveFavorite).toHaveBeenCalledWith('1');
  });

  it('確認ダイアログでキャンセルするとremoveFavoriteが呼ばれない', () => {
    render(<FavoritesPage />);

    const deleteButtons = screen.getAllByLabelText('お気に入りから削除');
    fireEvent.click(deleteButtons[0]);
    fireEvent.click(screen.getByText('キャンセル'));

    expect(mockRemoveFavorite).not.toHaveBeenCalled();
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('検索ボックスが表示される', () => {
    render(<FavoritesPage />);

    expect(screen.getByPlaceholderText('フレーズを検索...')).toBeInTheDocument();
  });

  it('検索でsetSearchQueryが呼ばれる', () => {
    render(<FavoritesPage />);

    fireEvent.change(screen.getByPlaceholderText('フレーズを検索...'), {
      target: { value: '確認' },
    });

    expect(mockSetSearchQuery).toHaveBeenCalledWith('確認');
  });

  it('パターンフィルタボタンでsetPatternFilterが呼ばれる', () => {
    render(<FavoritesPage />);

    const filterButtons = screen.getAllByRole('button');
    const softButton = filterButtons.find(b => b.textContent === 'ソフト');
    fireEvent.click(softButton!);

    expect(mockSetPatternFilter).toHaveBeenCalledWith('ソフト');
  });

  it('filteredPhrasesが空の場合にメッセージを表示する', () => {
    mockReturnValue = { ...mockReturnValue, filteredPhrases: [] };

    render(<FavoritesPage />);

    expect(screen.getByText('該当するフレーズがありません')).toBeInTheDocument();
  });

  it('フレーズが空のときメッセージが表示される', () => {
    mockReturnValue = { ...mockReturnValue, phrases: [], filteredPhrases: [] };

    render(<FavoritesPage />);

    expect(screen.getByText('お気に入りフレーズがありません')).toBeInTheDocument();
  });

  it('フレーズ件数が表示される', () => {
    render(<FavoritesPage />);

    expect(screen.getByText('2件')).toBeInTheDocument();
  });

  it('フレーズが0件の場合は件数が表示されない', () => {
    mockReturnValue = { ...mockReturnValue, phrases: [], filteredPhrases: [] };

    render(<FavoritesPage />);

    expect(screen.queryByText('0件')).not.toBeInTheDocument();
  });
});
