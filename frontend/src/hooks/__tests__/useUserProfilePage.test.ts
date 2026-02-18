import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useUserProfilePage } from '../useUserProfilePage';

const mockFetchMyProfile = vi.fn();
const mockUpdateProfile = vi.fn();
const mockNavigate = vi.fn();

vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock('../useUserProfile', () => ({
  useUserProfile: () => ({
    profile: null,
    loading: false,
    fetchMyProfile: mockFetchMyProfile,
    updateProfile: mockUpdateProfile,
  }),
}));

describe('useUserProfilePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('初期化時にfetchMyProfileが呼ばれる', () => {
    renderHook(() => useUserProfilePage());
    expect(mockFetchMyProfile).toHaveBeenCalled();
  });

  it('formの初期値が空', () => {
    const { result } = renderHook(() => useUserProfilePage());
    expect(result.current.form.displayName).toBe('');
    expect(result.current.form.selfIntroduction).toBe('');
    expect(result.current.form.personalityTraits).toEqual([]);
  });

  it('messageの初期値がnull', () => {
    const { result } = renderHook(() => useUserProfilePage());
    expect(result.current.message).toBeNull();
  });

  it('isNewProfileの初期値がfalse', () => {
    const { result } = renderHook(() => useUserProfilePage());
    expect(result.current.isNewProfile).toBe(false);
  });

  it('loadingの値が返される', () => {
    const { result } = renderHook(() => useUserProfilePage());
    expect(result.current.loading).toBe(false);
  });

  it('setFormでフォーム値を更新できる', () => {
    const { result } = renderHook(() => useUserProfilePage());

    act(() => {
      result.current.setForm(prev => ({ ...prev, displayName: 'テスト' }));
    });

    expect(result.current.form.displayName).toBe('テスト');
  });

  it('togglePersonalityTraitで特性を追加できる', () => {
    const { result } = renderHook(() => useUserProfilePage());

    act(() => {
      result.current.togglePersonalityTrait('論理的');
    });

    expect(result.current.form.personalityTraits).toContain('論理的');
  });

  it('togglePersonalityTraitで特性を削除できる', () => {
    const { result } = renderHook(() => useUserProfilePage());

    act(() => {
      result.current.togglePersonalityTrait('論理的');
    });

    act(() => {
      result.current.togglePersonalityTrait('論理的');
    });

    expect(result.current.form.personalityTraits).not.toContain('論理的');
  });

  it('handleSave成功時にsuccessメッセージを設定する', async () => {
    mockUpdateProfile.mockResolvedValue(true);
    const { result } = renderHook(() => useUserProfilePage());

    await act(async () => {
      await result.current.handleSave({ preventDefault: vi.fn() } as any);
    });

    expect(result.current.message?.type).toBe('success');
  });

  it('handleSave失敗時にerrorメッセージを設定する', async () => {
    mockUpdateProfile.mockResolvedValue(false);
    const { result } = renderHook(() => useUserProfilePage());

    await act(async () => {
      await result.current.handleSave({ preventDefault: vi.fn() } as any);
    });

    expect(result.current.message?.type).toBe('error');
  });
});
