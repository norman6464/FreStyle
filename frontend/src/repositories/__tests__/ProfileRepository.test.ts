import { describe, it, expect, vi, beforeEach } from 'vitest';
import ProfileRepository from '../ProfileRepository';
import apiClient from '../../lib/axios';

vi.mock('../../lib/axios');

const mockedApiClient = vi.mocked(apiClient);

describe('ProfileRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchProfile: プロフィールを取得できる', async () => {
    const mockData = { name: 'テスト', bio: '自己紹介' };
    mockedApiClient.get.mockResolvedValue({ data: mockData });

    const result = await ProfileRepository.fetchProfile();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/profile/me');
    expect(result).toEqual(mockData);
  });

  it('updateProfile: プロフィールを更新できる', async () => {
    const mockData = { success: 'プロフィールを更新しました。' };
    mockedApiClient.put.mockResolvedValue({ data: mockData });

    const result = await ProfileRepository.updateProfile({ name: 'テスト', bio: '更新' });

    expect(mockedApiClient.put).toHaveBeenCalledWith('/api/profile/me/update', { name: 'テスト', bio: '更新' });
    expect(result).toEqual(mockData);
  });
});
