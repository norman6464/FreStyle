import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useScoreGoal } from '../useScoreGoal';

const mockGet = vi.fn();
const mockPut = vi.fn();

vi.mock('../../lib/axios', () => ({
  default: {
    get: (...args: unknown[]) => mockGet(...args),
    put: (...args: unknown[]) => mockPut(...args),
  },
}));

function createMockStorage(): Storage {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value; }),
    removeItem: vi.fn((key: string) => { delete store[key]; }),
    clear: vi.fn(() => { store = {}; }),
    get length() { return Object.keys(store).length; },
    key: vi.fn((index: number) => Object.keys(store)[index] ?? null),
  };
}

describe('useScoreGoal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('APIから目標スコアを取得する', async () => {
    mockGet.mockResolvedValue({ data: { goalScore: 9.0 } });

    const { result } = renderHook(() => useScoreGoal());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(result.current.goal).toBe(9.0);
    expect(mockGet).toHaveBeenCalledWith('/api/score-goal');
  });

  it('APIエラー時はlocalStorageのデフォルト値を使用する', async () => {
    mockGet.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useScoreGoal());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(result.current.goal).toBe(8.0);
  });

  it('目標スコアを保存できる', async () => {
    mockGet.mockResolvedValue({ data: { goalScore: 8.0 } });
    mockPut.mockResolvedValue({});

    const { result } = renderHook(() => useScoreGoal());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.saveGoal(9.5);
    });

    expect(result.current.goal).toBe(9.5);
    expect(mockPut).toHaveBeenCalledWith('/api/score-goal', { goalScore: 9.5 });
    expect(localStorage.setItem).toHaveBeenCalledWith('scoreGoal', '9.5');
  });

  it('API保存失敗時もlocalStorageには保存される', async () => {
    mockGet.mockResolvedValue({ data: { goalScore: 8.0 } });
    mockPut.mockRejectedValue(new Error('Network Error'));

    const { result } = renderHook(() => useScoreGoal());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      await result.current.saveGoal(7.5);
    });

    expect(result.current.goal).toBe(7.5);
    expect(localStorage.setItem).toHaveBeenCalledWith('scoreGoal', '7.5');
  });

  it('読み込み中はloadingがtrueになる', async () => {
    mockGet.mockResolvedValue({ data: { goalScore: 8.5 } });

    const { result } = renderHook(() => useScoreGoal());

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });
});
