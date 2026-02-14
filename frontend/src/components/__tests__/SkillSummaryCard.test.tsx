import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import SkillSummaryCard from '../SkillSummaryCard';

const mockScores = [
  { axis: '論理的構成力', score: 9, comment: '素晴らしい' },
  { axis: '配慮表現', score: 8, comment: '良い' },
  { axis: '要約力', score: 5, comment: '改善が必要' },
  { axis: '提案力', score: 7, comment: '普通' },
  { axis: '質問・傾聴力', score: 4, comment: '要改善' },
];

describe('SkillSummaryCard', () => {
  it('タイトルが表示される', () => {
    render(<SkillSummaryCard scores={mockScores} />);
    expect(screen.getByText('スキル強弱サマリー')).toBeInTheDocument();
  });

  it('強みの上位2軸が表示される', () => {
    render(<SkillSummaryCard scores={mockScores} />);
    const strengthSection = screen.getByTestId('strengths');
    expect(strengthSection).toHaveTextContent('論理的構成力');
    expect(strengthSection).toHaveTextContent('配慮表現');
  });

  it('弱みの下位2軸が表示される', () => {
    render(<SkillSummaryCard scores={mockScores} />);
    const weaknessSection = screen.getByTestId('weaknesses');
    expect(weaknessSection).toHaveTextContent('質問・傾聴力');
    expect(weaknessSection).toHaveTextContent('要約力');
  });

  it('スコアが空の場合はnullを返す', () => {
    const { container } = render(<SkillSummaryCard scores={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('各軸のスコアが表示される', () => {
    render(<SkillSummaryCard scores={mockScores} />);
    expect(screen.getByText('9')).toBeInTheDocument();
    expect(screen.getByText('4')).toBeInTheDocument();
  });
});
