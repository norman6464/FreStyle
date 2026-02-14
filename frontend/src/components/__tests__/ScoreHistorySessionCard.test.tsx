import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import ScoreHistorySessionCard from '../ScoreHistorySessionCard';
import type { ScoreHistoryItem } from '../../hooks/useScoreHistory';

const baseItem: ScoreHistoryItem = {
  sessionId: 1,
  sessionTitle: '障害報告の練習',
  overallScore: 7.5,
  scores: [
    { axis: '論理的構成力', score: 8, comment: '' },
    { axis: '配慮表現', score: 7, comment: '' },
    { axis: '要約力', score: 7, comment: '' },
    { axis: '提案力', score: 8, comment: '' },
    { axis: '質問・傾聴力', score: 7, comment: '' },
  ],
  createdAt: '2026-02-10T10:00:00Z',
};

describe('ScoreHistorySessionCard', () => {
  it('セッションタイトルが表示される', () => {
    render(<ScoreHistorySessionCard item={baseItem} delta={null} onClick={vi.fn()} />);
    expect(screen.getByText('障害報告の練習')).toBeInTheDocument();
  });

  it('総合スコアが表示される', () => {
    render(<ScoreHistorySessionCard item={baseItem} delta={null} onClick={vi.fn()} />);
    expect(screen.getByText('7.5')).toBeInTheDocument();
  });

  it('日付がフォーマットされて表示される', () => {
    render(<ScoreHistorySessionCard item={baseItem} delta={null} onClick={vi.fn()} />);
    expect(screen.getByText(/2026年/)).toBeInTheDocument();
  });

  it('各スキル軸が表示される', () => {
    render(<ScoreHistorySessionCard item={baseItem} delta={null} onClick={vi.fn()} />);
    expect(screen.getByText('論理的構成力')).toBeInTheDocument();
    expect(screen.getByText('配慮表現')).toBeInTheDocument();
  });

  it('プラスのdeltaが緑色で表示される', () => {
    render(<ScoreHistorySessionCard item={baseItem} delta={1.2} onClick={vi.fn()} />);
    expect(screen.getByText('+1.2')).toBeInTheDocument();
  });

  it('マイナスのdeltaが赤色で表示される', () => {
    render(<ScoreHistorySessionCard item={baseItem} delta={-0.5} onClick={vi.fn()} />);
    const el = screen.getByText(/0\.5/);
    expect(el.className).toContain('text-rose-400');
  });

  it('deltaがnullの場合は変動表示がない', () => {
    const { container } = render(<ScoreHistorySessionCard item={baseItem} delta={null} onClick={vi.fn()} />);
    expect(container.querySelector('.text-emerald-400')).toBeNull();
    expect(container.querySelector('.text-rose-400')).toBeNull();
  });

  it('クリックでonClickが呼ばれる', () => {
    const handleClick = vi.fn();
    render(<ScoreHistorySessionCard item={baseItem} delta={null} onClick={handleClick} />);
    fireEvent.click(screen.getByText('障害報告の練習').closest('div')!.parentElement!.parentElement!);
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('タイトルがない場合はセッションIDで表示される', () => {
    const noTitle = { ...baseItem, sessionTitle: '' };
    render(<ScoreHistorySessionCard item={noTitle} delta={null} onClick={vi.fn()} />);
    expect(screen.getByText('セッション #1')).toBeInTheDocument();
  });
});
