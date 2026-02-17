import apiClient from '../lib/axios';
import type { FriendshipUser, FollowStatus } from '../types';

export const FriendshipRepository = {
  async getFollowing(): Promise<FriendshipUser[]> {
    const response = await apiClient.get<FriendshipUser[]>('/api/friendships/following');
    return response.data;
  },

  async getFollowers(): Promise<FriendshipUser[]> {
    const response = await apiClient.get<FriendshipUser[]>('/api/friendships/followers');
    return response.data;
  },

  async follow(userId: number): Promise<FriendshipUser> {
    const response = await apiClient.post<FriendshipUser>(`/api/friendships/${userId}/follow`);
    return response.data;
  },

  async unfollow(userId: number): Promise<void> {
    await apiClient.delete(`/api/friendships/${userId}/follow`);
  },

  async checkStatus(userId: number): Promise<FollowStatus> {
    const response = await apiClient.get<FollowStatus>(`/api/friendships/${userId}/status`);
    return response.data;
  },
};
