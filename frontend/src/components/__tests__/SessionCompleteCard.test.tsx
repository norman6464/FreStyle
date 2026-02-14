import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import SessionCompleteCard from '../SessionCompleteCard';

describe('SessionCompleteCard', () => {
  const defaultProps = {
    duration: 300,
    messageCount: 12,
    onViewScores: vi.fn(),
    onPracticeAgain: vi.fn(),
  };

  it('完了メッセージが表示される', () => {
    render(<SessionCompleteCard {...defaultProps} />);
    expect(screen.getByText('セッション完了')).toBeInTheDocument();
  });

  it('経過時間がフォーマットされて表示される', () => {
    render(<SessionCompleteCard {...defaultProps} />);
    expect(screen.getByText('5:00')).toBeInTheDocument();
  });

  it('メッセージ数が表示される', () => {
    render(<SessionCompleteCard {...defaultProps} />);
    expect(screen.getByText('12')).toBeInTheDocument();
  });

  it('スコア確認ボタンクリックでonViewScoresが呼ばれる', () => {
    const onViewScores = vi.fn();
    render(<SessionCompleteCard {...defaultProps} onViewScores={onViewScores} />);
    fireEvent.click(screen.getByText('スコアを確認'));
    expect(onViewScores).toHaveBeenCalled();
  });

  it('もう一度練習ボタンクリックでonPracticeAgainが呼ばれる', () => {
    const onPracticeAgain = vi.fn();
    render(<SessionCompleteCard {...defaultProps} onPracticeAgain={onPracticeAgain} />);
    fireEvent.click(screen.getByText('もう一度練習'));
    expect(onPracticeAgain).toHaveBeenCalled();
  });

  it('ラベルが表示される', () => {
    render(<SessionCompleteCard {...defaultProps} />);
    expect(screen.getByText('経過時間')).toBeInTheDocument();
    expect(screen.getByText('メッセージ数')).toBeInTheDocument();
  });

  it('1時間以上の場合も正しくフォーマットされる', () => {
    render(<SessionCompleteCard {...defaultProps} duration={3661} />);
    expect(screen.getByText('61:01')).toBeInTheDocument();
  });
});
