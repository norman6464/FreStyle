import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ProgressBar from '../ProgressBar';

describe('ProgressBar', () => {
  it('progressbarロールが表示される', () => {
    render(<ProgressBar percentage={50} />);
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('パーセンテージに応じた幅が設定される', () => {
    render(<ProgressBar percentage={75} />);
    const bar = screen.getByRole('progressbar').firstElementChild as HTMLElement;
    expect(bar.style.width).toBe('75%');
  });

  it('デフォルトでprimary色が適用される', () => {
    render(<ProgressBar percentage={50} />);
    const bar = screen.getByRole('progressbar').firstElementChild as HTMLElement;
    expect(bar.className).toContain('bg-primary-500');
  });

  it('カスタム色クラスを指定できる', () => {
    render(<ProgressBar percentage={50} barColorClass="bg-emerald-500" />);
    const bar = screen.getByRole('progressbar').firstElementChild as HTMLElement;
    expect(bar.className).toContain('bg-emerald-500');
  });

  it('aria属性が正しく設定される', () => {
    render(<ProgressBar percentage={60} />);
    const progressbar = screen.getByRole('progressbar');
    expect(progressbar).toHaveAttribute('aria-valuenow', '60');
    expect(progressbar).toHaveAttribute('aria-valuemin', '0');
    expect(progressbar).toHaveAttribute('aria-valuemax', '100');
  });
});
