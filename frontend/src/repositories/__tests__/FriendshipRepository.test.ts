import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../lib/axios', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

import apiClient from '../../lib/axios';
import { FriendshipRepository } from '../FriendshipRepository';

const mockedGet = vi.mocked(apiClient.get);
const mockedPost = vi.mocked(apiClient.post);
const mockedDelete = vi.mocked(apiClient.delete);

describe('FriendshipRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('getFollowing: フォロー中一覧を取得する', async () => {
    const users = [{ id: 1, userId: 10, username: 'user1', iconUrl: null, bio: null, mutual: false, createdAt: '2024-01-01', status: null }];
    mockedGet.mockResolvedValue({ data: users });

    const result = await FriendshipRepository.getFollowing();
    expect(result).toEqual(users);
    expect(mockedGet).toHaveBeenCalledWith('/api/friendships/following');
  });

  it('getFollowers: フォロワー一覧を取得する', async () => {
    const users = [{ id: 2, userId: 20, username: 'user2', iconUrl: null, bio: null, mutual: true, createdAt: '2024-01-01', status: null }];
    mockedGet.mockResolvedValue({ data: users });

    const result = await FriendshipRepository.getFollowers();
    expect(result).toEqual(users);
    expect(mockedGet).toHaveBeenCalledWith('/api/friendships/followers');
  });

  it('follow: フォローリクエストを送る', async () => {
    const user = { id: 3, userId: 30, username: 'user3', iconUrl: null, bio: null, mutual: false, createdAt: '2024-01-01', status: null };
    mockedPost.mockResolvedValue({ data: user });

    const result = await FriendshipRepository.follow(30);
    expect(result).toEqual(user);
    expect(mockedPost).toHaveBeenCalledWith('/api/friendships/30/follow');
  });

  it('unfollow: フォロー解除する', async () => {
    mockedDelete.mockResolvedValue({});

    await FriendshipRepository.unfollow(30);
    expect(mockedDelete).toHaveBeenCalledWith('/api/friendships/30/follow');
  });

  it('checkStatus: フォロー状態を確認する', async () => {
    const status = { isFollowing: true, isFollowedBy: false, isMutual: false };
    mockedGet.mockResolvedValue({ data: status });

    const result = await FriendshipRepository.checkStatus(30);
    expect(result).toEqual(status);
    expect(mockedGet).toHaveBeenCalledWith('/api/friendships/30/status');
  });
});
