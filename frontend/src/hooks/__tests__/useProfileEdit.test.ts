import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useProfileEdit } from '../useProfileEdit';

const mockFetchProfile = vi.fn();
const mockUpdateProfile = vi.fn();

vi.mock('../../repositories/ProfileRepository', () => ({
  default: {
    fetchProfile: (...args: unknown[]) => mockFetchProfile(...args),
    updateProfile: (...args: unknown[]) => mockUpdateProfile(...args),
  },
}));

describe('useProfileEdit', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockFetchProfile.mockResolvedValue({ name: 'テスト太郎', bio: '自己紹介文' });
    mockUpdateProfile.mockResolvedValue({ success: '更新しました' });
  });

  it('プロフィール取得成功時にフォームに値がセットされる', async () => {
    const { result } = renderHook(() => useProfileEdit());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.form.name).toBe('テスト太郎');
    expect(result.current.form.bio).toBe('自己紹介文');
  });

  it('プロフィール取得失敗時にエラーメッセージが表示される', async () => {
    mockFetchProfile.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useProfileEdit());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('プロフィール取得に失敗しました。');
  });

  it('updateFieldでフォームの値が更新される', async () => {
    const { result } = renderHook(() => useProfileEdit());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    act(() => {
      result.current.updateField('name', '新しい名前');
    });

    expect(result.current.form.name).toBe('新しい名前');
  });

  it('handleUpdate成功時に成功メッセージが表示される', async () => {
    const { result } = renderHook(() => useProfileEdit());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.handleUpdate();
    });

    expect(result.current.message?.type).toBe('success');
    expect(result.current.message?.text).toBe('更新しました');
  });

  it('handleUpdate失敗時にエラーメッセージが表示される', async () => {
    mockUpdateProfile.mockRejectedValue(new Error('Server Error'));

    const { result } = renderHook(() => useProfileEdit());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.handleUpdate();
    });

    expect(result.current.message?.type).toBe('error');
    expect(result.current.message?.text).toBe('通信エラーが発生しました。');
  });
});
