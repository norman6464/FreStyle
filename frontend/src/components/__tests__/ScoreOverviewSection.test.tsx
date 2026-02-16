import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import ScoreOverviewSection from '../ScoreOverviewSection';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock('../ScoreGoalCard', () => ({
  default: ({ averageScore }: { averageScore: number }) => (
    <div data-testid="score-goal-card">目標: {averageScore}</div>
  ),
}));

const mockHistory = [
  {
    sessionId: 1,
    sessionTitle: 'テスト',
    overallScore: 7.5,
    scores: [{ axis: '論理的構成力', score: 8, comment: '良い' }],
    createdAt: '2026-01-15T10:00:00Z',
  },
];

const latestSession = mockHistory[0];

describe('ScoreOverviewSection', () => {
  it('最新スコアが表示される', () => {
    render(
      <ScoreOverviewSection
        history={mockHistory}
        latestSession={latestSession}
        averageScore={7.5}
        weakestAxis={{ axis: '論理的構成力', score: 8, comment: '良い' }}
        scoreGoal={8.0}
      />
    );
    expect(screen.getAllByText('7.5').length).toBeGreaterThanOrEqual(1);
  });

  it('統計サマリーが表示される', () => {
    render(
      <ScoreOverviewSection
        history={mockHistory}
        latestSession={latestSession}
        averageScore={7.5}
        weakestAxis={null}
        scoreGoal={8.0}
      />
    );
    expect(screen.getByText('総セッション')).toBeInTheDocument();
  });

  it('weakestAxisがnullの場合はWeakAxisAdviceCardが表示されない', () => {
    render(
      <ScoreOverviewSection
        history={mockHistory}
        latestSession={latestSession}
        averageScore={7.5}
        weakestAxis={null}
        scoreGoal={8.0}
      />
    );
    expect(screen.queryByText(/弱点/)).not.toBeInTheDocument();
  });

  it('scoresが空の場合はSkillGapAnalysisCard等が表示されない', () => {
    const emptyScoresSession = { ...latestSession, scores: [] };

    render(
      <ScoreOverviewSection
        history={[emptyScoresSession]}
        latestSession={emptyScoresSession}
        averageScore={7.5}
        weakestAxis={null}
        scoreGoal={8.0}
      />
    );
    expect(screen.queryByText('最新のスキルバランス')).not.toBeInTheDocument();
  });

  it('historyが2件以上の場合はSkillRadarOverlayCardが表示される', () => {
    const twoHistory = [
      {
        sessionId: 1,
        sessionTitle: 'セッション1',
        overallScore: 6.0,
        scores: [{ axis: '論理的構成力', score: 6, comment: 'まあまあ' }],
        createdAt: '2026-01-14T10:00:00Z',
      },
      {
        sessionId: 2,
        sessionTitle: 'セッション2',
        overallScore: 7.5,
        scores: [{ axis: '論理的構成力', score: 8, comment: '良い' }],
        createdAt: '2026-01-15T10:00:00Z',
      },
    ];

    render(
      <ScoreOverviewSection
        history={twoHistory}
        latestSession={twoHistory[1]}
        averageScore={6.75}
        weakestAxis={null}
        scoreGoal={8.0}
      />
    );
    expect(screen.getByText('スキル変化レーダー')).toBeInTheDocument();
  });

  it('historyが1件のみの場合はScoreComparisonCardが表示されない', () => {
    render(
      <ScoreOverviewSection
        history={mockHistory}
        latestSession={latestSession}
        averageScore={7.5}
        weakestAxis={null}
        scoreGoal={8.0}
      />
    );
    expect(screen.queryByText('スキル変化レーダー')).not.toBeInTheDocument();
  });
});
