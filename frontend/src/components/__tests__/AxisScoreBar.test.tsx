import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import AxisScoreBar from '../AxisScoreBar';

describe('AxisScoreBar', () => {
  it('プログレスバーが表示される', () => {
    render(<AxisScoreBar score={7} />);
    const bar = screen.getByRole('progressbar');
    expect(bar).toBeInTheDocument();
  });

  it('スコアに応じた幅が設定される', () => {
    render(<AxisScoreBar score={8} />);
    const bar = screen.getByRole('progressbar');
    const inner = bar.firstChild as HTMLElement;
    expect(inner.style.width).toBe('80%');
  });

  it('スコア0で幅が0%になる', () => {
    render(<AxisScoreBar score={0} />);
    const bar = screen.getByRole('progressbar');
    const inner = bar.firstChild as HTMLElement;
    expect(inner.style.width).toBe('0%');
  });

  it('カスタムバー色クラスを適用できる', () => {
    render(<AxisScoreBar score={5} barColorClass="bg-emerald-500" />);
    const bar = screen.getByRole('progressbar');
    const inner = bar.firstChild as HTMLElement;
    expect(inner.className).toContain('bg-emerald-500');
  });

  it('デフォルトでbg-primary-500が適用される', () => {
    render(<AxisScoreBar score={5} />);
    const bar = screen.getByRole('progressbar');
    const inner = bar.firstChild as HTMLElement;
    expect(inner.className).toContain('bg-primary-500');
  });
});
