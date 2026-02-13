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

const mockScenarios = [
  {
    id: 1,
    name: '本番障害の緊急報告',
    description: '本番環境で重大な障害が発生',
    category: 'customer',
    roleName: '怒っている顧客（SIer企業のPM）',
    difficulty: 'intermediate',
  },
  {
    id: 6,
    name: '設計方針の意見対立',
    description: 'シニアエンジニアとの設計レビュー',
    category: 'senior',
    roleName: '厳格なシニアエンジニア',
    difficulty: 'advanced',
  },
  {
    id: 11,
    name: 'チーム内の進捗共有',
    description: 'チームメンバーとの進捗共有',
    category: 'team',
    roleName: 'チームリーダー',
    difficulty: 'beginner',
  },
];

vi.mock('../../hooks/usePractice', () => ({
  usePractice: () => ({
    scenarios: mockScenarios,
    loading: false,
    fetchScenarios: mockFetchScenarios,
    createPracticeSession: mockCreatePracticeSession,
  }),
}));

vi.mock('../../hooks/useBookmark', () => ({
  useBookmark: () => ({
    bookmarkedIds: [1],
    toggleBookmark: vi.fn(),
    isBookmarked: (id: number) => id === 1,
  }),
}));

describe('PracticePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreatePracticeSession.mockResolvedValue({ id: 123, title: '練習: 本番障害の緊急報告' });
  });

  describe('カテゴリフィルタリング', () => {
    it('「すべて」タブで全シナリオが表示される', () => {
      render(
        <BrowserRouter>
          <PracticePage />
        </BrowserRouter>
      );

      expect(screen.getByText('本番障害の緊急報告')).toBeInTheDocument();
      expect(screen.getByText('設計方針の意見対立')).toBeInTheDocument();
      expect(screen.getByText('チーム内の進捗共有')).toBeInTheDocument();
    });

    it('「顧客折衝」タブでcustomerカテゴリのシナリオのみ表示される', () => {
      render(
        <BrowserRouter>
          <PracticePage />
        </BrowserRouter>
      );

      fireEvent.click(screen.getByRole('button', { name: '顧客折衝' }));

      expect(screen.getByText('本番障害の緊急報告')).toBeInTheDocument();
      expect(screen.queryByText('設計方針の意見対立')).not.toBeInTheDocument();
      expect(screen.queryByText('チーム内の進捗共有')).not.toBeInTheDocument();
    });

    it('「シニア・上司」タブでseniorカテゴリのシナリオのみ表示される', () => {
      render(
        <BrowserRouter>
          <PracticePage />
        </BrowserRouter>
      );

      fireEvent.click(screen.getByRole('button', { name: 'シニア・上司' }));

      expect(screen.queryByText('本番障害の緊急報告')).not.toBeInTheDocument();
      expect(screen.getByText('設計方針の意見対立')).toBeInTheDocument();
      expect(screen.queryByText('チーム内の進捗共有')).not.toBeInTheDocument();
    });

    it('「チーム内」タブでteamカテゴリのシナリオのみ表示される', () => {
      render(
        <BrowserRouter>
          <PracticePage />
        </BrowserRouter>
      );

      fireEvent.click(screen.getByRole('button', { name: 'チーム内' }));

      expect(screen.queryByText('本番障害の緊急報告')).not.toBeInTheDocument();
      expect(screen.queryByText('設計方針の意見対立')).not.toBeInTheDocument();
      expect(screen.getByText('チーム内の進捗共有')).toBeInTheDocument();
    });
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
