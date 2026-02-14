import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import SkillGapAnalysisCard from '../SkillGapAnalysisCard';

describe('SkillGapAnalysisCard', () => {
  const defaultScores = [
    { axis: '論理的構成力', score: 6.0 },
    { axis: '配慮表現', score: 8.0 },
    { axis: '要約力', score: 5.0 },
    { axis: '提案力', score: 7.0 },
    { axis: '質問・傾聴力', score: 9.0 },
  ];

  it('タイトルが表示される', () => {
    render(<SkillGapAnalysisCard scores={defaultScores} goal={8.0} />);
    expect(screen.getByText('スキルギャップ分析')).toBeInTheDocument();
  });

  it('各軸名が表示される', () => {
    render(<SkillGapAnalysisCard scores={defaultScores} goal={8.0} />);
    expect(screen.getByText('論理的構成力')).toBeInTheDocument();
    expect(screen.getByText('配慮表現')).toBeInTheDocument();
    expect(screen.getByText('要約力')).toBeInTheDocument();
    expect(screen.getByText('提案力')).toBeInTheDocument();
    expect(screen.getByText('質問・傾聴力')).toBeInTheDocument();
  });

  it('ギャップが大きい順にソートされる', () => {
    render(<SkillGapAnalysisCard scores={defaultScores} goal={8.0} />);
    const axisLabels = screen.getAllByTestId('gap-axis');
    // 要約力(gap=3.0) > 論理的構成力(gap=2.0) > 提案力(gap=1.0) > 配慮表現(gap=0) > 質問(gap=0, over)
    expect(axisLabels[0]).toHaveTextContent('要約力');
    expect(axisLabels[1]).toHaveTextContent('論理的構成力');
    expect(axisLabels[2]).toHaveTextContent('提案力');
  });

  it('ギャップ値が表示される', () => {
    render(<SkillGapAnalysisCard scores={defaultScores} goal={8.0} />);
    const gapValues = screen.getAllByTestId('gap-value');
    expect(gapValues[0]).toHaveTextContent('-3.0');
    expect(gapValues[1]).toHaveTextContent('-2.0');
    expect(gapValues[2]).toHaveTextContent('-1.0');
  });

  it('目標達成済みの軸には達成表示がある', () => {
    render(<SkillGapAnalysisCard scores={defaultScores} goal={8.0} />);
    const achievedItems = screen.getAllByTestId('gap-achieved');
    expect(achievedItems.length).toBeGreaterThanOrEqual(2); // 配慮表現(8.0), 質問・傾聴力(9.0)
  });

  it('全軸達成時に祝福メッセージが表示される', () => {
    const highScores = [
      { axis: '論理的構成力', score: 9.0 },
      { axis: '配慮表現', score: 8.5 },
      { axis: '要約力', score: 8.0 },
      { axis: '提案力', score: 9.5 },
      { axis: '質問・傾聴力', score: 8.0 },
    ];
    render(<SkillGapAnalysisCard scores={highScores} goal={8.0} />);
    expect(screen.getByText(/全スキル目標達成/)).toBeInTheDocument();
  });

  it('空のスコア配列の場合は何も表示しない', () => {
    const { container } = render(<SkillGapAnalysisCard scores={[]} goal={8.0} />);
    expect(container.firstChild).toBeNull();
  });
});
