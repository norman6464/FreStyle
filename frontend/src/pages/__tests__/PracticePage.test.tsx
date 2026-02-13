import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import PracticePage from '../PracticePage';

const mockNavigate = vi.fn();
const mockFetchScenarios = vi.fn();
const mockCreatePracticeSession = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('../../hooks/usePractice', () => ({
  usePractice: () => ({
    scenarios: [
      {
        id: 1,
        name: '本番障害の緊急報告',
        description: '本番環境で重大な障害が発生',
        category: 'customer',
        roleName: '怒っている顧客（SIer企業のPM）',
        difficulty: 'intermediate',
      },
    ],
    loading: false,
    fetchScenarios: mockFetchScenarios,
    createPracticeSession: mockCreatePracticeSession,
  }),
}));

describe('PracticePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreatePracticeSession.mockResolvedValue({ id: 123, title: '練習: 本番障害の緊急報告' });
  });

  describe('シナリオカードクリック時の動作', () => {
    it('シナリオカードクリック時にinitialPromptを含むstateでnavigateする', async () => {
      render(
        <BrowserRouter>
          <PracticePage />
        </BrowserRouter>
      );

      await waitFor(() => {
        expect(screen.getByText('本番障害の緊急報告')).toBeInTheDocument();
      });

      const scenarioCard = screen.getByText('本番障害の緊急報告').closest('div[class*="cursor-pointer"]');
      if (scenarioCard) {
        fireEvent.click(scenarioCard);
      }

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith(
          '/chat/ask-ai/123',
          expect.objectContaining({
            state: expect.objectContaining({
              sessionType: 'practice',
              scenarioId: 1,
              scenarioName: '本番障害の緊急報告',
              initialPrompt: '練習開始',
            }),
          })
        );
      });
    });
  });
});
