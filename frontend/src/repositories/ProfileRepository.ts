import apiClient from '../lib/axios';

/**
 * プロフィールリポジトリ
 *
 * プロフィール関連のAPI呼び出しを抽象化し、
 * axiosインターセプターによる自動トークンリフレッシュを活用する。
 */

interface ProfileData {
  name: string;
  bio: string;
}

const ProfileRepository = {
  async fetchProfile(): Promise<ProfileData> {
    const res = await apiClient.get('/api/profile/me');
    return res.data;
  },

  async updateProfile(data: ProfileData): Promise<{ success: string }> {
    const res = await apiClient.put('/api/profile/me/update', data);
    return res.data;
  },
};

export default ProfileRepository;
