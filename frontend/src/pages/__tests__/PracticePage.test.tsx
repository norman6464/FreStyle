import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import PracticePage from '../PracticePage';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

global.fetch = vi.fn();

describe('PracticePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => ([
        {
          id: 1,
          name: '本番障害の緊急報告',
          description: '本番環境で重大な障害が発生',
          category: 'customer',
          roleName: '怒っている顧客（SIer企業のPM）',
          difficulty: 'intermediate',
        },
      ]),
    } as Response);
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

      // セッション作成APIのモック
      (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ id: 123, title: '練習: 本番障害の緊急報告' }),
      } as Response);

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
