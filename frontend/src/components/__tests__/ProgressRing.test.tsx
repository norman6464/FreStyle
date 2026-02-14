import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import ProgressRing from '../ProgressRing';

describe('ProgressRing', () => {
  it('ラベルが表示される', () => {
    render(<ProgressRing value={75} max={100} label="75%" />);
    expect(screen.getByText('75%')).toBeInTheDocument();
  });

  it('SVG要素がレンダリングされる', () => {
    const { container } = render(<ProgressRing value={50} max={100} />);
    expect(container.querySelector('svg')).toBeTruthy();
  });

  it('role=progressbarが設定される', () => {
    render(<ProgressRing value={30} max={100} />);
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('aria-valuenowが正しく設定される', () => {
    render(<ProgressRing value={60} max={100} />);
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '60');
  });

  it('aria-valuemaxが正しく設定される', () => {
    render(<ProgressRing value={60} max={100} />);
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuemax', '100');
  });

  it('高スコア時（80%以上）に緑色のストロークが適用される', () => {
    const { container } = render(<ProgressRing value={85} max={100} />);
    const circles = container.querySelectorAll('circle');
    const progressCircle = circles[1];
    expect(progressCircle.getAttribute('class')).toContain('text-emerald-400');
  });

  it('中スコア時（60-80%）に黄色のストロークが適用される', () => {
    const { container } = render(<ProgressRing value={65} max={100} />);
    const circles = container.querySelectorAll('circle');
    const progressCircle = circles[1];
    expect(progressCircle.getAttribute('class')).toContain('text-amber-400');
  });

  it('低スコア時（60%未満）に赤色のストロークが適用される', () => {
    const { container } = render(<ProgressRing value={40} max={100} />);
    const circles = container.querySelectorAll('circle');
    const progressCircle = circles[1];
    expect(progressCircle.getAttribute('class')).toContain('text-rose-400');
  });

  it('value=0の場合にプログレスが0になる', () => {
    render(<ProgressRing value={0} max={100} />);
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '0');
  });

  it('value=maxの場合にプログレスが100%になる', () => {
    render(<ProgressRing value={100} max={100} />);
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '100');
  });

  it('ラベルが未指定の場合にテキストが表示されない', () => {
    const { container } = render(<ProgressRing value={50} max={100} />);
    const span = container.querySelector('span');
    expect(span).toBeNull();
  });

  it('aria-valueminが0に設定される', () => {
    render(<ProgressRing value={50} max={100} />);
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuemin', '0');
  });

  it('背景円と進捗円の2つのcircle要素がレンダリングされる', () => {
    const { container } = render(<ProgressRing value={50} max={100} />);
    const circles = container.querySelectorAll('circle');
    expect(circles).toHaveLength(2);
  });
});
