import { describe, it, expect, vi, beforeEach } from 'vitest';
import aiChatRepository from '../AiChatRepository';
import apiClient from '../../lib/axios';

vi.mock('../../lib/axios');

const mockedApiClient = vi.mocked(apiClient);

describe('AiChatRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('getSessions: セッション一覧を取得できる', async () => {
    const mockSessions = [{ id: 1, title: 'テストセッション' }];
    mockedApiClient.get.mockResolvedValue({ data: mockSessions });

    const result = await aiChatRepository.getSessions();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/chat/ai/sessions');
    expect(result).toEqual(mockSessions);
  });

  it('getSession: セッション詳細を取得できる', async () => {
    const mockSession = { id: 1, title: 'テストセッション' };
    mockedApiClient.get.mockResolvedValue({ data: mockSession });

    const result = await aiChatRepository.getSession(1);

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/chat/ai/sessions/1');
    expect(result).toEqual(mockSession);
  });

  it('createSession: 新規セッションを作成できる', async () => {
    const mockSession = { id: 2, title: '新しいセッション' };
    mockedApiClient.post.mockResolvedValue({ data: mockSession });

    const result = await aiChatRepository.createSession({ title: '新しいセッション' });

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/chat/ai/sessions', { title: '新しいセッション' });
    expect(result).toEqual(mockSession);
  });

  it('updateSessionTitle: セッションタイトルを更新できる', async () => {
    const mockSession = { id: 1, title: '更新タイトル' };
    mockedApiClient.put.mockResolvedValue({ data: mockSession });

    const result = await aiChatRepository.updateSessionTitle(1, { title: '更新タイトル' });

    expect(mockedApiClient.put).toHaveBeenCalledWith('/api/chat/ai/sessions/1', { title: '更新タイトル' });
    expect(result).toEqual(mockSession);
  });

  it('deleteSession: セッションを削除できる', async () => {
    mockedApiClient.delete.mockResolvedValue({});

    await aiChatRepository.deleteSession(1);

    expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/chat/ai/sessions/1');
  });

  it('getMessages: メッセージ一覧を取得できる', async () => {
    const mockMessages = [{ id: 1, content: 'テストメッセージ', role: 'user' }];
    mockedApiClient.get.mockResolvedValue({ data: mockMessages });

    const result = await aiChatRepository.getMessages(1);

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/chat/ai/sessions/1/messages');
    expect(result).toEqual(mockMessages);
  });

  it('addMessage: メッセージを追加できる', async () => {
    const mockMessage = { id: 2, content: '新しいメッセージ', role: 'user' };
    mockedApiClient.post.mockResolvedValue({ data: mockMessage });

    const result = await aiChatRepository.addMessage(1, { content: '新しいメッセージ', role: 'user' });

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/chat/ai/sessions/1/messages', { content: '新しいメッセージ', role: 'user' });
    expect(result).toEqual(mockMessage);
  });

  it('rephrase: 言い換え提案を取得できる', async () => {
    const mockResult = { result: '言い換え結果' };
    mockedApiClient.post.mockResolvedValue({ data: mockResult });

    const result = await aiChatRepository.rephrase({ originalMessage: '元のメッセージ', scene: 'meeting' });

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/chat/ai/rephrase', { originalMessage: '元のメッセージ', scene: 'meeting' });
    expect(result).toEqual(mockResult);
  });

  it('getScoreCard: スコアカードを取得できる', async () => {
    const mockScoreCard = { overallScore: 7.5, scores: [] };
    mockedApiClient.get.mockResolvedValue({ data: mockScoreCard });

    const result = await aiChatRepository.getScoreCard(1);

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/scores/sessions/1');
    expect(result).toEqual(mockScoreCard);
  });

  it('getScoreHistory: スコア履歴を取得できる', async () => {
    const mockHistory = [{ sessionId: 1, overallScore: 7.5 }];
    mockedApiClient.get.mockResolvedValue({ data: mockHistory });

    const result = await aiChatRepository.getScoreHistory();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/scores/history');
    expect(result).toEqual(mockHistory);
  });
});
