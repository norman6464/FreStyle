import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import ToolbarDivider from '../ToolbarDivider';

describe('ToolbarDivider', () => {
  it('区切り線が表示される', () => {
    const { container } = render(<ToolbarDivider />);
    const divider = container.firstChild as HTMLElement;
    expect(divider).toBeInTheDocument();
    expect(divider.tagName).toBe('DIV');
  });

  it('正しいクラスが適用される', () => {
    const { container } = render(<ToolbarDivider />);
    const divider = container.firstChild as HTMLElement;
    expect(divider.className).toContain('w-px');
    expect(divider.className).toContain('h-4');
  });
});
