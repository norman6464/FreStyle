import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import RecommendedScenarioCard from '../RecommendedScenarioCard';
import PracticeRepository from '../../repositories/PracticeRepository';

vi.mock('../../repositories/PracticeRepository');

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return { ...actual, useNavigate: () => mockNavigate };
});

const mockedRepo = vi.mocked(PracticeRepository);

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

  it('ボタンクリックでセッション作成後にAIチャットページに遷移する', async () => {
    mockedRepo.createPracticeSession.mockResolvedValue({ id: 42 });

    render(
      <MemoryRouter>
        <RecommendedScenarioCard scenario={defaultScenario} weakAxis="提案力" />
      </MemoryRouter>
    );

    screen.getByText('このシナリオで練習する').click();

    await waitFor(() => {
      expect(mockedRepo.createPracticeSession).toHaveBeenCalledWith({ scenarioId: 1 });
      expect(mockNavigate).toHaveBeenCalledWith('/chat/ask-ai/42', {
        state: {
          sessionType: 'practice',
          scenarioId: 1,
          scenarioName: '本番障害の緊急報告',
          initialPrompt: '練習開始',
        },
      });
    });
  });

  it('セッション作成失敗時は練習一覧に遷移する', async () => {
    mockedRepo.createPracticeSession.mockRejectedValue(new Error('失敗'));

    render(
      <MemoryRouter>
        <RecommendedScenarioCard scenario={defaultScenario} weakAxis="要約力" />
      </MemoryRouter>
    );

    screen.getByText('このシナリオで練習する').click();

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/practice');
    });
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
