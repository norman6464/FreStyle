import { describe, it, expect, vi, beforeEach } from 'vitest';
import ChatRepository from '../ChatRepository';
import apiClient from '../../lib/axios';

vi.mock('../../lib/axios');

const mockedApiClient = vi.mocked(apiClient);

describe('ChatRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchChatUsers: チャットユーザー一覧を取得できる', async () => {
    const mockData = {
      chatUsers: [
        { roomId: 1, userId: 1, name: 'テスト', unreadCount: 0 },
      ],
    };
    mockedApiClient.get.mockResolvedValue({ data: mockData });

    const result = await ChatRepository.fetchChatUsers();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/chat/rooms', { params: {} });
    expect(result).toEqual(mockData.chatUsers);
  });

  it('fetchChatUsers: 検索クエリ付きで取得できる', async () => {
    const mockData = { chatUsers: [] };
    mockedApiClient.get.mockResolvedValue({ data: mockData });

    await ChatRepository.fetchChatUsers('テスト');

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/chat/rooms', { params: { query: 'テスト' } });
  });

  it('fetchCurrentUser: 現在のユーザー情報を取得できる', async () => {
    const mockUser = { id: 1, name: 'テスト' };
    mockedApiClient.get.mockResolvedValue({ data: mockUser });

    const result = await ChatRepository.fetchCurrentUser();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/auth/cognito/me');
    expect(result).toEqual(mockUser);
  });
});
