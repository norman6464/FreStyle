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

  it('createRoom: チャットルームを作成できる', async () => {
    const mockData = { roomId: 42 };
    mockedApiClient.post.mockResolvedValue({ data: mockData });

    const result = await ChatRepository.createRoom(5);

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/chat/users/5/create');
    expect(result).toEqual(mockData);
  });

  it('markAsRead: 既読をマークできる', async () => {
    mockedApiClient.post.mockResolvedValue({ data: undefined });

    await ChatRepository.markAsRead('10');

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/chat/rooms/10/read');
  });

  it('fetchHistory: チャット履歴を取得できる', async () => {
    const mockMessages = [
      { id: 1, content: 'Hello', senderId: 1 },
      { id: 2, content: 'Hi', senderId: 2 },
    ];
    mockedApiClient.get.mockResolvedValue({ data: mockMessages });

    const result = await ChatRepository.fetchHistory('5');

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/chat/users/5/history');
    expect(result).toEqual(mockMessages);
  });

  it('rephrase: 言い換え結果を取得できる', async () => {
    const mockResult = { result: '{"formal":"丁寧版","soft":"柔らかい版","concise":"簡潔版"}' };
    mockedApiClient.post.mockResolvedValue({ data: mockResult });

    const result = await ChatRepository.rephrase('テスト文', null);

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/chat/ai/rephrase', {
      originalMessage: 'テスト文',
      scene: null,
    });
    expect(result).toEqual(mockResult);
  });
});
