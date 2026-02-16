import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import RecommendedScenarioCard from '../RecommendedScenarioCard';

const mockStartSession = vi.fn();
vi.mock('../../hooks/useStartPracticeSession', () => ({
  useStartPracticeSession: () => ({
    startSession: mockStartSession,
    starting: false,
  }),
}));

const defaultScenario = {
  id: 1,
  name: '本番障害の緊急報告',
  description: '本番環境で障害が発生し、顧客に緊急報告する場面です。',
  category: 'customer',
  roleName: '顧客',
  difficulty: 'intermediate',
};

describe('RecommendedScenarioCard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('シナリオ名と説明が表示される', () => {
    render(
      <MemoryRouter>
        <RecommendedScenarioCard scenario={defaultScenario} weakAxis="論理的構成力" />
      </MemoryRouter>
    );

    expect(screen.getByText('本番障害の緊急報告')).toBeInTheDocument();
    expect(screen.getByText('本番環境で障害が発生し、顧客に緊急報告する場面です。')).toBeInTheDocument();
  });

  it('弱点軸名が表示される', () => {
    render(
      <MemoryRouter>
        <RecommendedScenarioCard scenario={defaultScenario} weakAxis="配慮表現" />
      </MemoryRouter>
    );

    expect(screen.getByText(/配慮表現/)).toBeInTheDocument();
  });

  it('チャレンジボタンが表示される', () => {
    render(
      <MemoryRouter>
        <RecommendedScenarioCard scenario={defaultScenario} weakAxis="提案力" />
      </MemoryRouter>
    );

    expect(screen.getByText('このシナリオで練習する')).toBeInTheDocument();
  });

  it('ボタンクリックでstartSessionが呼ばれる', () => {
    render(
      <MemoryRouter>
        <RecommendedScenarioCard scenario={defaultScenario} weakAxis="提案力" />
      </MemoryRouter>
    );

    screen.getByText('このシナリオで練習する').click();

    expect(mockStartSession).toHaveBeenCalledWith(defaultScenario);
  });

  it('難易度バッジが表示される', () => {
    render(
      <MemoryRouter>
        <RecommendedScenarioCard scenario={defaultScenario} weakAxis="要約力" />
      </MemoryRouter>
    );

    expect(screen.getByText('中級')).toBeInTheDocument();
  });
});
