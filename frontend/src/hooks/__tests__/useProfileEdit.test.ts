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

  it('loading状態が初期trueからfalseに変化する', async () => {
    const { result } = renderHook(() => useProfileEdit());

    // 初期状態はloading true
    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('updateFieldでbioフィールドも更新できる', async () => {
    const { result } = renderHook(() => useProfileEdit());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    act(() => {
      result.current.updateField('bio', '新しい自己紹介');
    });

    expect(result.current.form.bio).toBe('新しい自己紹介');
    // nameは変更されていないこと
    expect(result.current.form.name).toBe('テスト太郎');
  });

  it('handleUpdate中はsubmittingがtrueになる', async () => {
    let resolveUpdate: (value: any) => void;
    mockUpdateProfile.mockImplementation(
      () => new Promise((resolve) => { resolveUpdate = resolve; })
    );

    const { result } = renderHook(() => useProfileEdit());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.submitting).toBe(false);

    let updatePromise: Promise<void>;
    act(() => {
      updatePromise = result.current.handleUpdate();
    });

    expect(result.current.submitting).toBe(true);

    await act(async () => {
      resolveUpdate!({ success: 'OK' });
      await updatePromise!;
    });

    expect(result.current.submitting).toBe(false);
  });

  it('handleUpdate時にupdateProfileにフォーム値が渡される', async () => {
    const { result } = renderHook(() => useProfileEdit());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    act(() => {
      result.current.updateField('name', '更新太郎');
    });

    await act(async () => {
      await result.current.handleUpdate();
    });

    expect(mockUpdateProfile).toHaveBeenCalledWith({ name: '更新太郎', bio: '自己紹介文' });
  });
});
