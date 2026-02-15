import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import PracticeLevelCard from '../PracticeLevelCard';

describe('PracticeLevelCard', () => {
  it('タイトルが表示される', () => {
    render(<PracticeLevelCard totalSessions={0} />);

    expect(screen.getByText('練習レベル')).toBeInTheDocument();
  });

  it('セッション0回でLv.1が表示される', () => {
    render(<PracticeLevelCard totalSessions={0} />);

    expect(screen.getByText('Lv.1')).toBeInTheDocument();
    expect(screen.getByText('ビギナー')).toBeInTheDocument();
  });

  it('セッション5回でLv.2が表示される', () => {
    render(<PracticeLevelCard totalSessions={5} />);

    expect(screen.getByText('Lv.2')).toBeInTheDocument();
  });

  it('次のレベルまでの残り回数が表示される', () => {
    render(<PracticeLevelCard totalSessions={3} />);

    expect(screen.getByText('次のレベルまであと2回')).toBeInTheDocument();
  });

  it('最高レベルの場合は祝福メッセージが表示される', () => {
    render(<PracticeLevelCard totalSessions={100} />);

    expect(screen.getByText('最高レベル到達！')).toBeInTheDocument();
  });

  it('プログレスバーが表示される', () => {
    render(<PracticeLevelCard totalSessions={3} />);

    const progressbar = screen.getByRole('progressbar');
    const bar = progressbar.firstElementChild;
    expect(bar).toHaveStyle({ width: '60%' });
  });
});
