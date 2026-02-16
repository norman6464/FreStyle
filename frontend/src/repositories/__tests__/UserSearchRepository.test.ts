import { describe, it, expect, vi, beforeEach } from 'vitest';
import UserSearchRepository from '../UserSearchRepository';
import apiClient from '../../lib/axios';

vi.mock('../../lib/axios');

describe('UserSearchRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('ユーザー検索結果を返す', async () => {
    const mockUsers = [
      { id: 1, name: '山田太郎', email: 'yamada@example.com', roomId: null },
    ];
    vi.mocked(apiClient.get).mockResolvedValue({ data: { users: mockUsers } });

    const result = await UserSearchRepository.searchUsers('山田');

    expect(apiClient.get).toHaveBeenCalledWith('/api/chat/users', { params: { query: '山田' } });
    expect(result).toEqual(mockUsers);
  });

  it('クエリなしで全ユーザーを返す', async () => {
    vi.mocked(apiClient.get).mockResolvedValue({ data: { users: [] } });

    const result = await UserSearchRepository.searchUsers();

    expect(apiClient.get).toHaveBeenCalledWith('/api/chat/users', { params: {} });
    expect(result).toEqual([]);
  });

  it('usersが未定義の場合は空配列を返す', async () => {
    vi.mocked(apiClient.get).mockResolvedValue({ data: {} });

    const result = await UserSearchRepository.searchUsers();

    expect(result).toEqual([]);
  });

  it('APIエラー時に例外が伝搬する', async () => {
    vi.mocked(apiClient.get).mockRejectedValue(new Error('ネットワークエラー'));

    await expect(UserSearchRepository.searchUsers('テスト')).rejects.toThrow('ネットワークエラー');
  });

  it('複数ユーザーの検索結果を返す', async () => {
    const mockUsers = [
      { id: 1, name: '山田太郎', email: 'yamada@example.com', roomId: null },
      { id: 2, name: '山田花子', email: 'hanako@example.com', roomId: 5 },
    ];
    vi.mocked(apiClient.get).mockResolvedValue({ data: { users: mockUsers } });

    const result = await UserSearchRepository.searchUsers('山田');

    expect(result).toHaveLength(2);
    expect(result[0].name).toBe('山田太郎');
    expect(result[1].name).toBe('山田花子');
  });
});
