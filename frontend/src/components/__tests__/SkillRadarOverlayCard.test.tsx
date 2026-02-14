import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import SkillRadarOverlayCard from '../SkillRadarOverlayCard';

const prevScores = [
  { axis: '論理的構成力', score: 5, comment: '' },
  { axis: '配慮表現', score: 6, comment: '' },
  { axis: '要約力', score: 4, comment: '' },
  { axis: '提案力', score: 7, comment: '' },
  { axis: '質問・傾聴力', score: 5, comment: '' },
];

const currentScores = [
  { axis: '論理的構成力', score: 7, comment: '' },
  { axis: '配慮表現', score: 8, comment: '' },
  { axis: '要約力', score: 6, comment: '' },
  { axis: '提案力', score: 8, comment: '' },
  { axis: '質問・傾聴力', score: 7, comment: '' },
];

describe('SkillRadarOverlayCard', () => {
  it('タイトルが表示される', () => {
    render(<SkillRadarOverlayCard previousScores={prevScores} currentScores={currentScores} />);
    expect(screen.getByText('スキル変化レーダー')).toBeInTheDocument();
  });

  it('SVGが表示される', () => {
    const { container } = render(
      <SkillRadarOverlayCard previousScores={prevScores} currentScores={currentScores} />
    );
    expect(container.querySelector('svg')).toBeTruthy();
  });

  it('前回と今回の2つのポリゴンが表示される', () => {
    const { container } = render(
      <SkillRadarOverlayCard previousScores={prevScores} currentScores={currentScores} />
    );
    const polygons = container.querySelectorAll('polygon');
    // グリッドポリゴン(5) + 前回(1) + 今回(1) = 7
    const dataPolygons = Array.from(polygons).filter(p => p.getAttribute('data-testid'));
    expect(dataPolygons).toHaveLength(2);
  });

  it('凡例が表示される', () => {
    render(<SkillRadarOverlayCard previousScores={prevScores} currentScores={currentScores} />);
    expect(screen.getByText('前回')).toBeInTheDocument();
    expect(screen.getByText('今回')).toBeInTheDocument();
  });

  it('軸ラベルが表示される', () => {
    const { container } = render(
      <SkillRadarOverlayCard previousScores={prevScores} currentScores={currentScores} />
    );
    const texts = container.querySelectorAll('text');
    const labels = Array.from(texts).map(t => t.textContent);
    expect(labels).toContain('論理的構成力');
    expect(labels).toContain('配慮表現');
  });

  it('スコアが空の場合はポリゴンが表示されない', () => {
    const { container } = render(
      <SkillRadarOverlayCard previousScores={[]} currentScores={[]} />
    );
    const dataPolygons = container.querySelectorAll('[data-testid]');
    expect(dataPolygons).toHaveLength(0);
  });
});
