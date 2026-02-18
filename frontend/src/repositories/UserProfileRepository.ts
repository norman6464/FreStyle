import apiClient from '../lib/axios';

/**
 * ユーザープロファイルリポジトリ
 *
 * <p>役割:</p>
 * <ul>
 *   <li>ユーザープロファイル関連のAPI呼び出しを抽象化</li>
 * </ul>
 *
 * <p>インフラ層（Infrastructure Layer）:</p>
 * <ul>
 *   <li>外部APIとの通信を担当</li>
 * </ul>
 */

export interface UserProfile {
  id: number;
  userId: number;
  displayName: string;
  selfIntroduction?: string;
  communicationStyle?: string;
  personalityTraits?: string[];
  goals?: string;
  concerns?: string;
  preferredFeedbackStyle?: string;
}

export interface UpdateUserProfileRequest {
  displayName: string;
  selfIntroduction?: string;
  communicationStyle?: string;
  personalityTraits?: string[];
  goals?: string;
  concerns?: string;
  preferredFeedbackStyle?: string;
}

class UserProfileRepository {
  /**
   * 自分のプロファイルを取得
   */
  async getMyProfile(): Promise<UserProfile> {
    const response = await apiClient.get('/api/user-profile/me');
    return response.data;
  }

  /**
   * プロファイルを更新
   */
  async updateProfile(request: UpdateUserProfileRequest): Promise<UserProfile> {
    const response = await apiClient.post('/api/user-profile/me/upsert', request);
    return response.data;
  }
}

export default new UserProfileRepository();
