import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import SessionFeedbackSummary from '../SessionFeedbackSummary';

const mockScores = [
  { axis: '論理的構成力', score: 8.5, comment: '良い' },
  { axis: '配慮表現', score: 7.0, comment: '普通' },
  { axis: '要約力', score: 9.0, comment: '優秀' },
  { axis: '提案力', score: 6.0, comment: '改善点あり' },
  { axis: '質問・傾聴力', score: 5.5, comment: '要練習' },
];

describe('SessionFeedbackSummary', () => {
  it('全スキル軸名が表示される', () => {
    render(<SessionFeedbackSummary scores={mockScores} overallScore={7.2} />);
    expect(screen.getByText('論理的構成力')).toBeInTheDocument();
    expect(screen.getByText('配慮表現')).toBeInTheDocument();
    expect(screen.getByText('要約力')).toBeInTheDocument();
    expect(screen.getByText('提案力')).toBeInTheDocument();
    expect(screen.getByText('質問・傾聴力')).toBeInTheDocument();
  });

  it('各スキル軸のスコアが表示される', () => {
    render(<SessionFeedbackSummary scores={mockScores} overallScore={7.2} />);
    expect(screen.getByText('8.5')).toBeInTheDocument();
    expect(screen.getByText('7.0')).toBeInTheDocument();
    expect(screen.getByText('9.0')).toBeInTheDocument();
    expect(screen.getByText('6.0')).toBeInTheDocument();
    expect(screen.getByText('5.5')).toBeInTheDocument();
  });

  it('総合スコアが高い場合に秀の判定が表示される', () => {
    render(<SessionFeedbackSummary scores={mockScores} overallScore={9.5} />);
    expect(screen.getByText('秀')).toBeInTheDocument();
  });

  it('総合スコアが中程度の場合に良の判定が表示される', () => {
    render(<SessionFeedbackSummary scores={mockScores} overallScore={7.0} />);
    expect(screen.getByText('良')).toBeInTheDocument();
  });

  it('スコアが空の場合は表示されない', () => {
    const { container } = render(<SessionFeedbackSummary scores={[]} overallScore={0} />);
    expect(container.firstChild).toBeNull();
  });

  it('見出しが表示される', () => {
    render(<SessionFeedbackSummary scores={mockScores} overallScore={7.2} />);
    expect(screen.getByText('スキル別スコア')).toBeInTheDocument();
  });

  it('プログレスバーが表示される', () => {
    const { container } = render(<SessionFeedbackSummary scores={mockScores} overallScore={7.0} />);
    const bars = container.querySelectorAll('[role="progressbar"]');
    expect(bars).toHaveLength(5);
  });
});
