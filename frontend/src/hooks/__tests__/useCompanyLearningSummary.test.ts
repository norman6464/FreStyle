import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useCompanyLearningSummary } from '../useCompanyLearningSummary';
import AdminMemberRepository, {
  type CompanyLearningSummary,
} from '../../repositories/AdminMemberRepository';

vi.mock('../../repositories/AdminMemberRepository', () => ({
  default: {
    learningSummary: vi.fn(),
  },
}));

const mockLearningSummary = vi.mocked(AdminMemberRepository.learningSummary);

const sample: CompanyLearningSummary = {
  traineeCount: 3,
  activeToday: 1,
  activeThisWeek: 2,
  recentMembers: [
    { userId: 11, name: 'member-a', lastActiveDate: '2026-07-09', recentActivityCount: 2 },
  ],
};

describe('useCompanyLearningSummary', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('サマリーを取得して返す', async () => {
    mockLearningSummary.mockResolvedValue(sample);
    const { result } = renderHook(() => useCompanyLearningSummary());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.summary?.traineeCount).toBe(3);
    expect(result.current.error).toBeNull();
  });

  it('enabled=false のときはリクエストを発行しない', async () => {
    const { result } = renderHook(() => useCompanyLearningSummary({ enabled: false }));
    await new Promise((r) => setTimeout(r, 10));
    expect(mockLearningSummary).not.toHaveBeenCalled();
    expect(result.current.summary).toBeNull();
  });

  it('取得失敗時は error をセットする', async () => {
    mockLearningSummary.mockRejectedValue(new Error('network'));
    const { result } = renderHook(() => useCompanyLearningSummary());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.summary).toBeNull();
    expect(result.current.error).toBe('学習状況の取得に失敗しました');
  });
});
