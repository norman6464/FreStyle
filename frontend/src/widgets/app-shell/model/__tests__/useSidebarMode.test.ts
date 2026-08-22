import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useSidebarMode } from '../useSidebarMode';

// この環境の jsdom は localStorage を提供しないため、既存テストと同じ流儀でスタブする。
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

describe('useSidebarMode', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', createMockStorage());
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('初期状態は collapsed（隠れている・一時表示もしていない）', () => {
    const { result } = renderHook(() => useSidebarMode());
    expect(result.current.mode).toBe('collapsed');
    expect(result.current.isPeeking).toBe(false);
    expect(result.current.isVisible).toBe(false);
  });

  it('openPeek で一時表示になり、closePeek は猶予後に閉じる', () => {
    const { result } = renderHook(() => useSidebarMode());
    act(() => result.current.openPeek());
    expect(result.current.isPeeking).toBe(true);
    expect(result.current.isVisible).toBe(true);

    act(() => result.current.closePeek());
    // 猶予中はまだ見えている（ボタンへ移動する途中で消えない）。
    expect(result.current.isPeeking).toBe(true);
    act(() => vi.runAllTimers());
    expect(result.current.isPeeking).toBe(false);
  });

  it('closePeek の猶予中に openPeek すると閉じない（ちらつき防止）', () => {
    const { result } = renderHook(() => useSidebarMode());
    act(() => result.current.openPeek());
    act(() => result.current.closePeek());
    act(() => result.current.openPeek());
    act(() => vi.runAllTimers());
    expect(result.current.isPeeking).toBe(true);
  });

  it('pin で固定表示になり localStorage に保存される', () => {
    const { result } = renderHook(() => useSidebarMode());
    act(() => result.current.pin());
    expect(result.current.mode).toBe('pinned');
    expect(result.current.isVisible).toBe(true);
    expect(JSON.parse(localStorage.getItem('frestyle.sidebar.mode')!)).toBe('pinned');
  });

  it('collapse で一時表示モードへ戻る', () => {
    const { result } = renderHook(() => useSidebarMode());
    act(() => result.current.pin());
    act(() => result.current.collapse());
    expect(result.current.mode).toBe('collapsed');
    expect(result.current.isVisible).toBe(false);
  });

  it('保存済みモード（pinned）で初期化される', () => {
    localStorage.setItem('frestyle.sidebar.mode', JSON.stringify('pinned'));
    const { result } = renderHook(() => useSidebarMode());
    expect(result.current.mode).toBe('pinned');
    expect(result.current.isVisible).toBe(true);
  });

  it('⌘\\ でモードをトグルする', () => {
    const { result } = renderHook(() => useSidebarMode());
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: '\\', metaKey: true }));
    });
    expect(result.current.mode).toBe('pinned');
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: '\\', ctrlKey: true }));
    });
    expect(result.current.mode).toBe('collapsed');
  });
});
