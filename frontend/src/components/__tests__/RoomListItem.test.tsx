import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import RoomListItem from '../RoomListItem';

describe('RoomListItem', () => {
  it('名前と最新メッセージが表示される', () => {
    render(<RoomListItem name="テストユーザー" lastMessage="こんにちは" unreadCount={0} />);

    expect(screen.getByText('テストユーザー')).toBeInTheDocument();
    expect(screen.getByText('こんにちは')).toBeInTheDocument();
  });

  it('未読数が0のときバッジが表示されない', () => {
    const { container } = render(<RoomListItem name="テスト" lastMessage="メッセージ" unreadCount={0} />);

    expect(container.querySelector('span')).toBeNull();
  });

  it('未読数が1以上のときバッジが表示される', () => {
    render(<RoomListItem name="テスト" lastMessage="メッセージ" unreadCount={5} />);

    expect(screen.getByText('5')).toBeInTheDocument();
  });

  it('長いメッセージが切り詰められるスタイルが適用される', () => {
    const { container } = render(<RoomListItem name="テスト" lastMessage="とても長いメッセージ" unreadCount={0} />);
    const messageEl = container.querySelector('.truncate');
    expect(messageEl).toBeTruthy();
  });

  it('名前がフォント太字で表示される', () => {
    const { container } = render(<RoomListItem name="太字テスト" lastMessage="メッセージ" unreadCount={0} />);
    const nameEl = container.querySelector('.font-semibold');
    expect(nameEl).toHaveTextContent('太字テスト');
  });

  it('未読バッジにprimary背景色が適用される', () => {
    render(<RoomListItem name="テスト" lastMessage="メッセージ" unreadCount={3} />);
    const badge = screen.getByText('3');
    expect(badge.className).toContain('bg-primary-500');
  });
});
