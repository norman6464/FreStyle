import apiClient from '../lib/axios';
import type { FriendshipUser, FollowStatus } from '../types';

export const FriendshipRepository = {
  async getFollowing(): Promise<FriendshipUser[]> {
    const response = await apiClient.get<FriendshipUser[]>('/api/v2/friendships/following');
    return response.data;
  },

  async getFollowers(): Promise<FriendshipUser[]> {
    const response = await apiClient.get<FriendshipUser[]>('/api/v2/friendships/followers');
    return response.data;
  },

  async follow(userId: number): Promise<FriendshipUser> {
    const response = await apiClient.post<FriendshipUser>(`/api/v2/friendships/${userId}/follow`);
    return response.data;
  },

  async unfollow(userId: number): Promise<void> {
    await apiClient.delete(`/api/v2/friendships/${userId}/follow`);
  },

  async checkStatus(userId: number): Promise<FollowStatus> {
    const response = await apiClient.get<FollowStatus>(`/api/v2/friendships/${userId}/status`);
    return response.data;
  },
};
