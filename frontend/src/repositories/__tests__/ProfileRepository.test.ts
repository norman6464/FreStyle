import { describe, it, expect, vi, beforeEach } from 'vitest';
import ProfileRepository from '../ProfileRepository';
import apiClient from '../../lib/axios';
import axios from 'axios';

vi.mock('../../lib/axios', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
    post: vi.fn(),
  },
}));

vi.mock('axios', () => ({
  default: {
    put: vi.fn(),
  },
}));

const fixtureProfile = {
  userId: 1,
  displayName: 'テスト',
  bio: '自己紹介',
  avatarUrl: '',
  status: '',
  updatedAt: '2026-04-28T00:00:00Z',
};

describe('ProfileRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchProfile: プロフィールを取得できる', async () => {
    vi.mocked(apiClient.get).mockResolvedValue({ data: fixtureProfile });
    const result = await ProfileRepository.fetchProfile();
    expect(apiClient.get).toHaveBeenCalledWith('/api/v2/profile/me');
    expect(result).toEqual(fixtureProfile);
  });

  it('updateProfile: プロフィールを更新できる', async () => {
    const updated = { ...fixtureProfile, bio: '更新' };
    vi.mocked(apiClient.put).mockResolvedValue({ data: updated });

    const result = await ProfileRepository.updateProfile({
      displayName: 'テスト',
      bio: '更新',
      avatarUrl: '',
      status: '',
    });

    expect(apiClient.put).toHaveBeenCalledWith('/api/v2/profile/me/update', {
      displayName: 'テスト',
      bio: '更新',
      avatarUrl: '',
      status: '',
    });
    expect(result).toEqual(updated);
  });

  it('getImagePresignedUrl: Presigned URLを取得できる', async () => {
    const mockData = { uploadUrl: 'https://s3.example.com/upload', imageUrl: 'https://cdn.example.com/image.png' };
    vi.mocked(apiClient.post).mockResolvedValue({ data: mockData });

    const result = await ProfileRepository.getImagePresignedUrl('avatar.png', 'image/png');

    expect(apiClient.post).toHaveBeenCalledWith('/api/v2/profile/me/image/presigned-url', {
      fileName: 'avatar.png',
      contentType: 'image/png',
    });
    expect(result).toEqual(mockData);
  });

  it('uploadToS3: S3にファイルをアップロードできる', async () => {
    vi.mocked(axios.put).mockResolvedValue({});
    const file = new File(['test'], 'avatar.png', { type: 'image/png' });

    await ProfileRepository.uploadToS3('https://s3.example.com/upload', file);

    expect(axios.put).toHaveBeenCalledWith('https://s3.example.com/upload', file, {
      headers: { 'Content-Type': 'image/png' },
    });
  });
});
