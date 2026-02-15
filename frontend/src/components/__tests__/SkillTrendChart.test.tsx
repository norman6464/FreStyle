import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import SkillTrendChart from '../SkillTrendChart';

const mockHistory = [
  {
    sessionId: 1,
    scores: [
      { axis: '論理的構成力', score: 6 },
      { axis: '配慮表現', score: 7 },
      { axis: '要約力', score: 5 },
      { axis: '提案力', score: 4 },
      { axis: '質問・傾聴力', score: 8 },
    ],
  },
  {
    sessionId: 2,
    scores: [
      { axis: '論理的構成力', score: 8 },
      { axis: '配慮表現', score: 7 },
      { axis: '要約力', score: 6 },
      { axis: '提案力', score: 5 },
      { axis: '質問・傾聴力', score: 9 },
    ],
  },
];

describe('SkillTrendChart', () => {
  it('タイトルが表示される', () => {
    render(<SkillTrendChart history={mockHistory} />);

    expect(screen.getByText('スキル別推移')).toBeInTheDocument();
  });

  it('各スキル軸名が表示される', () => {
    render(<SkillTrendChart history={mockHistory} />);

    expect(screen.getByText('論理的構成力')).toBeInTheDocument();
    expect(screen.getByText('配慮表現')).toBeInTheDocument();
    expect(screen.getByText('要約力')).toBeInTheDocument();
    expect(screen.getByText('提案力')).toBeInTheDocument();
    expect(screen.getByText('質問・傾聴力')).toBeInTheDocument();
  });

  it('最新スコアが表示される', () => {
    render(<SkillTrendChart history={mockHistory} />);

    // 最新セッション(sessionId:2)のスコア
    const scoreElements = screen.getAllByTestId('skill-latest-score');
    expect(scoreElements).toHaveLength(5);
  });

  it('スコア変動が表示される', () => {
    render(<SkillTrendChart history={mockHistory} />);

    // 論理的構成力: 6→8 = +2
    expect(screen.getByText('+2')).toBeInTheDocument();
  });

  it('scoresがnullのセッションでもエラーにならない', () => {
    const history = [
      { sessionId: 1, scores: null as unknown as { axis: string; score: number }[] },
    ];
    const { container } = render(<SkillTrendChart history={history} />);
    expect(container.firstChild).toBeNull();
  });

  it('前回セッションのscoresがnullでもエラーにならない', () => {
    const history = [
      { sessionId: 1, scores: null as unknown as { axis: string; score: number }[] },
      {
        sessionId: 2,
        scores: [
          { axis: '論理的構成力', score: 8 },
        ],
      },
    ];
    render(<SkillTrendChart history={history} />);
    expect(screen.getByText('論理的構成力')).toBeInTheDocument();
  });

  it('履歴が空の場合は何も表示しない', () => {
    const { container } = render(<SkillTrendChart history={[]} />);

    expect(container.firstChild).toBeNull();
  });
});
