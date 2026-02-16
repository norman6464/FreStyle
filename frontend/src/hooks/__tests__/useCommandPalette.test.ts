import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useCommandPalette } from '../useCommandPalette';
import { COMMAND_ITEMS } from '../../constants/commandPaletteItems';

describe('useCommandPalette', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('初期状態ではパレットが閉じている', () => {
    const { result } = renderHook(() => useCommandPalette());
    expect(result.current.isOpen).toBe(false);
    expect(result.current.query).toBe('');
    expect(result.current.selectedIndex).toBe(0);
  });

  it('openでパレットが開く', () => {
    const { result } = renderHook(() => useCommandPalette());
    act(() => {
      result.current.open();
    });
    expect(result.current.isOpen).toBe(true);
  });

  it('closeでパレットが閉じてクエリがリセットされる', () => {
    const { result } = renderHook(() => useCommandPalette());
    act(() => {
      result.current.open();
      result.current.setQuery('テスト');
    });
    act(() => {
      result.current.close();
    });
    expect(result.current.isOpen).toBe(false);
    expect(result.current.query).toBe('');
    expect(result.current.selectedIndex).toBe(0);
  });

  it('クエリが空のとき全コマンドを返す', () => {
    const { result } = renderHook(() => useCommandPalette());
    act(() => {
      result.current.open();
    });
    expect(result.current.filteredItems).toEqual(COMMAND_ITEMS);
  });

  it('クエリでラベルをフィルタリングできる', () => {
    const { result } = renderHook(() => useCommandPalette());
    act(() => {
      result.current.open();
      result.current.setQuery('ノート');
    });
    const labels = result.current.filteredItems.map(i => i.label);
    expect(labels).toContain('ノート');
    expect(labels).toContain('新規ノート作成');
    expect(labels).not.toContain('ホーム');
  });

  it('キーワードでもフィルタリングできる', () => {
    const { result } = renderHook(() => useCommandPalette());
    act(() => {
      result.current.open();
      result.current.setQuery('dark');
    });
    const labels = result.current.filteredItems.map(i => i.label);
    expect(labels).toContain('テーマ切替');
  });

  it('descriptionでもフィルタリングできる', () => {
    const { result } = renderHook(() => useCommandPalette());
    act(() => {
      result.current.open();
      result.current.setQuery('ダークモード');
    });
    const labels = result.current.filteredItems.map(i => i.label);
    expect(labels).toContain('テーマ切替');
  });

  it('フィルタリングは大文字小文字を区別しない', () => {
    const { result } = renderHook(() => useCommandPalette());
    act(() => {
      result.current.open();
      result.current.setQuery('AI');
    });
    const labels = result.current.filteredItems.map(i => i.label);
    expect(labels).toContain('AI アシスタント');
  });

  it('クエリ変更時にselectedIndexが0にリセットされる', () => {
    const { result } = renderHook(() => useCommandPalette());
    act(() => {
      result.current.open();
      result.current.selectNext();
      result.current.selectNext();
    });
    expect(result.current.selectedIndex).toBe(2);
    act(() => {
      result.current.setQuery('ノート');
    });
    expect(result.current.selectedIndex).toBe(0);
  });

  it('selectNextでインデックスが進む', () => {
    const { result } = renderHook(() => useCommandPalette());
    act(() => {
      result.current.open();
      result.current.selectNext();
    });
    expect(result.current.selectedIndex).toBe(1);
  });

  it('selectNextは最後のアイテムでループする', () => {
    const { result } = renderHook(() => useCommandPalette());
    act(() => {
      result.current.open();
    });
    const count = result.current.filteredItems.length;
    for (let i = 0; i < count; i++) {
      act(() => {
        result.current.selectNext();
      });
    }
    expect(result.current.selectedIndex).toBe(0);
  });

  it('selectPrevでインデックスが戻る', () => {
    const { result } = renderHook(() => useCommandPalette());
    act(() => {
      result.current.open();
      result.current.selectNext();
      result.current.selectNext();
      result.current.selectPrev();
    });
    expect(result.current.selectedIndex).toBe(1);
  });

  it('selectPrevは最初のアイテムで最後にループする', () => {
    const { result } = renderHook(() => useCommandPalette());
    act(() => {
      result.current.open();
      result.current.selectPrev();
    });
    expect(result.current.selectedIndex).toBe(result.current.filteredItems.length - 1);
  });
});
