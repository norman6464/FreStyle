import { describe, it, expect, vi, beforeEach } from 'vitest';
import apiClient from '../../lib/axios';
import { SharedSessionRepository } from '../SharedSessionRepository';

vi.mock('../../lib/axios');
const mockedApiClient = vi.mocked(apiClient);

describe('SharedSessionRepository', () => {
  beforeEach(() => vi.clearAllMocks());

  it('公開セッション一覧を取得する', async () => {
    mockedApiClient.get.mockResolvedValue({ data: [] });
    const result = await SharedSessionRepository.fetchPublicSessions();
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/shared-sessions');
    expect(result).toEqual([]);
  });

  it('セッションを共有する', async () => {
    const mockShared = { id: 1, sessionId: 10, sessionTitle: 'test' };
    mockedApiClient.post.mockResolvedValue({ data: mockShared });
    const result = await SharedSessionRepository.shareSession(10, 'desc');
    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/shared-sessions', { sessionId: 10, description: 'desc' });
    expect(result).toEqual(mockShared);
  });

  it('セッション共有を解除する', async () => {
    mockedApiClient.delete.mockResolvedValue({});
    await SharedSessionRepository.unshareSession(10);
    expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/shared-sessions/10');
  });
});
