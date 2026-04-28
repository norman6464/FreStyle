import apiClient from '../lib/axios';
import axios from 'axios';
import type { Profile } from '../types';

/**
 * Profile 関連 API の薄いラッパ。
 * フロント `Profile` 型は backend `domain.ProfileView` と 1:1 なのでマッパー不要。
 */

export interface UpdateProfileRequest {
  displayName: string;
  bio: string;
  avatarUrl: string;
  status: string;
}

interface PresignedUrlResponse {
  uploadUrl: string;
  imageUrl: string;
}

const ProfileRepository = {
  async fetchProfile(): Promise<Profile> {
    const res = await apiClient.get<Profile>('/api/v2/profile/me');
    return res.data;
  },

  async updateProfile(data: UpdateProfileRequest): Promise<Profile> {
    const res = await apiClient.put<Profile>('/api/v2/profile/me/update', data);
    return res.data;
  },

  async getImagePresignedUrl(fileName: string, contentType: string): Promise<PresignedUrlResponse> {
    const res = await apiClient.post<PresignedUrlResponse>('/api/v2/profile/me/image/presigned-url', {
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
