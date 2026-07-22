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

  it('大きな数値が正しく表示される', () => {
    render(<LineCount lineCount={1000} />);
    expect(screen.getByText('1000行')).toBeInTheDocument();
  });

  it('span要素としてレンダリングされる', () => {
    render(<LineCount lineCount={3} />);
    const element = screen.getByText('3行');
    expect(element.tagName).toBe('SPAN');
  });
});
