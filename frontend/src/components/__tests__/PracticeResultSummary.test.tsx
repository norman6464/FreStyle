import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import PracticeResultSummary from '../PracticeResultSummary';
import type { ScoreCard } from '../../types';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

const mockStartSession = vi.fn();
vi.mock('../../hooks/useStartPracticeSession', () => ({
  useStartPracticeSession: () => ({ startSession: mockStartSession, starting: false }),
}));

describe('PracticeResultSummary', () => {
  const scoreCard: ScoreCard = {
    sessionId: 1,
    overallScore: 7.2,
    scores: [
      { axis: '論理的構成力', score: 9, comment: '非常に論理的です' },
      { axis: '配慮表現', score: 8, comment: '丁寧な表現です' },
      { axis: '要約力', score: 7, comment: '概ね良い要約です' },
      { axis: '提案力', score: 4, comment: '提案が不足しています' },
      { axis: '質問・傾聴力', score: 6, comment: '質問は適切です' },
    ],
  };

  beforeEach(() => {
    mockNavigate.mockClear();
    mockStartSession.mockClear();
  });

  it('練習結果サマリーのタイトルが表示される', () => {
    render(<PracticeResultSummary scoreCard={scoreCard} scenarioName="障害報告" />);

    expect(screen.getByText('練習結果サマリー')).toBeInTheDocument();
  });

  it('最も高い評価軸が強みとして表示される', () => {
    render(<PracticeResultSummary scoreCard={scoreCard} scenarioName="障害報告" />);

    expect(screen.getByText('強み')).toBeInTheDocument();
    expect(screen.getByText('論理的構成力')).toBeInTheDocument();
  });

  it('最も低い評価軸が課題として表示される', () => {
    render(<PracticeResultSummary scoreCard={scoreCard} scenarioName="障害報告" />);

    expect(screen.getByText('課題')).toBeInTheDocument();
    expect(screen.getByText('提案力')).toBeInTheDocument();
  });

  it('改善アドバイスが表示される', () => {
    render(<PracticeResultSummary scoreCard={scoreCard} scenarioName="障害報告" />);

    expect(screen.getByText(/課題だけでなく解決策をセットで伝える/)).toBeInTheDocument();
  });

  it('次の練習へボタンをクリックすると練習ページに遷移する', () => {
    render(<PracticeResultSummary scoreCard={scoreCard} scenarioName="障害報告" />);

    fireEvent.click(screen.getByText('次の練習へ'));
    expect(mockNavigate).toHaveBeenCalledWith('/practice');
  });

  it('シナリオ名が表示される', () => {
    render(<PracticeResultSummary scoreCard={scoreCard} scenarioName="障害報告" />);

    expect(screen.getByText(/障害報告/)).toBeInTheDocument();
  });

  it('scoresがnullでもエラーにならない', () => {
    const nullCard: ScoreCard = {
      sessionId: 1,
      overallScore: 7.0,
      scores: null as unknown as ScoreCard['scores'],
    };
    render(<PracticeResultSummary scoreCard={nullCard} scenarioName="テスト" />);
    expect(document.body).toBeTruthy();
  });

  it('全軸のスコアが同じ場合でも正しく表示される', () => {
    const equalScores: ScoreCard = {
      sessionId: 2,
      overallScore: 7.0,
      scores: [
        { axis: '論理的構成力', score: 7, comment: 'コメント' },
        { axis: '配慮表現', score: 7, comment: 'コメント' },
        { axis: '要約力', score: 7, comment: 'コメント' },
        { axis: '提案力', score: 7, comment: 'コメント' },
        { axis: '質問・傾聴力', score: 7, comment: 'コメント' },
      ],
    };

    render(<PracticeResultSummary scoreCard={equalScores} scenarioName="テスト" />);

    expect(screen.getByText('強み')).toBeInTheDocument();
    expect(screen.getByText('課題')).toBeInTheDocument();
  });

  it('scenarioIdがある場合「もう一度練習」ボタンが表示される', () => {
    render(<PracticeResultSummary scoreCard={scoreCard} scenarioName="障害報告" scenarioId={5} />);

    expect(screen.getByRole('button', { name: 'もう一度練習' })).toBeInTheDocument();
  });

  it('「もう一度練習」ボタンをクリックすると同じシナリオでセッション開始する', () => {
    render(<PracticeResultSummary scoreCard={scoreCard} scenarioName="障害報告" scenarioId={5} />);

    fireEvent.click(screen.getByRole('button', { name: 'もう一度練習' }));
    expect(mockStartSession).toHaveBeenCalledWith({ id: 5, name: '障害報告' });
  });

  it('scenarioIdがない場合「もう一度練習」ボタンは表示されない', () => {
    render(<PracticeResultSummary scoreCard={scoreCard} scenarioName="障害報告" />);

    expect(screen.queryByRole('button', { name: 'もう一度練習' })).not.toBeInTheDocument();
  });
});
