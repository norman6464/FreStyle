import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import CourseProgressBar from '../CourseProgressBar';

describe('CourseProgressBar', () => {
  it('完了数/全体数とパーセントを表示する', () => {
    render(<CourseProgressBar completed={3} total={12} />);
    expect(screen.getByText('3/12（25%）')).toBeInTheDocument();
    const bar = screen.getByRole('progressbar', { name: '学習の進捗' });
    expect(bar).toHaveAttribute('aria-valuenow', '25');
    expect(bar).toHaveAttribute('aria-valuemin', '0');
    expect(bar).toHaveAttribute('aria-valuemax', '100');
  });

  it('total=0 のときは 0% として表示する(ゼロ除算しない)', () => {
    render(<CourseProgressBar completed={0} total={0} />);
    expect(screen.getByText('0/0（0%）')).toBeInTheDocument();
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '0');
  });

  it('全完了は 100%', () => {
    render(<CourseProgressBar completed={8} total={8} />);
    expect(screen.getByText('8/8（100%）')).toBeInTheDocument();
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '100');
  });
});
