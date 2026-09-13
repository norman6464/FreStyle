import { renderHook } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { createRef } from 'react';
import { useClickOutside } from '../useClickOutside';

function fireMouseDown(target: Element) {
  target.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }));
}

describe('useClickOutside', () => {
  it('ref の外側をクリックすると onOutside を呼ぶ', () => {
    const container = document.createElement('div');
    const inside = document.createElement('button');
    const outside = document.createElement('button');
    container.appendChild(inside);
    document.body.appendChild(container);
    document.body.appendChild(outside);

    const ref = createRef<HTMLDivElement>();
    Object.defineProperty(ref, 'current', { value: container, writable: true });
    const onOutside = vi.fn();
    renderHook(() => useClickOutside(ref, true, onOutside));

    fireMouseDown(outside);
    expect(onOutside).toHaveBeenCalledTimes(1);

    document.body.removeChild(container);
    document.body.removeChild(outside);
  });

  it('ref の内側をクリックしても呼ばない', () => {
    const container = document.createElement('div');
    const inside = document.createElement('button');
    container.appendChild(inside);
    document.body.appendChild(container);

    const ref = createRef<HTMLDivElement>();
    Object.defineProperty(ref, 'current', { value: container, writable: true });
    const onOutside = vi.fn();
    renderHook(() => useClickOutside(ref, true, onOutside));

    fireMouseDown(inside);
    expect(onOutside).not.toHaveBeenCalled();

    document.body.removeChild(container);
  });

  it('enabled が false のあいだは購読しない', () => {
    const outside = document.createElement('button');
    document.body.appendChild(outside);

    const ref = createRef<HTMLDivElement>();
    const onOutside = vi.fn();
    renderHook(() => useClickOutside(ref, false, onOutside));

    fireMouseDown(outside);
    expect(onOutside).not.toHaveBeenCalled();

    document.body.removeChild(outside);
  });

  it('unmount 後は購読を解除する', () => {
    const container = document.createElement('div');
    const outside = document.createElement('button');
    document.body.appendChild(container);
    document.body.appendChild(outside);

    const ref = createRef<HTMLDivElement>();
    Object.defineProperty(ref, 'current', { value: container, writable: true });
    const onOutside = vi.fn();
    const { unmount } = renderHook(() => useClickOutside(ref, true, onOutside));
    unmount();

    fireMouseDown(outside);
    expect(onOutside).not.toHaveBeenCalled();

    document.body.removeChild(container);
    document.body.removeChild(outside);
  });
});
