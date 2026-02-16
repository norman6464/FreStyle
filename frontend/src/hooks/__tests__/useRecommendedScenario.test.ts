import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useRecommendedScenario } from '../useRecommendedScenario';
import PracticeRepository from '../../repositories/PracticeRepository';

vi.mock('../../repositories/PracticeRepository');

const mockedRepo = vi.mocked(PracticeRepository);

const mockScenarios = [
  { id: 1, name: '本番障害の緊急報告', description: '説明1', category: 'customer', roleName: '顧客', difficulty: 'intermediate', systemPrompt: '' },
  { id: 2, name: '設計レビュー', description: '説明2', category: 'senior', roleName: '上司', difficulty: 'intermediate', systemPrompt: '' },
  { id: 3, name: 'オンボーディング', description: '説明3', category: 'team', roleName: '新人', difficulty: 'beginner', systemPrompt: '' },
  { id: 4, name: '見積もり交渉', description: '説明4', category: 'customer', roleName: '顧客', difficulty: 'advanced', systemPrompt: '' },
];

describe('useRecommendedScenario', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedRepo.getScenarios.mockResolvedValue(mockScenarios);
  });

  it('弱点軸がない場合はnullを返す', async () => {
    const { result } = renderHook(() => useRecommendedScenario(null));

    await waitFor(() => {
      expect(result.current.scenario).toBeNull();
    });
  });

  it('論理的構成力が弱点の場合、customer/seniorのシナリオを推薦する', async () => {
    const weakAxis = { axis: '論理的構成力', score: 5, comment: 'もう少し' };
    const { result } = renderHook(() => useRecommendedScenario(weakAxis));

    await waitFor(() => {
      expect(result.current.scenario).not.toBeNull();
    });
    expect(['customer', 'senior']).toContain(result.current.scenario?.category);
  });

  it('質問・傾聴力が弱点の場合、team/seniorのシナリオを推薦する', async () => {
    const weakAxis = { axis: '質問・傾聴力', score: 4, comment: '要改善' };
    const { result } = renderHook(() => useRecommendedScenario(weakAxis));

    await waitFor(() => {
      expect(result.current.scenario).not.toBeNull();
    });
    expect(['team', 'senior']).toContain(result.current.scenario?.category);
  });

  it('ローディング状態を管理する', async () => {
    const weakAxis = { axis: '提案力', score: 3, comment: '' };
    const { result } = renderHook(() => useRecommendedScenario(weakAxis));

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('シナリオ取得失敗時はnullを返す', async () => {
    mockedRepo.getScenarios.mockRejectedValue(new Error('取得失敗'));
    const weakAxis = { axis: '配慮表現', score: 6, comment: '' };
    const { result } = renderHook(() => useRecommendedScenario(weakAxis));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(result.current.scenario).toBeNull();
  });

  it('マッピングは存在するが一致するシナリオがない場合はnullを返す', async () => {
    mockedRepo.getScenarios.mockResolvedValue([]);
    const weakAxis = { axis: '論理的構成力', score: 5, comment: '' };
    const { result } = renderHook(() => useRecommendedScenario(weakAxis));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(result.current.scenario).toBeNull();
  });

  it('未知の軸名の場合はnullを返す', async () => {
    const weakAxis = { axis: '未知の軸', score: 3, comment: '' };
    const { result } = renderHook(() => useRecommendedScenario(weakAxis));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(result.current.scenario).toBeNull();
  });
});
