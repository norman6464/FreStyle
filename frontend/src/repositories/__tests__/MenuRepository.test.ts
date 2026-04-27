import { describe, it, expect, vi, afterEach } from 'vitest';
import apiClient from '../../lib/axios';
import { MenuRepository } from '../MenuRepository';

vi.mock('../../lib/axios');

const mockedApiClient = vi.mocked(apiClient);

describe('MenuRepository', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('チャット統計は chat/rooms の件数から導出する', async () => {
    mockedApiClient.get.mockResolvedValue({ data: { chatUsers: [{ roomId: 1, unreadCount: 0 }, { roomId: 2, unreadCount: 1 }] } });

    const result = await MenuRepository.fetchChatStats();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v2/chat/rooms');
    expect(result).toEqual({ chatPartnerCount: 2 });
  });

  it('チャットルーム一覧を取得できる', async () => {
    const mockRooms = { chatUsers: [{ roomId: 1, unreadCount: 3 }] };
    mockedApiClient.get.mockResolvedValue({ data: mockRooms });

    const result = await MenuRepository.fetchChatRooms();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v2/chat/rooms');
    expect(result).toEqual(mockRooms);
  });

  it('スコア履歴を取得できる', async () => {
    const mockScores = [{ sessionId: 1, overallScore: 7.5 }];
    mockedApiClient.get.mockResolvedValue({ data: mockScores });

    const result = await MenuRepository.fetchScoreHistory();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v2/score-cards');
    expect(result).toEqual(mockScores);
  });

  it('fetchChatStats: API エラーは握りつぶして 0 を返す（暫定対応）', async () => {
    mockedApiClient.get.mockRejectedValue(new Error('Network Error'));

    await expect(MenuRepository.fetchChatStats()).resolves.toEqual({ chatPartnerCount: 0 });
  });
});
