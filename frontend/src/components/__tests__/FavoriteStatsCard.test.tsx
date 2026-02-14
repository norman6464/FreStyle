import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import FavoriteStatsCard from '../FavoriteStatsCard';

const phrases = [
  { id: '1', originalText: 'a', rephrasedText: 'b', pattern: 'フォーマル', createdAt: '' },
  { id: '2', originalText: 'c', rephrasedText: 'd', pattern: 'フォーマル', createdAt: '' },
  { id: '3', originalText: 'e', rephrasedText: 'f', pattern: 'ソフト', createdAt: '' },
  { id: '4', originalText: 'g', rephrasedText: 'h', pattern: '簡潔', createdAt: '' },
  { id: '5', originalText: 'i', rephrasedText: 'j', pattern: '質問型', createdAt: '' },
];

describe('FavoriteStatsCard', () => {
  it('タイトルが表示される', () => {
    render(<FavoriteStatsCard phrases={phrases} />);
    expect(screen.getByText('お気に入り統計')).toBeInTheDocument();
  });

  it('総数が表示される', () => {
    render(<FavoriteStatsCard phrases={phrases} />);
    expect(screen.getByText('5')).toBeInTheDocument();
    expect(screen.getByText('フレーズ')).toBeInTheDocument();
  });

  it('パターン別の件数が表示される', () => {
    render(<FavoriteStatsCard phrases={phrases} />);
    const counts = screen.getAllByTestId('pattern-count');
    // フォーマル: 2, ソフト: 1, 簡潔: 1, 質問型: 1
    expect(counts.length).toBeGreaterThanOrEqual(4);
  });

  it('パターンラベルが表示される', () => {
    render(<FavoriteStatsCard phrases={phrases} />);
    expect(screen.getByText('フォーマル')).toBeInTheDocument();
    expect(screen.getByText('ソフト')).toBeInTheDocument();
    expect(screen.getByText('簡潔')).toBeInTheDocument();
    expect(screen.getByText('質問型')).toBeInTheDocument();
  });

  it('フレーズが空の場合は何も表示しない', () => {
    const { container } = render(<FavoriteStatsCard phrases={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('1つのパターンのみの場合も正しく表示される', () => {
    const singlePattern = [
      { id: '1', originalText: 'a', rephrasedText: 'b', pattern: '提案型', createdAt: '' },
    ];
    render(<FavoriteStatsCard phrases={singlePattern} />);
    expect(screen.getAllByText('1').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('提案型')).toBeInTheDocument();
  });
});
