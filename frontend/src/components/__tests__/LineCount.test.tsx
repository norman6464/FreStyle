import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import LineCount from '../LineCount';

describe('LineCount', () => {
  it('行数が表示される', () => {
    render(<LineCount lineCount={10} />);
    expect(screen.getByText('10行')).toBeInTheDocument();
  });

  it('0行が正しく表示される', () => {
    render(<LineCount lineCount={0} />);
    expect(screen.getByText('0行')).toBeInTheDocument();
  });

  it('正しいスタイルクラスが適用される', () => {
    render(<LineCount lineCount={5} />);
    const element = screen.getByText('5行');
    expect(element.className).toContain('text-[10px]');
  });
});
