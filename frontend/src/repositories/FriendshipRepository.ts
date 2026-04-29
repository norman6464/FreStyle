import apiClient from '../lib/axios';
import { FRIENDSHIPS } from '../constants/apiRoutes';
import type { FriendshipUser, FollowStatus } from '../types';

export const FriendshipRepository = {
  async getFollowing(): Promise<FriendshipUser[]> {
    const response = await apiClient.get<FriendshipUser[]>(FRIENDSHIPS.following);
    return response.data;
  },

  async getFollowers(): Promise<FriendshipUser[]> {
    const response = await apiClient.get<FriendshipUser[]>(FRIENDSHIPS.followers);
    return response.data;
  },

  async follow(userId: number): Promise<FriendshipUser> {
    const response = await apiClient.post<FriendshipUser>(FRIENDSHIPS.follow(userId));
    return response.data;
  },

  async unfollow(userId: number): Promise<void> {
    await apiClient.delete(FRIENDSHIPS.follow(userId));
  },

  async checkStatus(userId: number): Promise<FollowStatus> {
    const response = await apiClient.get<FollowStatus>(FRIENDSHIPS.status(userId));
    return response.data;
  },
};
