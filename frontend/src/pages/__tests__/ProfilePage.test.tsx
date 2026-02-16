import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import ProfilePage from '../ProfilePage';
import ProfileRepository from '../../repositories/ProfileRepository';

vi.mock('../../repositories/ProfileRepository');

const mockedRepo = vi.mocked(ProfileRepository);

describe('ProfilePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('ローディング中はスピナーが表示される', () => {
    mockedRepo.fetchProfile.mockReturnValue(new Promise(() => {}));
    render(<ProfilePage />);

    expect(document.querySelector('.animate-spin')).toBeInTheDocument();
  });

  it('プロファイル取得後にフォームが表示される', async () => {
    mockedRepo.fetchProfile.mockResolvedValue({ name: 'テストユーザー', bio: '自己紹介文' });

    render(<ProfilePage />);

    await waitFor(() => {
      expect(screen.getByText('プロフィールを編集')).toBeInTheDocument();
      expect(screen.getByText('プロフィールを更新')).toBeInTheDocument();
    });
  });

  it('プロファイル取得失敗時にエラーが表示される', async () => {
    mockedRepo.fetchProfile.mockRejectedValue(new Error('取得失敗'));

    render(<ProfilePage />);

    await waitFor(() => {
      expect(screen.getByText('プロフィール取得に失敗しました。')).toBeInTheDocument();
    });
  });

  it('プロファイル取得後にニックネーム欄が表示される', async () => {
    mockedRepo.fetchProfile.mockResolvedValue({ name: 'テストユーザー', bio: '自己紹介文' });

    render(<ProfilePage />);

    await waitFor(() => {
      expect(screen.getByDisplayValue('テストユーザー')).toBeInTheDocument();
    });
  });

  it('プロファイル取得後に自己紹介欄が表示される', async () => {
    mockedRepo.fetchProfile.mockResolvedValue({ name: 'テスト', bio: 'テスト自己紹介' });

    render(<ProfilePage />);

    await waitFor(() => {
      expect(screen.getByDisplayValue('テスト自己紹介')).toBeInTheDocument();
    });
  });

  it('送信中はボタンが「更新中...」になり無効化される', async () => {
    mockedRepo.fetchProfile.mockResolvedValue({ name: 'テスト', bio: '' });
    mockedRepo.updateProfile.mockReturnValue(new Promise(() => {}));

    render(<ProfilePage />);

    await waitFor(() => {
      expect(screen.getByText('プロフィールを更新')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('プロフィールを更新'));

    await waitFor(() => {
      expect(screen.getByText('更新中...')).toBeInTheDocument();
      expect(screen.getByText('更新中...').closest('button')).toBeDisabled();
    });
  });

  it('アバターのイニシャルが表示される', async () => {
    mockedRepo.fetchProfile.mockResolvedValue({ name: 'テストユーザー', bio: '' });

    render(<ProfilePage />);

    await waitFor(() => {
      expect(screen.getByText('テ')).toBeInTheDocument();
    });
  });
});
