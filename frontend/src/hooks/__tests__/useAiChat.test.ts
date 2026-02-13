import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useAiChat } from '../useAiChat';
import AiChatRepository from '../../repositories/AiChatRepository';

vi.mock('../../repositories/AiChatRepository');

const mockedRepo = vi.mocked(AiChatRepository);

describe('useAiChat', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchSessions: セッション一覧を取得してstateに保存する', async () => {
    const mockSessions = [{ id: 1, title: 'テスト' }];
    mockedRepo.getSessions.mockResolvedValue(mockSessions as any);

    const { result } = renderHook(() => useAiChat());

    await act(async () => {
      await result.current.fetchSessions();
    });

    expect(result.current.sessions).toEqual(mockSessions);
    expect(result.current.loading).toBe(false);
  });

  it('fetchSessions: エラー時にerrorを設定する', async () => {
    mockedRepo.getSessions.mockRejectedValue(new Error('取得失敗'));

    const { result } = renderHook(() => useAiChat());

    await act(async () => {
      await result.current.fetchSessions();
    });

    expect(result.current.error).toBe('取得失敗');
    expect(result.current.loading).toBe(false);
  });

  it('createSession: セッションを作成してsessionsに追加する', async () => {
    const mockSession = { id: 2, title: '新セッション' };
    mockedRepo.createSession.mockResolvedValue(mockSession as any);

    const { result } = renderHook(() => useAiChat());

    let created: any;
    await act(async () => {
      created = await result.current.createSession({ title: '新セッション' });
    });

    expect(created).toEqual(mockSession);
    expect(result.current.sessions).toContainEqual(mockSession);
    expect(result.current.currentSession).toEqual(mockSession);
  });

  it('deleteSession: セッションを削除してsessionsから除外する', async () => {
    const mockSessions = [{ id: 1, title: 'A' }, { id: 2, title: 'B' }];
    mockedRepo.getSessions.mockResolvedValue(mockSessions as any);
    mockedRepo.deleteSession.mockResolvedValue(undefined);

    const { result } = renderHook(() => useAiChat());

    await act(async () => {
      await result.current.fetchSessions();
    });

    await act(async () => {
      await result.current.deleteSession(1);
    });

    expect(result.current.sessions).toHaveLength(1);
    expect(result.current.sessions[0]).toEqual({ id: 2, title: 'B' });
  });

  it('fetchMessages: メッセージ一覧を取得する', async () => {
    const mockMessages = [{ id: 1, content: 'テスト', role: 'user' }];
    mockedRepo.getMessages.mockResolvedValue(mockMessages as any);

    const { result } = renderHook(() => useAiChat());

    await act(async () => {
      await result.current.fetchMessages(1);
    });

    expect(result.current.messages).toEqual(mockMessages);
  });

  it('addMessage: メッセージを追加してmessagesに追記する', async () => {
    const mockMessage = { id: 2, content: '新メッセージ', role: 'user' };
    mockedRepo.addMessage.mockResolvedValue(mockMessage as any);

    const { result } = renderHook(() => useAiChat());

    await act(async () => {
      await result.current.addMessage(1, { content: '新メッセージ', role: 'user' });
    });

    expect(result.current.messages).toContainEqual(mockMessage);
  });

  it('rephrase: 言い換え結果を返す', async () => {
    mockedRepo.rephrase.mockResolvedValue({ result: '言い換え文' });

    const { result } = renderHook(() => useAiChat());

    let rephrased: string | null = null;
    await act(async () => {
      rephrased = await result.current.rephrase({ originalMessage: '元の文' });
    });

    expect(rephrased).toBe('言い換え文');
  });

  it('fetchScoreCard: スコアカードを取得する', async () => {
    const mockScoreCard = { overallScore: 7.5, scores: [] };
    mockedRepo.getScoreCard.mockResolvedValue(mockScoreCard as any);

    const { result } = renderHook(() => useAiChat());

    await act(async () => {
      await result.current.fetchScoreCard(1);
    });

    expect(result.current.scoreCard).toEqual(mockScoreCard);
  });
});
