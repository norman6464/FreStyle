import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useTemplates } from '../useTemplates';
import { TemplateRepository } from '../../repositories/TemplateRepository';

vi.mock('../../repositories/TemplateRepository');
const mockedRepo = vi.mocked(TemplateRepository);

describe('useTemplates', () => {
  const mockTemplates = [
    { id: 1, title: '会議の進行', description: 'desc', category: 'meeting', openingMessage: 'msg', difficulty: 'beginner' },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    mockedRepo.fetchTemplates.mockResolvedValue(mockTemplates);
  });

  it('初期ロード時に全テンプレートを取得する', async () => {
    const { result } = renderHook(() => useTemplates());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.templates).toEqual(mockTemplates);
    expect(mockedRepo.fetchTemplates).toHaveBeenCalledWith(undefined);
  });

  it('カテゴリ変更でデータを再取得する', async () => {
    const { result } = renderHook(() => useTemplates());
    await waitFor(() => expect(result.current.loading).toBe(false));
    act(() => result.current.changeCategory('meeting'));
    await waitFor(() => expect(mockedRepo.fetchTemplates).toHaveBeenCalledWith('meeting'));
  });

  it('エラー時にエラーメッセージを設定する', async () => {
    mockedRepo.fetchTemplates.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useTemplates());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe('テンプレートの取得に失敗しました');
  });
});
