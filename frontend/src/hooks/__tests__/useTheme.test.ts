import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useTheme } from '../useTheme';

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

describe('useTheme', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal('localStorage', createMockStorage());
    document.documentElement.classList.remove('light');
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('初期テーマはdark', () => {
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('dark');
  });

  it('toggleThemeでlight/darkを切り替えられる', () => {
    const { result } = renderHook(() => useTheme());

    act(() => {
      result.current.toggleTheme();
    });
    expect(result.current.theme).toBe('light');

    act(() => {
      result.current.toggleTheme();
    });
    expect(result.current.theme).toBe('dark');
  });

  it('lightテーマの場合htmlにlightクラスが追加される', () => {
    const { result } = renderHook(() => useTheme());

    act(() => {
      result.current.toggleTheme();
    });

    expect(document.documentElement.classList.contains('light')).toBe(true);
  });

  it('darkテーマの場合htmlからlightクラスが削除される', () => {
    document.documentElement.classList.add('light');
    const { result } = renderHook(() => useTheme());

    act(() => {
      result.current.toggleTheme(); // dark → light
    });
    act(() => {
      result.current.toggleTheme(); // light → dark
    });

    expect(document.documentElement.classList.contains('light')).toBe(false);
  });

  it('localStorageにテーマを保存する', () => {
    const { result } = renderHook(() => useTheme());

    act(() => {
      result.current.toggleTheme();
    });

    expect(localStorage.setItem).toHaveBeenCalledWith('theme', 'light');
  });

  it('localStorageからテーマを復元する', () => {
    (localStorage.getItem as ReturnType<typeof vi.fn>).mockReturnValue('light');
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('light');
  });

  it('無効なlocalStorage値の場合はdarkを使用する', () => {
    (localStorage.getItem as ReturnType<typeof vi.fn>).mockReturnValue('invalid');
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('dark');
  });

  it('localStorageが空の場合はdarkを使用する', () => {
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe('dark');
  });
});
