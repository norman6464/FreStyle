import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useLoginPage } from '../useLoginPage';

const mockLocationState = { message: '' };

vi.mock('react-router-dom', () => ({
  useLocation: () => ({ state: mockLocationState }),
}));

describe('useLoginPage', () => {
  beforeEach(() => {
    mockLocationState.message = '';
  });

  it('locationState のメッセージを取得する', () => {
    mockLocationState.message = '確認メールを送信しました';

    const { result } = renderHook(() => useLoginPage());

    expect(result.current.flashMessage).toBe('確認メールを送信しました');
  });

  it('メッセージが無い場合は空文字を返す', () => {
    const { result } = renderHook(() => useLoginPage());

    expect(result.current.flashMessage).toBe('');
  });
});
