import { useEffect, useState } from 'react';
import { ProfileRepository } from '@/entities/user';

/**
 * useCurrentUserId は自分の userId を 1 回引く。
 *
 * 発言の「自分の投稿だけに操作を出す」判定（author.userId との比較）に使う。
 * 引けなかった（未認証・通信失敗）ときは null のまま — 呼び出し側は「自分の発言か
 * 分からない」を「他人の発言」と同じ扱いにする（安全側。誤って他人の発言に
 * 操作を出す方が、自分の発言に操作が出ないより悪い）。
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
