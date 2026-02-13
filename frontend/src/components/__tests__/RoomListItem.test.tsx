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
});
