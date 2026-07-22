import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useLocalStorage } from '../useLocalStorage';

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

describe('useLocalStorage', () => {
  let mockStorage: Storage;

  beforeEach(() => {
    mockStorage = createMockStorage();
    vi.stubGlobal('localStorage', mockStorage);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('デフォルト値が返される', () => {
    const { result } = renderHook(() => useLocalStorage('test-key', 'default'));
    expect(result.current[0]).toBe('default');
  });

  it('localStorageに保存された値が読み込まれる', () => {
    mockStorage.setItem('test-key', JSON.stringify('stored-value'));
    const { result } = renderHook(() => useLocalStorage('test-key', 'default'));
    expect(result.current[0]).toBe('stored-value');
  });

  it('setValueでlocalStorageに値が保存される', () => {
    const { result } = renderHook(() => useLocalStorage('test-key', 'default'));
    act(() => {
      result.current[1]('new-value');
    });
    expect(result.current[0]).toBe('new-value');
    expect(mockStorage.setItem).toHaveBeenCalledWith('test-key', JSON.stringify('new-value'));
  });

  it('removeValueでlocalStorageから値が削除される', () => {
    mockStorage.setItem('test-key', JSON.stringify('value'));
    const { result } = renderHook(() => useLocalStorage('test-key', 'default'));
    act(() => {
      result.current[2]();
    });
    expect(result.current[0]).toBe('default');
    expect(mockStorage.removeItem).toHaveBeenCalledWith('test-key');
  });

  it('オブジェクト型の値が正しく保存・復元される', () => {
    const obj = { name: '太郎', age: 25 };
    const { result } = renderHook(() => useLocalStorage('test-obj', { name: '', age: 0 }));
    act(() => {
      result.current[1](obj);
    });
    expect(result.current[0]).toEqual(obj);
    expect(mockStorage.setItem).toHaveBeenCalledWith('test-obj', JSON.stringify(obj));
  });

  it('配列型の値が正しく保存・復元される', () => {
    const arr = [1, 2, 3];
    const { result } = renderHook(() => useLocalStorage<number[]>('test-arr', []));
    act(() => {
      result.current[1](arr);
    });
    expect(result.current[0]).toEqual(arr);
  });

  it('不正なJSONの場合デフォルト値にフォールバックする', () => {
    mockStorage.setItem('test-key', 'invalid-json');
    const { result } = renderHook(() => useLocalStorage('test-key', 'fallback'));
    expect(result.current[0]).toBe('fallback');
  });

  it('数値型が正しく扱われる', () => {
    const { result } = renderHook(() => useLocalStorage('test-num', 0));
    act(() => {
      result.current[1](42);
    });
    expect(result.current[0]).toBe(42);
  });

  it('boolean型が正しく扱われる', () => {
    const { result } = renderHook(() => useLocalStorage('test-bool', false));
    act(() => {
      result.current[1](true);
    });
    expect(result.current[0]).toBe(true);
  });

  it('関数型のsetValueが前の値を参照できる', () => {
    const { result } = renderHook(() => useLocalStorage('test-fn', 10));
    act(() => {
      result.current[1]((prev) => prev + 5);
    });
    expect(result.current[0]).toBe(15);
  });

  it('null値がlocalStorageに保存されてもデフォルト値にフォールバックしない', () => {
    mockStorage.setItem('test-null', JSON.stringify(null));
    const { result } = renderHook(() => useLocalStorage<string | null>('test-null', 'default'));
    expect(result.current[0]).toBeNull();
  });

  it('removeValue後にsetValueで再度値を保存できる', () => {
    const { result } = renderHook(() => useLocalStorage('test-reuse', 'initial'));
    act(() => {
      result.current[1]('saved');
    });
    expect(result.current[0]).toBe('saved');
    act(() => {
      result.current[2]();
    });
    expect(result.current[0]).toBe('initial');
    act(() => {
      result.current[1]('re-saved');
    });
    expect(result.current[0]).toBe('re-saved');
    expect(mockStorage.setItem).toHaveBeenLastCalledWith('test-reuse', JSON.stringify('re-saved'));
  });

  it('空文字列がデフォルト値にフォールバックしない', () => {
    mockStorage.setItem('test-empty', JSON.stringify(''));
    const { result } = renderHook(() => useLocalStorage('test-empty', 'fallback'));
    expect(result.current[0]).toBe('');
  });
});
