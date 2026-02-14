import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ToastProvider, useToast } from '../useToast';
import { ReactNode } from 'react';

const wrapper = ({ children }: { children: ReactNode }) => (
  <ToastProvider>{children}</ToastProvider>
);

describe('useToast', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('初期状態でトーストが空である', () => {
    const { result } = renderHook(() => useToast(), { wrapper });
    expect(result.current.toasts).toHaveLength(0);
  });

  it('showToastでトーストが追加される', () => {
    const { result } = renderHook(() => useToast(), { wrapper });
    act(() => {
      result.current.showToast('success', 'テストメッセージ');
    });
    expect(result.current.toasts).toHaveLength(1);
    expect(result.current.toasts[0].type).toBe('success');
    expect(result.current.toasts[0].message).toBe('テストメッセージ');
  });

  it('removeToastでトーストが削除される', () => {
    const { result } = renderHook(() => useToast(), { wrapper });
    act(() => {
      result.current.showToast('success', 'テスト');
    });
    const id = result.current.toasts[0].id;
    act(() => {
      result.current.removeToast(id);
    });
    expect(result.current.toasts).toHaveLength(0);
  });

  it('複数のトーストを追加できる', () => {
    const { result } = renderHook(() => useToast(), { wrapper });
    act(() => {
      result.current.showToast('success', 'メッセージ1');
      result.current.showToast('error', 'メッセージ2');
    });
    expect(result.current.toasts).toHaveLength(2);
  });
});
