import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { useResizablePanel } from '../useResizablePanel';

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

function dragHandle(onHandleMouseDown: (e: React.MouseEvent) => void, startX: number, moves: number[]) {
  act(() => {
    onHandleMouseDown({ clientX: startX, preventDefault: () => {} } as React.MouseEvent);
  });
  for (const x of moves) {
    act(() => {
      document.dispatchEvent(new MouseEvent('mousemove', { clientX: x }));
    });
  }
}

function releaseHandle() {
  act(() => {
    document.dispatchEvent(new MouseEvent('mouseup'));
  });
}

describe('useResizablePanel', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', createMockStorage());
    vi.stubGlobal('innerWidth', 1200);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('既定は defaultWidth（省略時 288px）', () => {
    const { result } = renderHook(() => useResizablePanel());
    expect(result.current.width).toBe(288);
    expect(result.current.isResizing).toBe(false);
  });

  it('左のパネルは右へドラッグすると広がる', () => {
    const { result } = renderHook(() => useResizablePanel({ side: 'left' }));
    dragHandle(result.current.onHandleMouseDown, 100, [140]);
    expect(result.current.width).toBe(328);
    expect(result.current.isResizing).toBe(true);
    releaseHandle();
    expect(result.current.isResizing).toBe(false);
  });

  it('右のパネルは左へドラッグすると広がる（符号が逆）', () => {
    const { result } = renderHook(() => useResizablePanel({ side: 'right', defaultWidth: 420 }));
    dragHandle(result.current.onHandleMouseDown, 500, [460]);
    expect(result.current.width).toBe(460);
    releaseHandle();
  });

  it('minWidth を下回らない', () => {
    const { result } = renderHook(() => useResizablePanel({ side: 'left', defaultWidth: 288, minWidth: 288 }));
    dragHandle(result.current.onHandleMouseDown, 100, [50]);
    expect(result.current.width).toBe(288);
    releaseHandle();
  });

  it('画面幅の maxWidthRatio（既定 0.5）を超えない', () => {
    // innerWidth=1200 → 上限 600px
    const { result } = renderHook(() => useResizablePanel({ side: 'left', defaultWidth: 288 }));
    dragHandle(result.current.onHandleMouseDown, 100, [1000]);
    expect(result.current.width).toBe(600);
    releaseHandle();
  });

  it('マウスアップ後は幅が storageKey に保存される', () => {
    const { result } = renderHook(() => useResizablePanel({ side: 'left', storageKey: 'test.panel.width' }));
    dragHandle(result.current.onHandleMouseDown, 100, [150]);
    releaseHandle();
    expect(JSON.parse(localStorage.getItem('test.panel.width')!)).toBe(338);
  });

  it('保存済みの幅があれば再マウントで復元される', () => {
    localStorage.setItem('test.panel.width2', JSON.stringify(400));
    const { result } = renderHook(() => useResizablePanel({ storageKey: 'test.panel.width2', defaultWidth: 288 }));
    expect(result.current.width).toBe(400);
  });

  it('storageKey を省略すると保存されない（再マウントで defaultWidth に戻る）', () => {
    const first = renderHook(() => useResizablePanel({ side: 'left', defaultWidth: 288 }));
    dragHandle(first.result.current.onHandleMouseDown, 100, [200]);
    releaseHandle();
    first.unmount();

    const second = renderHook(() => useResizablePanel({ side: 'left', defaultWidth: 288 }));
    expect(second.result.current.width).toBe(288);
  });
});
