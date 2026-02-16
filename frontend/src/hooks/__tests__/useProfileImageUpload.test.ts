import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useProfileImageUpload } from '../useProfileImageUpload';
import ProfileRepository from '../../repositories/ProfileRepository';

vi.mock('../../repositories/ProfileRepository', () => ({
  default: {
    getImagePresignedUrl: vi.fn(),
    uploadToS3: vi.fn(),
  },
}));

describe('useProfileImageUpload', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('画像ファイルをアップロードしてimageUrlを返す', async () => {
    vi.mocked(ProfileRepository.getImagePresignedUrl).mockResolvedValue({
      uploadUrl: 'https://s3.example.com/upload',
      imageUrl: 'https://cdn.example.com/profiles/1/avatar.png',
    });
    vi.mocked(ProfileRepository.uploadToS3).mockResolvedValue(undefined);

    const { result } = renderHook(() => useProfileImageUpload());

    const file = new File(['test'], 'avatar.png', { type: 'image/png' });
    let imageUrl: string | null = null;
    await act(async () => {
      imageUrl = await result.current.upload(file);
    });

    expect(ProfileRepository.getImagePresignedUrl).toHaveBeenCalledWith('avatar.png', 'image/png');
    expect(ProfileRepository.uploadToS3).toHaveBeenCalledWith('https://s3.example.com/upload', file);
    expect(imageUrl).toBe('https://cdn.example.com/profiles/1/avatar.png');
  });

  it('アップロード中はuploadingがtrueになる', async () => {
    vi.mocked(ProfileRepository.getImagePresignedUrl).mockResolvedValue({
      uploadUrl: 'https://s3.example.com/upload',
      imageUrl: 'https://cdn.example.com/profiles/1/avatar.png',
    });
    vi.mocked(ProfileRepository.uploadToS3).mockResolvedValue(undefined);

    const { result } = renderHook(() => useProfileImageUpload());

    expect(result.current.uploading).toBe(false);

    const file = new File(['test'], 'avatar.png', { type: 'image/png' });
    await act(async () => {
      await result.current.upload(file);
    });

    expect(result.current.uploading).toBe(false);
  });

  it('許可されていないファイル形式はnullを返す', async () => {
    const { result } = renderHook(() => useProfileImageUpload());

    const file = new File(['test'], 'file.pdf', { type: 'application/pdf' });
    let imageUrl: string | null = null;
    await act(async () => {
      imageUrl = await result.current.upload(file);
    });

    expect(ProfileRepository.getImagePresignedUrl).not.toHaveBeenCalled();
    expect(imageUrl).toBeNull();
  });

  it('エラー時にnullを返す', async () => {
    vi.mocked(ProfileRepository.getImagePresignedUrl).mockRejectedValue(new Error('network error'));
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    const { result } = renderHook(() => useProfileImageUpload());

    const file = new File(['test'], 'avatar.png', { type: 'image/png' });
    let imageUrl: string | null = null;
    await act(async () => {
      imageUrl = await result.current.upload(file);
    });

    expect(imageUrl).toBeNull();
    expect(consoleSpy).toHaveBeenCalled();
    consoleSpy.mockRestore();
  });
});
