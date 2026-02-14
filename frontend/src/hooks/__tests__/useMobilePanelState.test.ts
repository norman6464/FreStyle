import { describe, it, expect } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useMobilePanelState } from '../useMobilePanelState';

describe('useMobilePanelState', () => {
  it('初期状態でisOpenはfalse', () => {
    const { result } = renderHook(() => useMobilePanelState());
    expect(result.current.isOpen).toBe(false);
  });

  it('openでisOpenがtrueになる', () => {
    const { result } = renderHook(() => useMobilePanelState());

    act(() => {
      result.current.open();
    });

    expect(result.current.isOpen).toBe(true);
  });

  it('closeでisOpenがfalseになる', () => {
    const { result } = renderHook(() => useMobilePanelState());

    act(() => {
      result.current.open();
    });
    expect(result.current.isOpen).toBe(true);

    act(() => {
      result.current.close();
    });
    expect(result.current.isOpen).toBe(false);
  });

  it('open→close→openで正しく切り替わる', () => {
    const { result } = renderHook(() => useMobilePanelState());

    act(() => result.current.open());
    expect(result.current.isOpen).toBe(true);

    act(() => result.current.close());
    expect(result.current.isOpen).toBe(false);

    act(() => result.current.open());
    expect(result.current.isOpen).toBe(true);
  });

  it('closeを連続で呼んでもfalseのまま', () => {
    const { result } = renderHook(() => useMobilePanelState());

    act(() => result.current.close());
    act(() => result.current.close());
    expect(result.current.isOpen).toBe(false);
  });
});
