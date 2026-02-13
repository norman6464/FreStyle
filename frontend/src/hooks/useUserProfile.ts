import { useState, useCallback } from 'react';
import UserProfileRepository, {
  UserProfile,
  UpdateUserProfileRequest,
} from '../repositories/UserProfileRepository';

/**
 * ユーザープロファイルフック
 *
 * <p>役割:</p>
 * <ul>
 *   <li>ユーザープロファイル管理</li>
 *   <li>UserProfileRepositoryを使用してAPI呼び出し</li>
 * </ul>
 *
 * <p>Hooks層（Presentation Layer - Business Logic）:</p>
 * <ul>
 *   <li>コンポーネントからビジネスロジックを分離</li>
 *   <li>Repository層を使用してAPI呼び出し</li>
 * </ul>
 */
export const useUserProfile = () => {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /**
   * 自分のプロファイルを取得
   */
  const fetchMyProfile = useCallback(async (): Promise<void> => {
    setLoading(true);
    setError(null);

    try {
      const data = await UserProfileRepository.getMyProfile();
      setProfile(data);
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'プロファイルの取得に失敗しました。';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * プロファイルを更新
   */
  const updateProfile = useCallback(
    async (request: UpdateUserProfileRequest): Promise<boolean> => {
      setLoading(true);
      setError(null);

      try {
        const updatedProfile = await UserProfileRepository.updateProfile(request);
        setProfile(updatedProfile);
        return true;
      } catch (err) {
        const errorMessage =
          err instanceof Error ? err.message : 'プロファイルの更新に失敗しました。';
        setError(errorMessage);
        return false;
      } finally {
        setLoading(false);
      }
    },
    []
  );

  return {
    profile,
    loading,
    error,
    fetchMyProfile,
    updateProfile,
  };
};
