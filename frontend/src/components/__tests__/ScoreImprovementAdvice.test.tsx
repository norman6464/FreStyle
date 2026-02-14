import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ScoreImprovementAdvice from '../ScoreImprovementAdvice';

describe('ScoreImprovementAdvice', () => {
  const highScores = [
    { axis: '論理的構成力', score: 9, comment: '' },
    { axis: '配慮表現', score: 8, comment: '' },
    { axis: '要約力', score: 8.5, comment: '' },
    { axis: '提案力', score: 9, comment: '' },
    { axis: '質問・傾聴力', score: 8, comment: '' },
  ];

  const mixedScores = [
    { axis: '論理的構成力', score: 5, comment: '' },
    { axis: '配慮表現', score: 8, comment: '' },
    { axis: '要約力', score: 4, comment: '' },
    { axis: '提案力', score: 9, comment: '' },
    { axis: '質問・傾聴力', score: 6, comment: '' },
  ];

  it('タイトルが表示される', () => {
    render(<ScoreImprovementAdvice scores={mixedScores} />);

    expect(screen.getByText('改善アドバイス')).toBeInTheDocument();
  });

  it('低スコアの軸にアドバイスが表示される', () => {
    render(<ScoreImprovementAdvice scores={mixedScores} />);

    expect(screen.getByText(/論理的構成力/)).toBeInTheDocument();
    expect(screen.getByText(/要約力/)).toBeInTheDocument();
  });

  it('高スコアの軸にはアドバイスが表示されない', () => {
    render(<ScoreImprovementAdvice scores={mixedScores} />);

    const adviceItems = screen.getAllByTestId('advice-item');
    const axisTexts = adviceItems.map((item) => item.textContent);
    expect(axisTexts.some((t) => t?.includes('提案力'))).toBe(false);
  });

  it('全軸が高スコアの場合は祝福メッセージが表示される', () => {
    render(<ScoreImprovementAdvice scores={highScores} />);

    expect(screen.getByText('素晴らしい成績です！この調子で続けましょう。')).toBeInTheDocument();
  });

  it('スコア7未満の軸のみアドバイス対象になる', () => {
    render(<ScoreImprovementAdvice scores={mixedScores} />);

    const adviceItems = screen.getAllByTestId('advice-item');
    expect(adviceItems).toHaveLength(3);
  });
});
