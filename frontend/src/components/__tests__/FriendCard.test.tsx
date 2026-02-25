import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import FriendCard from '../FriendCard';
import type { FriendshipUser } from '../../types';

const baseUser: FriendshipUser = {
  id: 1,
  userId: 10,
  username: 'テストユーザー',
  iconUrl: null,
  bio: '自己紹介文です',
  mutual: false,
  createdAt: '2024-01-01',
  status: 'オンライン',
};

describe('FriendCard', () => {
  it('ユーザー名が表示される', () => {
    render(<FriendCard user={baseUser} onUnfollow={() => {}} showUnfollow={false} />);
    expect(screen.getByText('テストユーザー')).toBeInTheDocument();
  });

  it('相互フォロー時にバッジが表示される', () => {
    render(<FriendCard user={{ ...baseUser, mutual: true }} onUnfollow={() => {}} showUnfollow={false} />);
    expect(screen.getByText('相互')).toBeInTheDocument();
  });

  it('ステータスが表示される', () => {
    render(<FriendCard user={baseUser} onUnfollow={() => {}} showUnfollow={false} />);
    expect(screen.getByText('オンライン')).toBeInTheDocument();
  });

  it('showUnfollow=trueでフォロー解除ボタンが表示される', () => {
    const onUnfollow = vi.fn();
    render(<FriendCard user={baseUser} onUnfollow={onUnfollow} showUnfollow={true} />);
    const btn = screen.getByRole('button', { name: 'フォロー解除' });
    fireEvent.click(btn);
    expect(onUnfollow).toHaveBeenCalledWith(10);
  });

  it('showUnfollow=falseでフォロー解除ボタンが非表示', () => {
    render(<FriendCard user={baseUser} onUnfollow={() => {}} showUnfollow={false} />);
    expect(screen.queryByRole('button', { name: 'フォロー解除' })).not.toBeInTheDocument();
  });
});
