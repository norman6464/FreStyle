import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import FriendshipPage from '../FriendshipPage';

const mockUnfollow = vi.fn();

vi.mock('../../hooks/useFriendship', () => ({
  useFriendship: vi.fn(),
}));

import { useFriendship } from '../../hooks/useFriendship';
const mockedUseFriendship = vi.mocked(useFriendship);

const baseUser = {
  id: 1,
  userId: 10,
  username: 'テストユーザー',
  iconUrl: null,
  bio: '自己紹介',
  mutual: false,
  createdAt: '2024-01-01',
  status: 'オンライン',
};

describe('FriendshipPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('ローディング中はスピナーが表示される', () => {
    mockedUseFriendship.mockReturnValue({
      following: [],
      followers: [],
      loading: true,
      followUser: vi.fn(),
      unfollowUser: mockUnfollow,
      refetch: vi.fn(),
    });

    render(<FriendshipPage />);
    expect(screen.getByText('読み込み中...')).toBeInTheDocument();
  });

  it('フォロー中タブでユーザー一覧が表示される', () => {
    mockedUseFriendship.mockReturnValue({
      following: [baseUser],
      followers: [],
      loading: false,
      followUser: vi.fn(),
      unfollowUser: mockUnfollow,
      refetch: vi.fn(),
    });

    render(<FriendshipPage />);
    expect(screen.getByText('テストユーザー')).toBeInTheDocument();
    expect(screen.getByText('フォロー中 (1)')).toBeInTheDocument();
  });

  it('フォロワータブに切り替えできる', () => {
    mockedUseFriendship.mockReturnValue({
      following: [],
      followers: [{ ...baseUser, id: 2, userId: 20, username: 'フォロワー' }],
      loading: false,
      followUser: vi.fn(),
      unfollowUser: mockUnfollow,
      refetch: vi.fn(),
    });

    render(<FriendshipPage />);
    fireEvent.click(screen.getByText('フォロワー (1)'));
    expect(screen.getByText('フォロワー')).toBeInTheDocument();
  });

  it('フォロー中が0件の場合はEmptyStateが表示される', () => {
    mockedUseFriendship.mockReturnValue({
      following: [],
      followers: [],
      loading: false,
      followUser: vi.fn(),
      unfollowUser: mockUnfollow,
      refetch: vi.fn(),
    });

    render(<FriendshipPage />);
    expect(screen.getByText('まだ誰もフォローしていません')).toBeInTheDocument();
  });

  it('フォロワーが0件の場合はEmptyStateが表示される', () => {
    mockedUseFriendship.mockReturnValue({
      following: [],
      followers: [],
      loading: false,
      followUser: vi.fn(),
      unfollowUser: mockUnfollow,
      refetch: vi.fn(),
    });

    render(<FriendshipPage />);
    fireEvent.click(screen.getByText('フォロワー (0)'));
    expect(screen.getByText('まだフォロワーがいません')).toBeInTheDocument();
  });
});
