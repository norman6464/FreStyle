import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import FavoritesPage from '../FavoritesPage';

const mockRemoveFavorite = vi.fn();

vi.mock('../../hooks/useFavoritePhrase', () => ({
  useFavoritePhrase: () => ({
    phrases: mockPhrases,
    removeFavorite: mockRemoveFavorite,
    saveFavorite: vi.fn(),
    isFavorite: vi.fn(),
  }),
}));

let mockPhrases = [
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

describe('FavoritesPage', () => {
  beforeEach(() => {
    mockRemoveFavorite.mockClear();
  });

  it('お気に入りフレーズ一覧が表示される', () => {
    render(<FavoritesPage />);

    expect(screen.getByText('お気に入りフレーズ')).toBeInTheDocument();
    expect(screen.getByText('ご確認いただけますでしょうか')).toBeInTheDocument();
    expect(screen.getByText('こちらの内容でよろしいでしょうか')).toBeInTheDocument();
  });

  it('パターンラベルが表示される', () => {
    render(<FavoritesPage />);

    // フィルタボタンとカード内のラベルの両方に存在するため、getAllで確認
    expect(screen.getAllByText('フォーマル').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('ソフト').length).toBeGreaterThanOrEqual(1);
  });

  it('元のテキストが表示される', () => {
    render(<FavoritesPage />);

    expect(screen.getByText(/確認お願いします/)).toBeInTheDocument();
    expect(screen.getByText(/これでいいですか/)).toBeInTheDocument();
  });

  it('削除ボタンでremoveFavoriteが呼ばれる', () => {
    render(<FavoritesPage />);

    const deleteButtons = screen.getAllByLabelText('お気に入りから削除');
    fireEvent.click(deleteButtons[0]);

    expect(mockRemoveFavorite).toHaveBeenCalledWith('1');
  });

  it('検索ボックスが表示される', () => {
    render(<FavoritesPage />);

    expect(screen.getByPlaceholderText('フレーズを検索...')).toBeInTheDocument();
  });

  it('検索でフレーズがフィルタリングされる', () => {
    render(<FavoritesPage />);

    fireEvent.change(screen.getByPlaceholderText('フレーズを検索...'), {
      target: { value: '確認' },
    });

    expect(screen.getByText('ご確認いただけますでしょうか')).toBeInTheDocument();
    expect(screen.queryByText('こちらの内容でよろしいでしょうか')).not.toBeInTheDocument();
  });

  it('パターンフィルタでフレーズが絞り込まれる', () => {
    render(<FavoritesPage />);

    // フィルタボタンは「すべて」「フォーマル」「ソフト」「簡潔」
    const filterButtons = screen.getAllByRole('button');
    const softButton = filterButtons.find(b => b.textContent === 'ソフト');
    fireEvent.click(softButton!);

    expect(screen.getByText('こちらの内容でよろしいでしょうか')).toBeInTheDocument();
    expect(screen.queryByText('ご確認いただけますでしょうか')).not.toBeInTheDocument();
  });

  it('検索結果がない場合にメッセージを表示する', () => {
    render(<FavoritesPage />);

    fireEvent.change(screen.getByPlaceholderText('フレーズを検索...'), {
      target: { value: '存在しないフレーズ' },
    });

    expect(screen.getByText('該当するフレーズがありません')).toBeInTheDocument();
  });

  it('フレーズが空のときメッセージが表示される', () => {
    mockPhrases = [];

    render(<FavoritesPage />);

    expect(screen.getByText('お気に入りフレーズがありません')).toBeInTheDocument();

    // 元に戻す
    mockPhrases = [
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
  });
});
