import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import PracticeTimer from '../PracticeTimer';

const mockStart = vi.fn();
const mockStop = vi.fn();

vi.mock('../../hooks/usePracticeTimer', () => ({
  usePracticeTimer: () => ({
    seconds: 125,
    isRunning: true,
    formattedTime: '02:05',
    start: mockStart,
    stop: mockStop,
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

  it('マウント時にstartが呼ばれる', () => {
    render(<PracticeTimer />);

    expect(mockStart).toHaveBeenCalled();
  });

  it('アンマウント時にstopが呼ばれる', () => {
    const { unmount } = render(<PracticeTimer />);

    unmount();

    expect(mockStop).toHaveBeenCalled();
  });

  it('時間がmonospaceフォントで表示される', () => {
    render(<PracticeTimer />);
    const timeElement = screen.getByText('02:05');
    expect(timeElement.className).toContain('font-mono');
  });

  it('ラベルと時間が同じコンテナ内にある', () => {
    const { container } = render(<PracticeTimer />);
    const wrapper = container.firstElementChild as HTMLElement;
    expect(wrapper.className).toContain('flex');
    expect(wrapper.className).toContain('items-center');
  });
});
