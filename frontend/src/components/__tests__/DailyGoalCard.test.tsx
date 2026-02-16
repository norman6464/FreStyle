import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import DailyGoalCard from '../DailyGoalCard';

vi.mock('../../hooks/useDailyGoal', () => ({
  useDailyGoal: () => mockGoalState,
}));

let mockGoalState = {
  goal: { date: '2026-02-13', target: 3, completed: 1 },
  isAchieved: false,
  progress: 33,
  setTarget: vi.fn(),
  incrementCompleted: vi.fn(),
};

describe('DailyGoalCard', () => {
  it('今日の目標と進捗が表示される', () => {
    render(<DailyGoalCard />);

    expect(screen.getByText('今日の目標')).toBeInTheDocument();
    expect(screen.getByText(/1/)).toBeInTheDocument();
    expect(screen.getByText(/3/)).toBeInTheDocument();
  });

  it('プログレスバーが表示される', () => {
    const { container } = render(<DailyGoalCard />);

    const progressBar = container.querySelector('[role="progressbar"]');
    expect(progressBar).toBeInTheDocument();
  });

  it('目標達成時にメッセージが表示される', () => {
    mockGoalState = {
      goal: { date: '2026-02-13', target: 3, completed: 3 },
      isAchieved: true,
      progress: 100,
      setTarget: vi.fn(),
      incrementCompleted: vi.fn(),
    };

    render(<DailyGoalCard />);

    expect(screen.getByText(/達成/)).toBeInTheDocument();

    // 元に戻す
    mockGoalState = {
      goal: { date: '2026-02-13', target: 3, completed: 1 },
      isAchieved: false,
      progress: 33,
      setTarget: vi.fn(),
      incrementCompleted: vi.fn(),
    };
  });

  it('未達成時に達成メッセージが表示されない', () => {
    render(<DailyGoalCard />);
    expect(screen.queryByText(/達成/)).toBeNull();
  });

  it('プログレスバーにaria属性が設定される', () => {
    const { container } = render(<DailyGoalCard />);

    const progressBar = container.querySelector('[role="progressbar"]');
    expect(progressBar).toHaveAttribute('aria-valuenow', '33');
    expect(progressBar).toHaveAttribute('aria-valuemin', '0');
    expect(progressBar).toHaveAttribute('aria-valuemax', '100');
  });

  it('回ラベルが表示される', () => {
    render(<DailyGoalCard />);
    expect(screen.getByText(/回/)).toBeInTheDocument();
  });

  describe('目標回数変更', () => {
    beforeEach(() => {
      mockGoalState = {
        goal: { date: '2026-02-13', target: 3, completed: 1 },
        isAchieved: false,
        progress: 33,
        setTarget: vi.fn(),
        incrementCompleted: vi.fn(),
      };
    });

    it('目標回数をクリックすると編集モードになる', () => {
      render(<DailyGoalCard />);

      fireEvent.click(screen.getByRole('button', { name: /目標を変更/ }));

      expect(screen.getByRole('spinbutton')).toBeInTheDocument();
    });

    it('編集モードで値を変更してEnterで保存する', () => {
      render(<DailyGoalCard />);

      fireEvent.click(screen.getByRole('button', { name: /目標を変更/ }));
      const input = screen.getByRole('spinbutton');
      fireEvent.change(input, { target: { value: '5' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(mockGoalState.setTarget).toHaveBeenCalledWith(5);
    });

    it('編集モードでEscapeを押すとキャンセルされる', () => {
      render(<DailyGoalCard />);

      fireEvent.click(screen.getByRole('button', { name: /目標を変更/ }));
      const input = screen.getByRole('spinbutton');
      fireEvent.change(input, { target: { value: '8' } });
      fireEvent.keyDown(input, { key: 'Escape' });

      expect(mockGoalState.setTarget).not.toHaveBeenCalled();
      expect(screen.queryByRole('spinbutton')).not.toBeInTheDocument();
    });

    it('編集モードでblurすると保存される', () => {
      render(<DailyGoalCard />);

      fireEvent.click(screen.getByRole('button', { name: /目標を変更/ }));
      const input = screen.getByRole('spinbutton');
      fireEvent.change(input, { target: { value: '7' } });
      fireEvent.blur(input);

      expect(mockGoalState.setTarget).toHaveBeenCalledWith(7);
    });

    it('目標回数が下限未満のとき1にクランプされる', () => {
      render(<DailyGoalCard />);

      fireEvent.click(screen.getByRole('button', { name: /目標を変更/ }));
      const input = screen.getByRole('spinbutton');
      fireEvent.change(input, { target: { value: '0' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(mockGoalState.setTarget).toHaveBeenCalledWith(1);
    });

    it('目標回数が上限を超えたとき10にクランプされる', () => {
      render(<DailyGoalCard />);

      fireEvent.click(screen.getByRole('button', { name: /目標を変更/ }));
      const input = screen.getByRole('spinbutton');
      fireEvent.change(input, { target: { value: '11' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(mockGoalState.setTarget).toHaveBeenCalledWith(10);
    });

    it('負の値のとき1にクランプされる', () => {
      render(<DailyGoalCard />);

      fireEvent.click(screen.getByRole('button', { name: /目標を変更/ }));
      const input = screen.getByRole('spinbutton');
      fireEvent.change(input, { target: { value: '-3' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(mockGoalState.setTarget).toHaveBeenCalledWith(1);
    });
  });
});
