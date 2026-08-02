import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import NotificationItem from '../NotificationItem';
import type { Notification } from '../../model/types';

/**
 * 実際に backend が作る通知に合わせる（FRESTYLE-87）。
 *
 * 以前のテストは存在しない種別（GOAL_ACHIEVED）と存在しない項目（message）で
 * 書かれていたため、本文が画面に出ていないのにテストは通り続けていた。
 * backend が入れる値は type=company_application、本文は body。
 */
function makeNotification(overrides: Partial<Notification> = {}): Notification {
  return {
    id: 1,
    type: 'company_application',
    title: '新しい利用申請が届きました',
    body: '株式会社サンプル（山田 太郎 / taro@example.com）から利用申請がありました。',
    isRead: false,
    createdAt: '2026-08-02T10:00:00Z',
    ...overrides,
  };
}

function renderItem(overrides: Partial<Notification> = {}) {
  const onMarkAsRead = vi.fn();
  render(<NotificationItem notification={makeNotification(overrides)} onMarkAsRead={onMarkAsRead} />);
  return { onMarkAsRead };
}

describe('NotificationItem', () => {
  it('本文を表示する', () => {
    renderItem();

    expect(
      screen.getByText('株式会社サンプル（山田 太郎 / taro@example.com）から利用申請がありました。'),
    ).toBeInTheDocument();
  });

  it('タイトルを表示する', () => {
    renderItem();

    expect(screen.getByText('新しい利用申請が届きました')).toBeInTheDocument();
  });

  it('利用申請の種別を日本語のバッジで出す', () => {
    renderItem({ type: 'company_application' });

    expect(screen.getByText('利用申請')).toBeInTheDocument();
    expect(screen.queryByText('company_application')).not.toBeInTheDocument();
  });

  // ラベル未定義の種別が来ても表示が消えないこと（種別が増えたときの安全側の挙動）。
  it('未知の種別はそのまま表示する', () => {
    renderItem({ type: 'unknown_type' });

    expect(screen.getByText('unknown_type')).toBeInTheDocument();
  });

  it('本文が空でもタイトルは表示する', () => {
    renderItem({ body: '' });

    expect(screen.getByText('新しい利用申請が届きました')).toBeInTheDocument();
  });

  describe('既読の操作', () => {
    it('未読なら既読ボタンを出し、押すと id を渡す', () => {
      const { onMarkAsRead } = renderItem({ id: 42, isRead: false });

      fireEvent.click(screen.getByRole('button', { name: '既読にする' }));

      expect(onMarkAsRead).toHaveBeenCalledWith(42);
    });

    it('既読なら既読ボタンを出さない', () => {
      renderItem({ isRead: true });

      expect(screen.queryByRole('button', { name: '既読にする' })).not.toBeInTheDocument();
    });
  });
});
