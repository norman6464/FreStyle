import { describe, it, expect, vi, afterEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useMenuData } from '../useMenuData';
import { MenuRepository } from '../../repositories/MenuRepository';

vi.mock('../../repositories/MenuRepository');

const mockedRepo = vi.mocked(MenuRepository);

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

describe('useMenuData', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('ダッシュボードデータを取得する', async () => {
    mockedRepo.fetchChatStats.mockResolvedValue({ chatPartnerCount: 5 });
    mockedRepo.fetchChatRooms.mockResolvedValue({
      chatUsers: [
        { roomId: 1, unreadCount: 3 },
        { roomId: 2, unreadCount: 2 },
      ],
    });
    mockedRepo.fetchScoreHistory.mockResolvedValue([
      { sessionId: 1, sessionTitle: 'テスト', overallScore: 7.5, createdAt: '2026-02-13T00:00:00' },
      { sessionId: 2, sessionTitle: 'テスト2', overallScore: 8.5, createdAt: '2026-02-12T00:00:00' },
    ]);

    const { result } = renderHook(() => useMenuData());

    await waitFor(() => {
      expect(result.current.stats).not.toBeNull();
    });

    expect(result.current.stats?.chatPartnerCount).toBe(5);
    expect(result.current.totalUnread).toBe(5);
    expect(result.current.latestScore?.overallScore).toBe(7.5);
    expect(result.current.totalSessions).toBe(2);
    expect(result.current.averageScore).toBe(8.0);
    expect(result.current.uniqueDays).toBe(2);
  });

  it('APIエラー時にクラッシュしない', async () => {
    mockedRepo.fetchChatStats.mockRejectedValue(new Error('Network error'));
    mockedRepo.fetchChatRooms.mockRejectedValue(new Error('Network error'));
    mockedRepo.fetchScoreHistory.mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useMenuData());

    // エラーでもクラッシュしないことを確認
    await waitFor(() => {
      expect(result.current.stats).toBeNull();
    });
  });
});
