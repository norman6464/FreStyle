import { useEffect, useState } from 'react';
import { ProfileRepository } from '@/entities/user';

/**
 * useCurrentUserId は自分の userId を 1 回引く。
 *
 * backlog/model/useCurrentUserId.ts と同じ形（「自分自身には操作を出さない」判定に使う）。
 * 単一画面専用の小さい hook なので shared へは上げず、ここにも同じものを持つ
 * （entities/user にはまだ current user id を持つ状態が無い — Redux の auth スライスは
 * isAuthenticated/loading しか持たない）。
 */
export function useCurrentUserId(): number | null {
  const [userId, setUserId] = useState<number | null>(null);

  useEffect(() => {
    let active = true;
    ProfileRepository.fetchProfile()
      .then((profile) => {
        if (active) setUserId(profile.userId);
      })
      .catch(() => {
        if (active) setUserId(null);
      });
    return () => {
      active = false;
    };
  }, []);

  return userId;
}
