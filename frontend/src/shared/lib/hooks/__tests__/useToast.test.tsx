import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ToastProvider } from '@/app/providers/ToastProvider';
import { useToast } from '../useToast';
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

  it('同一メッセージを連続で出すと 1 枚にまとまり count が増える', () => {
    const { result } = renderHook(() => useToast(), { wrapper });
    act(() => {
      result.current.showToast('success', 'ノートを作成しました');
      result.current.showToast('success', 'ノートを作成しました');
      result.current.showToast('success', 'ノートを作成しました');
    });
    expect(result.current.toasts).toHaveLength(1);
    expect(result.current.toasts[0].count).toBe(3);
  });

  it('type が違えば同じ文言でも別枠になる', () => {
    const { result } = renderHook(() => useToast(), { wrapper });
    act(() => {
      result.current.showToast('success', '完了');
      result.current.showToast('error', '完了');
    });
    expect(result.current.toasts).toHaveLength(2);
  });

  it('総数は上限(3)でクランプされ古いものから落ちる', () => {
    const { result } = renderHook(() => useToast(), { wrapper });
    act(() => {
      result.current.showToast('info', 'A');
      result.current.showToast('info', 'B');
      result.current.showToast('info', 'C');
      result.current.showToast('info', 'D');
    });
    expect(result.current.toasts).toHaveLength(3);
    expect(result.current.toasts.map((t) => t.message)).toEqual(['B', 'C', 'D']);
  });
});
