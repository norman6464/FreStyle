import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import MenuPage from '../MenuPage';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('../../components/DailyGoalCard', () => ({
  default: () => <div data-testid="daily-goal-card">DailyGoalCard</div>,
}));

// フェッチモック
const mockFetch = vi.fn();
global.fetch = mockFetch;

function createFetchResponse(data: unknown, ok = true, status = 200) {
  return Promise.resolve({
    ok,
    status,
    json: () => Promise.resolve(data),
  });
}

describe('MenuPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/chat/stats')) {
        return createFetchResponse({ chatPartnerCount: 5 });
      }
      if (url.includes('/api/chat/rooms')) {
        return createFetchResponse({
          chatUsers: [
            { roomId: 1, unreadCount: 3 },
            { roomId: 2, unreadCount: 2 },
          ],
        });
      }
      if (url.includes('/api/scores/history')) {
        return createFetchResponse([
          { sessionId: 1, sessionTitle: 'テスト', overallScore: 7.5, createdAt: '2026-02-13' },
        ]);
      }
      return createFetchResponse({});
    });
  });

  it('メニュー項目が全て表示される', async () => {
    render(<BrowserRouter><MenuPage /></BrowserRouter>);

    await waitFor(() => {
      expect(screen.getByText('チャット')).toBeInTheDocument();
      expect(screen.getByText('AI アシスタント')).toBeInTheDocument();
      expect(screen.getByText('練習モード')).toBeInTheDocument();
      expect(screen.getByText('スコア履歴')).toBeInTheDocument();
    });
  });

  it('会話した人数を表示する', async () => {
    render(<BrowserRouter><MenuPage /></BrowserRouter>);

    await waitFor(() => {
      expect(screen.getByText('5')).toBeInTheDocument();
    });
  });

  it('未読メッセージ数バッジをチャットカードに表示する', async () => {
    render(<BrowserRouter><MenuPage /></BrowserRouter>);

    await waitFor(() => {
      expect(screen.getByText('5件の未読')).toBeInTheDocument();
    });
  });

  it('最新スコアをスコア履歴カードに表示する', async () => {
    render(<BrowserRouter><MenuPage /></BrowserRouter>);

    await waitFor(() => {
      expect(screen.getByText(/最新: 7\.5/)).toBeInTheDocument();
    });
  });

  it('スコア履歴がない場合はおすすめアクションを表示する', async () => {
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/chat/stats')) {
        return createFetchResponse({ chatPartnerCount: 0 });
      }
      if (url.includes('/api/chat/rooms')) {
        return createFetchResponse({ chatUsers: [] });
      }
      if (url.includes('/api/scores/history')) {
        return createFetchResponse([]);
      }
      return createFetchResponse({});
    });

    render(<BrowserRouter><MenuPage /></BrowserRouter>);

    await waitFor(() => {
      expect(screen.getByText(/練習モードから始めてみましょう/)).toBeInTheDocument();
    });
  });

  it('未読がない場合は未読バッジを表示しない', async () => {
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/chat/stats')) {
        return createFetchResponse({ chatPartnerCount: 3 });
      }
      if (url.includes('/api/chat/rooms')) {
        return createFetchResponse({ chatUsers: [{ roomId: 1, unreadCount: 0 }] });
      }
      if (url.includes('/api/scores/history')) {
        return createFetchResponse([]);
      }
      return createFetchResponse({});
    });

    render(<BrowserRouter><MenuPage /></BrowserRouter>);

    await waitFor(() => {
      expect(screen.getByText('チャット')).toBeInTheDocument();
    });

    expect(screen.queryByText(/件の未読/)).not.toBeInTheDocument();
  });
});
