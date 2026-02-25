import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import NotificationItem from '../NotificationItem';
import type { Notification } from '../../types';

const baseNotification: Notification = {
  id: 1,
  type: 'GOAL_ACHIEVED',
  title: '目標達成おめでとう',
  message: '今月の学習目標を達成しました',
  isRead: false,
  createdAt: '2024-06-15T10:00:00Z',
};

describe('NotificationItem', () => {
  it('タイトルとメッセージが表示される', () => {
    render(<NotificationItem notification={baseNotification} onMarkAsRead={() => {}} />);
    expect(screen.getByText('目標達成おめでとう')).toBeInTheDocument();
    expect(screen.getByText('今月の学習目標を達成しました')).toBeInTheDocument();
  });

  it('タイプラベルが表示される', () => {
    render(<NotificationItem notification={baseNotification} onMarkAsRead={() => {}} />);
    expect(screen.getByText('目標達成')).toBeInTheDocument();
  });

  it('未読時に既読ボタンが表示される', () => {
    const onMarkAsRead = vi.fn();
    render(<NotificationItem notification={baseNotification} onMarkAsRead={onMarkAsRead} />);
    const btn = screen.getByRole('button', { name: '既読にする' });
    fireEvent.click(btn);
    expect(onMarkAsRead).toHaveBeenCalledWith(1);
  });

  it('既読時に既読ボタンが非表示', () => {
    render(<NotificationItem notification={{ ...baseNotification, isRead: true }} onMarkAsRead={() => {}} />);
    expect(screen.queryByRole('button', { name: '既読にする' })).not.toBeInTheDocument();
  });
});
