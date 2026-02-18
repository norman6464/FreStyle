import apiClient from '../lib/axios';
import axios from 'axios';

/**
 * プロフィールリポジトリ
 *
 * プロフィール関連のAPI呼び出しを抽象化し、
 * axiosインターセプターによる自動トークンリフレッシュを活用する。
 */

export interface ProfileData {
  name: string;
  bio: string;
  iconUrl?: string;
  status?: string;
}

interface PresignedUrlResponse {
  uploadUrl: string;
  imageUrl: string;
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

  async getImagePresignedUrl(fileName: string, contentType: string): Promise<PresignedUrlResponse> {
    const res = await apiClient.post('/api/profile/me/image/presigned-url', {
      fileName,
      contentType,
    });
    return res.data;
  },

  async uploadToS3(uploadUrl: string, file: File): Promise<void> {
    await axios.put(uploadUrl, file, {
      headers: { 'Content-Type': file.type },
    });
  },
};

export default ProfileRepository;
