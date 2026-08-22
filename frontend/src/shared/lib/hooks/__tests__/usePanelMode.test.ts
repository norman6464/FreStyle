import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { usePanelMode } from '../usePanelMode';

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

const KEY = 'frestyle.panel.test';

describe('usePanelMode', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', createMockStorage());
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('既定は pinned（パネル型 UI の従来挙動を壊さない）', () => {
    const { result } = renderHook(() => usePanelMode(KEY));
    expect(result.current.mode).toBe('pinned');
    expect(result.current.isVisible).toBe(true);
  });

  it('defaultMode を collapsed にもできる', () => {
    const { result } = renderHook(() => usePanelMode(KEY, { defaultMode: 'collapsed' }));
    expect(result.current.mode).toBe('collapsed');
    expect(result.current.isVisible).toBe(false);
  });

  it('collapse → openPeek で一時表示、closePeek は猶予後に閉じる', () => {
    const { result } = renderHook(() => usePanelMode(KEY));
    act(() => result.current.collapse());
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
    const { result } = renderHook(() => usePanelMode(KEY));
    act(() => result.current.collapse());
    act(() => result.current.openPeek());
    act(() => result.current.closePeek());
    act(() => result.current.openPeek());
    act(() => vi.runAllTimers());
    expect(result.current.isPeeking).toBe(true);
  });

  it('モードは storageKey へ保存され、再マウントで復元される', () => {
    const first = renderHook(() => usePanelMode(KEY));
    act(() => first.result.current.collapse());
    expect(JSON.parse(localStorage.getItem(KEY)!)).toBe('collapsed');
    first.unmount();

    const second = renderHook(() => usePanelMode(KEY));
    expect(second.result.current.mode).toBe('collapsed');
  });

  it('⌘\\ でモードをトグルする（shortcut: false なら無効）', () => {
    const { result } = renderHook(() => usePanelMode(KEY));
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: '\\', metaKey: true }));
    });
    expect(result.current.mode).toBe('collapsed');

    const noShortcut = renderHook(() => usePanelMode('frestyle.panel.other', { shortcut: false }));
    act(() => {
      document.dispatchEvent(new KeyboardEvent('keydown', { key: '\\', ctrlKey: true }));
    });
    expect(noShortcut.result.current.mode).toBe('pinned');
  });
});
