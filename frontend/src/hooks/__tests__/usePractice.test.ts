import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { usePractice } from '../usePractice';
import PracticeRepository from '../../repositories/PracticeRepository';

vi.mock('../../repositories/PracticeRepository');

const mockedRepo = vi.mocked(PracticeRepository);

describe('usePractice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchScenarios: シナリオ一覧を取得する', async () => {
    const mockScenarios = [{ id: 1, name: 'テストシナリオ', description: '説明', category: 'customer', roleName: '顧客', difficulty: 'easy', systemPrompt: 'prompt' }];
    mockedRepo.getScenarios.mockResolvedValue(mockScenarios as any);

    const { result } = renderHook(() => usePractice());

    await act(async () => {
      await result.current.fetchScenarios();
    });

    expect(result.current.scenarios).toEqual(mockScenarios);
    expect(result.current.loading).toBe(false);
  });

  it('fetchScenarios: エラー時にerrorを設定する', async () => {
    mockedRepo.getScenarios.mockRejectedValue(new Error('取得失敗'));

    const { result } = renderHook(() => usePractice());

    await act(async () => {
      await result.current.fetchScenarios();
    });

    expect(result.current.error).toBe('取得失敗');
  });

  it('fetchScenario: シナリオ詳細を取得する', async () => {
    const mockScenario = { id: 1, name: 'テスト', description: '説明', category: 'customer', roleName: '顧客', difficulty: 'easy', systemPrompt: 'prompt' };
    mockedRepo.getScenario.mockResolvedValue(mockScenario as any);

    const { result } = renderHook(() => usePractice());

    let fetched: any;
    await act(async () => {
      fetched = await result.current.fetchScenario(1);
    });

    expect(fetched).toEqual(mockScenario);
    expect(result.current.currentScenario).toEqual(mockScenario);
  });

  it('createPracticeSession: 練習セッションを作成する', async () => {
    const mockSession = { id: 1, scenarioId: 1 };
    mockedRepo.createPracticeSession.mockResolvedValue(mockSession);

    const { result } = renderHook(() => usePractice());

    let created: any;
    await act(async () => {
      created = await result.current.createPracticeSession({ scenarioId: 1 });
    });

    expect(created).toEqual(mockSession);
  });

  it('createPracticeSession: エラー時にnullを返す', async () => {
    mockedRepo.createPracticeSession.mockRejectedValue(new Error('作成失敗'));

    const { result } = renderHook(() => usePractice());

    let created: any;
    await act(async () => {
      created = await result.current.createPracticeSession({ scenarioId: 1 });
    });

    expect(created).toBeNull();
    expect(result.current.error).toBe('作成失敗');
  });
});
