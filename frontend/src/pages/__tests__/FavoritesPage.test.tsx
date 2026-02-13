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

    expect(screen.getByText('フォーマル')).toBeInTheDocument();
    expect(screen.getByText('ソフト')).toBeInTheDocument();
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
