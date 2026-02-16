import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import ProfilePage from '../ProfilePage';
import ProfileRepository from '../../repositories/ProfileRepository';

const mockUpload = vi.fn();
vi.mock('../../hooks/useProfileImageUpload', () => ({
  useProfileImageUpload: () => ({
    upload: mockUpload,
    uploading: false,
  }),
}));

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

  it('カメラボタンが表示される', async () => {
    mockedRepo.fetchProfile.mockResolvedValue({ name: 'テスト', bio: '' });

    render(<ProfilePage />);

    await waitFor(() => {
      expect(screen.getByLabelText('プロフィール画像を変更')).toBeInTheDocument();
    });
  });

  it('画像アップロード成功時にアバターが更新される', async () => {
    mockedRepo.fetchProfile.mockResolvedValue({ name: 'テスト', bio: '' });
    mockUpload.mockResolvedValue('https://cdn.example.com/profiles/1/avatar.png');

    render(<ProfilePage />);

    await waitFor(() => {
      expect(screen.getByLabelText('プロフィール画像を変更')).toBeInTheDocument();
    });

    const fileInput = screen.getByTestId('profile-image-input');
    const file = new File(['test'], 'avatar.png', { type: 'image/png' });
    fireEvent.change(fileInput, { target: { files: [file] } });

    await waitFor(() => {
      expect(mockUpload).toHaveBeenCalledWith(file);
    });
  });

  it('画像アップロード失敗時にエラーメッセージが表示される', async () => {
    mockedRepo.fetchProfile.mockResolvedValue({ name: 'テスト', bio: '' });
    mockUpload.mockResolvedValue(null);

    render(<ProfilePage />);

    await waitFor(() => {
      expect(screen.getByLabelText('プロフィール画像を変更')).toBeInTheDocument();
    });

    const fileInput = screen.getByTestId('profile-image-input');
    const file = new File(['test'], 'avatar.png', { type: 'image/png' });
    fireEvent.change(fileInput, { target: { files: [file] } });

    await waitFor(() => {
      expect(screen.getByText('画像のアップロードに失敗しました。')).toBeInTheDocument();
    });
  });
});
