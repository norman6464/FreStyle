import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import TaskListButton from '../TaskListButton';

describe('TaskListButton', () => {
  it('タスクリストボタンが表示される', () => {
    render(<TaskListButton onTaskList={vi.fn()} />);
    expect(screen.getByLabelText('タスクリスト')).toBeInTheDocument();
  });

  it('クリックでonTaskListが呼ばれる', () => {
    const onTaskList = vi.fn();
    render(<TaskListButton onTaskList={onTaskList} />);
    fireEvent.click(screen.getByLabelText('タスクリスト'));
    expect(onTaskList).toHaveBeenCalledTimes(1);
  });

  it('ToolbarIconButtonを使用している', () => {
    render(<TaskListButton onTaskList={vi.fn()} />);
    const button = screen.getByLabelText('タスクリスト');
    expect(button.tagName).toBe('BUTTON');
  });
});
