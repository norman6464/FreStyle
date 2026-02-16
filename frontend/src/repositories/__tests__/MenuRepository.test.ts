import { describe, it, expect, vi, afterEach } from 'vitest';
import apiClient from '../../lib/axios';
import { MenuRepository } from '../MenuRepository';

vi.mock('../../lib/axios');

const mockedApiClient = vi.mocked(apiClient);

describe('MenuRepository', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('チャット統計を取得できる', async () => {
    mockedApiClient.get.mockResolvedValue({ data: { chatPartnerCount: 5 } });

    const result = await MenuRepository.fetchChatStats();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/chat/stats');
    expect(result).toEqual({ chatPartnerCount: 5 });
  });

  it('チャットルーム一覧を取得できる', async () => {
    const mockRooms = { chatUsers: [{ roomId: 1, unreadCount: 3 }] };
    mockedApiClient.get.mockResolvedValue({ data: mockRooms });

    const result = await MenuRepository.fetchChatRooms();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/chat/rooms');
    expect(result).toEqual(mockRooms);
  });

  it('スコア履歴を取得できる', async () => {
    const mockScores = [{ sessionId: 1, overallScore: 7.5 }];
    mockedApiClient.get.mockResolvedValue({ data: mockScores });

    const result = await MenuRepository.fetchScoreHistory();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/scores/history');
    expect(result).toEqual(mockScores);
  });

  it('fetchChatStats: APIエラーが呼び出し元に伝播する', async () => {
    mockedApiClient.get.mockRejectedValue(new Error('Network Error'));

    await expect(MenuRepository.fetchChatStats()).rejects.toThrow('Network Error');
  });

  it('fetchChatRooms: 複数のチャットルームを取得できる', async () => {
    const mockRooms = { chatUsers: [{ roomId: 1, unreadCount: 3 }, { roomId: 2, unreadCount: 0 }] };
    mockedApiClient.get.mockResolvedValue({ data: mockRooms });

    const result = await MenuRepository.fetchChatRooms();

    expect(result.chatUsers).toHaveLength(2);
    expect(result.chatUsers[0].unreadCount).toBe(3);
    expect(result.chatUsers[1].unreadCount).toBe(0);
  });
});
