import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import PracticeTimer from '../PracticeTimer';

vi.mock('../../hooks/usePracticeTimer', () => ({
  usePracticeTimer: () => ({
    seconds: 125,
    isRunning: true,
    formattedTime: '02:05',
    start: vi.fn(),
    stop: vi.fn(),
    reset: vi.fn(),
  }),
}));

describe('PracticeTimer', () => {
  it('経過時間が表示される', () => {
    render(<PracticeTimer />);

    expect(screen.getByText('02:05')).toBeInTheDocument();
  });

  it('タイマーラベルが表示される', () => {
    render(<PracticeTimer />);

    expect(screen.getByText('経過時間')).toBeInTheDocument();
  });
});
