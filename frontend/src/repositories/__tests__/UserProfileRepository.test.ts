import { describe, it, expect, vi, beforeEach } from 'vitest';
import userProfileRepository from '../UserProfileRepository';
import apiClient from '../../lib/axios';

vi.mock('../../lib/axios');

const mockedApiClient = vi.mocked(apiClient);

describe('UserProfileRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('getMyProfile: 自分のプロファイルを取得できる', async () => {
    const mockProfile = {
      id: 1,
      userId: 1,
      displayName: 'テストユーザー',
      selfIntroduction: '自己紹介',
      communicationStyle: 'assertive',
    };
    mockedApiClient.get.mockResolvedValue({ data: mockProfile });

    const result = await userProfileRepository.getMyProfile();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/user-profile/me');
    expect(result).toEqual(mockProfile);
  });

  it('updateProfile: プロファイルを更新できる', async () => {
    const updateRequest = {
      displayName: '更新ユーザー',
      selfIntroduction: '更新された自己紹介',
    };
    const mockProfile = { id: 1, userId: 1, ...updateRequest };
    mockedApiClient.post.mockResolvedValue({ data: mockProfile });

    const result = await userProfileRepository.updateProfile(updateRequest);

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/user-profile/me', updateRequest);
    expect(result).toEqual(mockProfile);
  });
});
