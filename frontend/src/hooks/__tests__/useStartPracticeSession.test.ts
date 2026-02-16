import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useStartPracticeSession } from '../useStartPracticeSession';
import PracticeRepository from '../../repositories/PracticeRepository';

vi.mock('../../repositories/PracticeRepository');

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

const mockedRepo = vi.mocked(PracticeRepository);

describe('useStartPracticeSession', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('セッション作成成功後にAIチャットページに遷移する', async () => {
    mockedRepo.createPracticeSession.mockResolvedValue({ id: 42 });

    const { result } = renderHook(() => useStartPracticeSession());

    await act(async () => {
      await result.current.startSession({
        id: 1,
        name: 'テストシナリオ',
        description: '説明',
        category: 'customer',
        roleName: '顧客',
        difficulty: 'intermediate',
        systemPrompt: '',
      });
    });

    expect(mockedRepo.createPracticeSession).toHaveBeenCalledWith({ scenarioId: 1 });
    expect(mockNavigate).toHaveBeenCalledWith('/chat/ask-ai/42', {
      state: {
        sessionType: 'practice',
        scenarioId: 1,
        scenarioName: 'テストシナリオ',
        initialPrompt: '練習開始',
      },
    });
  });

  it('セッション作成失敗時は練習一覧に遷移する', async () => {
    mockedRepo.createPracticeSession.mockRejectedValue(new Error('失敗'));

    const { result } = renderHook(() => useStartPracticeSession());

    await act(async () => {
      await result.current.startSession({
        id: 1,
        name: 'テスト',
        description: '',
        category: 'customer',
        roleName: '顧客',
        difficulty: 'beginner',
        systemPrompt: '',
      });
    });

    expect(mockNavigate).toHaveBeenCalledWith('/practice');
  });

  it('開始中はstarting状態がtrueになる', async () => {
    let resolvePromise: (value: { id: number }) => void;
    mockedRepo.createPracticeSession.mockReturnValue(
      new Promise((resolve) => { resolvePromise = resolve; })
    );

    const { result } = renderHook(() => useStartPracticeSession());

    expect(result.current.starting).toBe(false);

    let startPromise: Promise<void>;
    act(() => {
      startPromise = result.current.startSession({
        id: 1, name: 'テスト', description: '', category: 'customer',
        roleName: '顧客', difficulty: 'beginner', systemPrompt: '',
      });
    });

    expect(result.current.starting).toBe(true);

    await act(async () => {
      resolvePromise!({ id: 1 });
      await startPromise!;
    });

    expect(result.current.starting).toBe(false);
  });
});
