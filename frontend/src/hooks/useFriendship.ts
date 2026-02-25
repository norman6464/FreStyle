import { useState, useEffect, useCallback } from 'react';
import { FriendshipRepository } from '../repositories/FriendshipRepository';
import type { FriendshipUser } from '../types';

export function useFriendship() {
  const [following, setFollowing] = useState<FriendshipUser[]>([]);
  const [followers, setFollowers] = useState<FriendshipUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    setError(null);
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
      setError('フレンド情報の取得に失敗しました');
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
      setError('フォローに失敗しました');
    }
  }, [fetchData]);

  const unfollowUser = useCallback(async (userId: number) => {
    try {
      await FriendshipRepository.unfollow(userId);
      await fetchData();
    } catch {
      setError('フォロー解除に失敗しました');
    }
  }, [fetchData]);

  return { following, followers, loading, error, followUser, unfollowUser, refetch: fetchData };
}
