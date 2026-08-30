import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useWorkspaceLearningSummary } from '../useWorkspaceLearningSummary';
import AdminMemberRepository, {
  type WorkspaceLearningSummary,
} from '@/entities/member/api/adminMemberRepository';

vi.mock('@/entities/member/api/adminMemberRepository', () => ({
  default: {
    learningSummary: vi.fn(),
  },
}));

const mockLearningSummary = vi.mocked(AdminMemberRepository.learningSummary);

const sample: WorkspaceLearningSummary = {
  traineeCount: 3,
  activeToday: 1,
  activeThisWeek: 2,
  recentMembers: [
    { userId: 11, name: 'member-a', lastActiveDate: '2026-07-09', recentActivityCount: 2 },
  ],
};

describe('useWorkspaceLearningSummary', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('サマリーを取得して返す', async () => {
    mockLearningSummary.mockResolvedValue(sample);
    const { result } = renderHook(() => useWorkspaceLearningSummary());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.summary?.traineeCount).toBe(3);
    expect(result.current.error).toBeNull();
  });

  it('enabled=false のときはリクエストを発行しない', async () => {
    const { result } = renderHook(() => useWorkspaceLearningSummary({ enabled: false }));
    await new Promise((r) => setTimeout(r, 10));
    expect(mockLearningSummary).not.toHaveBeenCalled();
    expect(result.current.summary).toBeNull();
  });

  it('取得失敗時は error をセットする', async () => {
    mockLearningSummary.mockRejectedValue(new Error('network'));
    const { result } = renderHook(() => useWorkspaceLearningSummary());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.summary).toBeNull();
    expect(result.current.error).toBe('学習状況の取得に失敗しました');
  });
});
