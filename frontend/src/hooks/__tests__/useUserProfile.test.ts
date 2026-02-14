import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useUserProfile } from '../useUserProfile';
import UserProfileRepository from '../../repositories/UserProfileRepository';

vi.mock('../../repositories/UserProfileRepository');

const mockedRepo = vi.mocked(UserProfileRepository);

describe('useUserProfile', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchMyProfile: プロファイルを取得する', async () => {
    const mockProfile = { id: 1, userId: 1, displayName: 'テスト' };
    mockedRepo.getMyProfile.mockResolvedValue(mockProfile as any);

    const { result } = renderHook(() => useUserProfile());

    await act(async () => {
      await result.current.fetchMyProfile();
    });

    expect(result.current.profile).toEqual(mockProfile);
    expect(result.current.loading).toBe(false);
  });

  it('fetchMyProfile: エラー時にerrorを設定する', async () => {
    mockedRepo.getMyProfile.mockRejectedValue(new Error('取得失敗'));

    const { result } = renderHook(() => useUserProfile());

    await act(async () => {
      await result.current.fetchMyProfile();
    });

    expect(result.current.error).toBe('取得失敗');
  });

  it('updateProfile: プロファイルを更新する', async () => {
    const mockUpdated = { id: 1, userId: 1, displayName: '更新済み' };
    mockedRepo.updateProfile.mockResolvedValue(mockUpdated as any);

    const { result } = renderHook(() => useUserProfile());

    let success: boolean = false;
    await act(async () => {
      success = await result.current.updateProfile({ displayName: '更新済み' });
    });

    expect(success).toBe(true);
    expect(result.current.profile).toEqual(mockUpdated);
  });

  it('updateProfile: エラー時にfalseを返す', async () => {
    mockedRepo.updateProfile.mockRejectedValue(new Error('更新失敗'));

    const { result } = renderHook(() => useUserProfile());

    let success: boolean = true;
    await act(async () => {
      success = await result.current.updateProfile({ displayName: '更新済み' });
    });

    expect(success).toBe(false);
    expect(result.current.error).toBe('更新失敗');
  });

  it('fetchMyProfile: non-Errorオブジェクトのreject時にデフォルトメッセージを設定する', async () => {
    mockedRepo.getMyProfile.mockRejectedValue('文字列エラー');

    const { result } = renderHook(() => useUserProfile());

    await act(async () => {
      await result.current.fetchMyProfile();
    });

    expect(result.current.error).toBe('プロファイルの取得に失敗しました。');
  });

  it('updateProfile: non-Errorオブジェクトのreject時にデフォルトメッセージを設定する', async () => {
    mockedRepo.updateProfile.mockRejectedValue('文字列エラー');

    const { result } = renderHook(() => useUserProfile());

    let success: boolean = true;
    await act(async () => {
      success = await result.current.updateProfile({ displayName: 'テスト' });
    });

    expect(success).toBe(false);
    expect(result.current.error).toBe('プロファイルの更新に失敗しました。');
  });

  it('初期状態でprofileがnull・loadingがfalse・errorがnull', () => {
    const { result } = renderHook(() => useUserProfile());
    expect(result.current.profile).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it('fetchMyProfile: loading状態が正しく変化する', async () => {
    let resolvePromise: (value: any) => void;
    const pendingPromise = new Promise((resolve) => {
      resolvePromise = resolve;
    });
    mockedRepo.getMyProfile.mockReturnValue(pendingPromise as any);

    const { result } = renderHook(() => useUserProfile());

    expect(result.current.loading).toBe(false);

    let fetchPromise: Promise<void>;
    act(() => {
      fetchPromise = result.current.fetchMyProfile();
    });

    expect(result.current.loading).toBe(true);

    await act(async () => {
      resolvePromise!({ id: 1, displayName: 'テスト' });
      await fetchPromise;
    });

    expect(result.current.loading).toBe(false);
  });
});
