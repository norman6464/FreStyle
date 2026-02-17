import { useState, useEffect, useCallback } from 'react';
import { FriendshipRepository } from '../repositories/FriendshipRepository';
import type { FriendshipUser } from '../types';

export function useFriendship() {
  const [following, setFollowing] = useState<FriendshipUser[]>([]);
  const [followers, setFollowers] = useState<FriendshipUser[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    try {
      const [followingData, followersData] = await Promise.all([
        FriendshipRepository.getFollowing(),
        FriendshipRepository.getFollowers(),
      ]);
      setFollowing(followingData);
      setFollowers(followersData);
    } catch {
      setFollowing([]);
      setFollowers([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const followUser = useCallback(async (userId: number) => {
    try {
      await FriendshipRepository.follow(userId);
      await fetchData();
    } catch {
      // エラーハンドリング
    }
  }, [fetchData]);

  const unfollowUser = useCallback(async (userId: number) => {
    try {
      await FriendshipRepository.unfollow(userId);
      await fetchData();
    } catch {
      // エラーハンドリング
    }
  }, [fetchData]);

  return { following, followers, loading, followUser, unfollowUser, refetch: fetchData };
}
