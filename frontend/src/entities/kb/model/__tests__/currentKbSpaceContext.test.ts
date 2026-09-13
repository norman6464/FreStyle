import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, afterEach } from 'vitest';
import { setCurrentKbSpace, useCurrentKbSpace } from '../currentKbSpaceContext';

describe('currentKbSpaceContext', () => {
  afterEach(() => {
    setCurrentKbSpace(null);
  });

  it('既定は null', () => {
    const { result } = renderHook(() => useCurrentKbSpace());
    expect(result.current).toBeNull();
  });

  it('setCurrentKbSpace で更新すると購読しているフックへ反映される', () => {
    const { result } = renderHook(() => useCurrentKbSpace());
    act(() => {
      setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-1' });
    });
    expect(result.current).toEqual({ workspaceSlug: 'acme', spaceId: 'space-1' });
  });

  it('複数のフックインスタンスへ同時に配る', () => {
    const a = renderHook(() => useCurrentKbSpace());
    const b = renderHook(() => useCurrentKbSpace());
    act(() => {
      setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-2' });
    });
    expect(a.result.current).toEqual({ workspaceSlug: 'acme', spaceId: 'space-2' });
    expect(b.result.current).toEqual({ workspaceSlug: 'acme', spaceId: 'space-2' });
  });

  it('null に戻せる', () => {
    const { result } = renderHook(() => useCurrentKbSpace());
    act(() => {
      setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-1' });
    });
    act(() => {
      setCurrentKbSpace(null);
    });
    expect(result.current).toBeNull();
  });

  it('マウント時点で既に確定している値をすぐ受け取る', () => {
    setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-1' });
    const { result } = renderHook(() => useCurrentKbSpace());
    expect(result.current).toEqual({ workspaceSlug: 'acme', spaceId: 'space-1' });
  });

  it('unmount 後は listener から外れる（後続の setCurrentKbSpace で例外にならない）', () => {
    const { unmount } = renderHook(() => useCurrentKbSpace());
    unmount();
    expect(() => setCurrentKbSpace({ workspaceSlug: 'acme', spaceId: 'space-1' })).not.toThrow();
  });
});
