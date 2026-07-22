import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import CourseProgressBar from '../CourseProgressBar';

describe('CourseProgressBar', () => {
  it('完了数/全体数・パーセント・残り章数を表示する', () => {
    render(<CourseProgressBar completed={3} total={12} />);
    expect(screen.getByText('3/12（25%・残り 9 章）')).toBeInTheDocument();
    const bar = screen.getByRole('progressbar', { name: '学習の進捗' });
    expect(bar).toHaveAttribute('aria-valuenow', '25');
    expect(bar).toHaveAttribute('aria-valuemin', '0');
    expect(bar).toHaveAttribute('aria-valuemax', '100');
  });

  it('total=0 のときは 0% として表示する(ゼロ除算しない・「すべて完了」にもしない)', () => {
    render(<CourseProgressBar completed={0} total={0} />);
    expect(screen.getByText('0/0（0%・残り 0 章）')).toBeInTheDocument();
    expect(screen.queryByText(/すべて完了/)).not.toBeInTheDocument();
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '0');
  });

  it('全完了は「すべて完了」と表示する', () => {
    render(<CourseProgressBar completed={8} total={8} />);
    expect(screen.getByText('すべて完了（8 章）')).toBeInTheDocument();
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '100');
  });

  it('completed > total はクランプして「すべて完了」止まり(呼び出し元のデータ差に対する防御)', () => {
    render(<CourseProgressBar completed={5} total={3} />);
    expect(screen.getByText('すべて完了（3 章）')).toBeInTheDocument();
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '100');
  });

  it('負の completed は 0 にクランプする', () => {
    render(<CourseProgressBar completed={-1} total={3} />);
    expect(screen.getByText('0/3（0%・残り 3 章）')).toBeInTheDocument();
  });
});
