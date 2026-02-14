import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import SkillMilestoneCard from '../SkillMilestoneCard';

const mockScores = [
  { axis: '論理的構成力', score: 8.5, comment: '' },
  { axis: '配慮表現', score: 6.2, comment: '' },
  { axis: '要約力', score: 4.0, comment: '' },
  { axis: '提案力', score: 9.1, comment: '' },
  { axis: '質問・傾聴力', score: 2.5, comment: '' },
];

describe('SkillMilestoneCard', () => {
  it('タイトルが表示される', () => {
    render(<SkillMilestoneCard scores={mockScores} />);

    expect(screen.getByText('スキル到達レベル')).toBeInTheDocument();
  });

  it('各スキル軸が表示される', () => {
    render(<SkillMilestoneCard scores={mockScores} />);

    expect(screen.getByText('論理的構成力')).toBeInTheDocument();
    expect(screen.getByText('配慮表現')).toBeInTheDocument();
    expect(screen.getByText('要約力')).toBeInTheDocument();
    expect(screen.getByText('提案力')).toBeInTheDocument();
    expect(screen.getByText('質問・傾聴力')).toBeInTheDocument();
  });

  it('スコアに応じたレベルラベルが表示される', () => {
    render(<SkillMilestoneCard scores={mockScores} />);

    expect(screen.getByText('上級')).toBeInTheDocument();
    expect(screen.getByText('エキスパート')).toBeInTheDocument();
  });

  it('次のマイルストーン情報が表示される', () => {
    render(<SkillMilestoneCard scores={mockScores} />);

    // 配慮表現 6.2 → 上級(7.0)まであと0.8
    expect(screen.getByText('あと 0.8')).toBeInTheDocument();
  });

  it('スコアがない場合は何も表示しない', () => {
    const { container } = render(<SkillMilestoneCard scores={[]} />);

    expect(container.firstChild).toBeNull();
  });
});
