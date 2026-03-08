import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useSharedSessions } from '../useSharedSessions';
import { SharedSessionRepository } from '../../repositories/SharedSessionRepository';

vi.mock('../../repositories/SharedSessionRepository');
const mockedRepo = vi.mocked(SharedSessionRepository);

describe('useSharedSessions', () => {
  const mockSessions = [
    { id: 1, sessionId: 10, sessionTitle: 'Test Session', userId: 1, username: 'user1', userIconUrl: null, description: null, createdAt: '2024-01-01' },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    mockedRepo.fetchPublicSessions.mockResolvedValue(mockSessions);
  });

  it('公開セッションを取得する', async () => {
    const { result } = renderHook(() => useSharedSessions());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.sessions).toEqual(mockSessions);
  });

  it('セッションを共有する', async () => {
    const newShared = { ...mockSessions[0], id: 2, sessionId: 20, sessionTitle: 'New' };
    mockedRepo.shareSession.mockResolvedValue(newShared);
    const { result } = renderHook(() => useSharedSessions());
    await waitFor(() => expect(result.current.loading).toBe(false));
    await act(async () => { await result.current.shareSession(20); });
    expect(result.current.sessions).toHaveLength(2);
  });

  it('エラー時にエラーメッセージを設定する', async () => {
    mockedRepo.fetchPublicSessions.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useSharedSessions());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe('共有セッションの取得に失敗しました');
  });
});
