import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useAutoResizeTextarea } from '../useAutoResizeTextarea';

describe('useAutoResizeTextarea', () => {
  let mockTextarea: HTMLTextAreaElement;

  beforeEach(() => {
    mockTextarea = document.createElement('textarea');
    vi.spyOn(window, 'getComputedStyle').mockReturnValue({
      lineHeight: '24px',
      paddingTop: '8px',
      paddingBottom: '8px',
    } as CSSStyleDeclaration);
  });

  it('refオブジェクトを返す', () => {
    const { result } = renderHook(() => useAutoResizeTextarea({ text: '' }));

    expect(result.current).toHaveProperty('current');
  });

  it('テキスト変更時にtextareaの高さが更新される', () => {
    const { result, rerender } = renderHook(
      ({ text }) => useAutoResizeTextarea({ text }),
      { initialProps: { text: '' } }
    );

    // refにテキストエリアを設定
    Object.defineProperty(result.current, 'current', {
      value: mockTextarea,
      writable: true,
    });
    Object.defineProperty(mockTextarea, 'scrollHeight', { value: 40, configurable: true });

    rerender({ text: 'テスト' });

    expect(mockTextarea.style.height).not.toBe('');
  });

  it('maxRowsを超えた場合overflowYがscrollになる', () => {
    const { result, rerender } = renderHook(
      ({ text }) => useAutoResizeTextarea({ text, maxRows: 2 }),
      { initialProps: { text: '' } }
    );

    Object.defineProperty(result.current, 'current', {
      value: mockTextarea,
      writable: true,
    });
    // lineHeight=24, padding=16, maxHeight = 2*24+16 = 64
    // scrollHeight > maxHeight → scroll
    Object.defineProperty(mockTextarea, 'scrollHeight', { value: 100, configurable: true });

    rerender({ text: '長いテキスト\n複数行' });

    expect(mockTextarea.style.overflowY).toBe('scroll');
  });

  it('maxRows以内の場合overflowYがhiddenになる', () => {
    const { result, rerender } = renderHook(
      ({ text }) => useAutoResizeTextarea({ text, maxRows: 8 }),
      { initialProps: { text: '' } }
    );

    Object.defineProperty(result.current, 'current', {
      value: mockTextarea,
      writable: true,
    });
    Object.defineProperty(mockTextarea, 'scrollHeight', { value: 40, configurable: true });

    rerender({ text: 'テスト' });

    expect(mockTextarea.style.overflowY).toBe('hidden');
  });

  it('デフォルトのminRowsは1、maxRowsは8', () => {
    const { result, rerender } = renderHook(
      ({ text }) => useAutoResizeTextarea({ text }),
      { initialProps: { text: '' } }
    );

    Object.defineProperty(result.current, 'current', {
      value: mockTextarea,
      writable: true,
    });
    Object.defineProperty(mockTextarea, 'scrollHeight', { value: 40, configurable: true });

    rerender({ text: 'テスト' });

    // maxRows=8, lineHeight=24, padding=16 → maxHeight=208
    // scrollHeight(40) < maxHeight(208) → height=40px
    expect(mockTextarea.style.height).toBe('40px');
  });
});
