import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import SkillRadarChart from '../SkillRadarChart';
import type { AxisScore } from '../../types';

const mockScores: AxisScore[] = [
  { axis: '論理的構成力', score: 8, comment: '' },
  { axis: '配慮表現', score: 6, comment: '' },
  { axis: '要約力', score: 7, comment: '' },
  { axis: '提案力', score: 5, comment: '' },
  { axis: '質問・傾聴力', score: 9, comment: '' },
];

describe('SkillRadarChart', () => {
  it('レーダーチャートが描画される', () => {
    const { container } = render(<SkillRadarChart scores={mockScores} />);

    expect(container.querySelector('svg')).toBeInTheDocument();
  });

  it('5軸のラベルが表示される', () => {
    render(<SkillRadarChart scores={mockScores} />);

    expect(screen.getByText('論理的構成力')).toBeInTheDocument();
    expect(screen.getByText('配慮表現')).toBeInTheDocument();
    expect(screen.getByText('要約力')).toBeInTheDocument();
    expect(screen.getByText('提案力')).toBeInTheDocument();
    expect(screen.getByText('質問・傾聴力')).toBeInTheDocument();
  });

  it('スコアのポリゴンが描画される', () => {
    const { container } = render(<SkillRadarChart scores={mockScores} />);

    const polygon = container.querySelector('polygon');
    expect(polygon).toBeInTheDocument();
  });

  it('グリッド線が描画される', () => {
    const { container } = render(<SkillRadarChart scores={mockScores} />);

    // 5段階のグリッド（2, 4, 6, 8, 10）
    const circles = container.querySelectorAll('polygon[class*="grid"]');
    expect(circles.length).toBeGreaterThanOrEqual(1);
  });

  it('タイトルが表示される', () => {
    render(<SkillRadarChart scores={mockScores} title="スキルバランス" />);

    expect(screen.getByText('スキルバランス')).toBeInTheDocument();
  });

  it('空のスコアでもクラッシュしない', () => {
    const { container } = render(<SkillRadarChart scores={[]} />);

    expect(container.querySelector('svg')).toBeInTheDocument();
  });
});
