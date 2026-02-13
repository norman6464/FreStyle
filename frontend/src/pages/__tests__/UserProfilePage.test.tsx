import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import UserProfilePage from '../UserProfilePage';

const mockFetchMyProfile = vi.fn();
const mockUpdateProfile = vi.fn();

vi.mock('../../hooks/useUserProfile', () => ({
  useUserProfile: () => ({
    profile: null,
    loading: false,
    error: null,
    fetchMyProfile: mockFetchMyProfile,
    updateProfile: mockUpdateProfile,
  }),
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => vi.fn(),
  };
});

function renderUserProfilePage() {
  return render(
    <MemoryRouter>
      <UserProfilePage />
    </MemoryRouter>
  );
}

describe('UserProfilePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('基本情報セクションが表示される', () => {
    renderUserProfilePage();

    expect(screen.getByText('基本情報')).toBeInTheDocument();
  });

  it('コミュニケーションスタイルセクションが表示される', () => {
    renderUserProfilePage();

    expect(screen.getByText('コミュニケーションスタイル')).toBeInTheDocument();
  });

  it('AIフィードバック設定セクションが表示される', () => {
    renderUserProfilePage();

    expect(screen.getByText('AIフィードバック設定')).toBeInTheDocument();
  });

  it('性格特性の選択肢が表示される', () => {
    renderUserProfilePage();

    expect(screen.getByText('内向的')).toBeInTheDocument();
    expect(screen.getByText('外向的')).toBeInTheDocument();
    expect(screen.getByText('論理的')).toBeInTheDocument();
  });

  it('保存ボタンが表示される', () => {
    renderUserProfilePage();

    expect(screen.getByRole('button', { name: /パーソナリティを/ })).toBeInTheDocument();
  });
});
