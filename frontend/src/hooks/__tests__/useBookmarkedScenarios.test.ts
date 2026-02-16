import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useBookmarkedScenarios } from '../useBookmarkedScenarios';

vi.mock('../../repositories/BookmarkRepository', () => ({
  BookmarkRepository: {
    getAll: vi.fn(),
  },
}));

vi.mock('../../repositories/PracticeRepository', () => ({
  default: {
    getScenarios: vi.fn(),
  },
}));

import { BookmarkRepository } from '../../repositories/BookmarkRepository';
import PracticeRepository from '../../repositories/PracticeRepository';

const mockScenarios = [
  { id: 1, name: 'クレーム対応', description: '説明1', category: 'ビジネス', roleName: '顧客', difficulty: '中級', systemPrompt: '' },
  { id: 2, name: '会議ファシリテーション', description: '説明2', category: 'ビジネス', roleName: '同僚', difficulty: '上級', systemPrompt: '' },
  { id: 3, name: '面接練習', description: '説明3', category: '就活', roleName: '面接官', difficulty: '初級', systemPrompt: '' },
];

describe('useBookmarkedScenarios', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('ブックマーク済みシナリオを取得する', async () => {
    vi.mocked(BookmarkRepository.getAll).mockResolvedValue([1, 3]);
    vi.mocked(PracticeRepository.getScenarios).mockResolvedValue(mockScenarios);

    const { result } = renderHook(() => useBookmarkedScenarios());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.scenarios).toHaveLength(2);
    expect(result.current.scenarios[0].name).toBe('クレーム対応');
    expect(result.current.scenarios[1].name).toBe('面接練習');
  });

  it('ブックマークがない場合は空配列を返す', async () => {
    vi.mocked(BookmarkRepository.getAll).mockResolvedValue([]);
    vi.mocked(PracticeRepository.getScenarios).mockResolvedValue(mockScenarios);

    const { result } = renderHook(() => useBookmarkedScenarios());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.scenarios).toHaveLength(0);
  });

  it('最大件数を指定できる', async () => {
    vi.mocked(BookmarkRepository.getAll).mockResolvedValue([1, 2, 3]);
    vi.mocked(PracticeRepository.getScenarios).mockResolvedValue(mockScenarios);

    const { result } = renderHook(() => useBookmarkedScenarios(2));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.scenarios).toHaveLength(2);
  });

  it('ローディング中はtrueを返す', () => {
    vi.mocked(BookmarkRepository.getAll).mockReturnValue(new Promise(() => {}));
    vi.mocked(PracticeRepository.getScenarios).mockReturnValue(new Promise(() => {}));

    const { result } = renderHook(() => useBookmarkedScenarios());

    expect(result.current.loading).toBe(true);
    expect(result.current.scenarios).toHaveLength(0);
  });
});
