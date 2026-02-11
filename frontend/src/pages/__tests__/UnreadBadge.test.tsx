import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';

/**
 * UnreadBadgeコンポーネント（ChatListPage内で使用される未読バッジ）
 * テスト用に独立して抽出
 */
interface UnreadBadgeProps {
  unreadCount: number | null;
}

function UnreadBadge({ unreadCount }: UnreadBadgeProps) {
  if (!unreadCount || unreadCount <= 0) return null;
  return (
    <span
      data-testid="unread-badge"
      className="bg-pink-500 text-white text-xs font-bold px-2 py-0.5 rounded-full flex-shrink-0"
    >
      {unreadCount > 99 ? '99+' : unreadCount}
    </span>
  );
}

describe('UnreadBadge', () => {
  it('unreadCount > 0 の時バッジが表示される', () => {
    render(<UnreadBadge unreadCount={3} />);
    const badge = screen.getByTestId('unread-badge');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveTextContent('3');
  });

  it('unreadCount === 0 の時バッジが表示されない', () => {
    render(<UnreadBadge unreadCount={0} />);
    expect(screen.queryByTestId('unread-badge')).not.toBeInTheDocument();
  });

  it('unreadCount > 99 の時 "99+" が表示される', () => {
    render(<UnreadBadge unreadCount={150} />);
    const badge = screen.getByTestId('unread-badge');
    expect(badge).toHaveTextContent('99+');
  });

  it('unreadCount が null の時バッジが表示されない', () => {
    render(<UnreadBadge unreadCount={null} />);
    expect(screen.queryByTestId('unread-badge')).not.toBeInTheDocument();
  });
});
