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

describe('ProfileRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchProfile: プロフィールを取得できる', async () => {
    const mockData = { name: 'テスト', bio: '自己紹介' };
    vi.mocked(apiClient.get).mockResolvedValue({ data: mockData });

    const result = await ProfileRepository.fetchProfile();

    expect(apiClient.get).toHaveBeenCalledWith('/api/profile/me');
    expect(result).toEqual(mockData);
  });

  it('updateProfile: プロフィールを更新できる', async () => {
    const mockData = { success: 'プロフィールを更新しました。' };
    vi.mocked(apiClient.put).mockResolvedValue({ data: mockData });

    const result = await ProfileRepository.updateProfile({ name: 'テスト', bio: '更新' });

    expect(apiClient.put).toHaveBeenCalledWith('/api/profile/me/update', { name: 'テスト', bio: '更新' });
    expect(result).toEqual(mockData);
  });

  it('getImagePresignedUrl: Presigned URLを取得できる', async () => {
    const mockData = { uploadUrl: 'https://s3.example.com/upload', imageUrl: 'https://cdn.example.com/image.png' };
    vi.mocked(apiClient.post).mockResolvedValue({ data: mockData });

    const result = await ProfileRepository.getImagePresignedUrl('avatar.png', 'image/png');

    expect(apiClient.post).toHaveBeenCalledWith('/api/profile/me/image/presigned-url', {
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
