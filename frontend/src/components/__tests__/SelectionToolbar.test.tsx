import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import SelectionToolbar from '../SelectionToolbar';

describe('SelectionToolbar', () => {
  it('editor=nullの時は何も表示しない', () => {
    const containerRef = { current: document.createElement('div') };
    const { container } = render(<SelectionToolbar editor={null} containerRef={containerRef} />);
    expect(container.innerHTML).toBe('');
  });

  it('選択がない場合は表示されない', () => {
    const mockEditor = {
      state: { selection: { from: 0, to: 0 } },
      on: vi.fn(),
      off: vi.fn(),
      isActive: vi.fn(() => false),
    };
    const containerRef = { current: document.createElement('div') };
    const { container } = render(<SelectionToolbar editor={mockEditor as never} containerRef={containerRef} />);
    expect(container.innerHTML).toBe('');
  });
});
