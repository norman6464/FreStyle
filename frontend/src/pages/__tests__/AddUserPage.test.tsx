import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import AddUserPage from '../AddUserPage';

const mockUseUserSearch = vi.fn();
vi.mock('../../hooks/useUserSearch', () => ({
  useUserSearch: () => mockUseUserSearch(),
}));

function defaultData() {
  return {
    users: [],
    error: null,
    searchQuery: '',
    setSearchQuery: vi.fn(),
    debounceQuery: '',
  };
}

describe('AddUserPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseUserSearch.mockReturnValue(defaultData());
  });

  it('検索ボックスが表示される', () => {
    render(<BrowserRouter><AddUserPage /></BrowserRouter>);

    expect(screen.getByPlaceholderText('ユーザー名またはメールアドレスで検索...')).toBeInTheDocument();
  });

  it('初期状態で検索案内を表示する', () => {
    render(<BrowserRouter><AddUserPage /></BrowserRouter>);

    expect(screen.getByText('ユーザーを検索してみましょう')).toBeInTheDocument();
  });

  it('ユーザー一覧を表示する', () => {
    mockUseUserSearch.mockReturnValue({
      ...defaultData(),
      users: [
        { id: 1, name: '山田太郎', email: 'yamada@example.com', roomId: null },
      ],
    });

    render(<BrowserRouter><AddUserPage /></BrowserRouter>);

    expect(screen.getByText('山田太郎')).toBeInTheDocument();
  });

  it('エラーメッセージを表示する', () => {
    mockUseUserSearch.mockReturnValue({
      ...defaultData(),
      error: 'ユーザー取得に失敗しました',
    });

    render(<BrowserRouter><AddUserPage /></BrowserRouter>);

    expect(screen.getByText('ユーザー取得に失敗しました')).toBeInTheDocument();
  });

  it('検索結果なしの場合メッセージを表示する', () => {
    mockUseUserSearch.mockReturnValue({
      ...defaultData(),
      debounceQuery: 'テスト',
    });

    render(<BrowserRouter><AddUserPage /></BrowserRouter>);

    expect(screen.getByText('ユーザーが見つかりませんでした')).toBeInTheDocument();
  });

  it('ユーザー数を表示する', () => {
    mockUseUserSearch.mockReturnValue({
      ...defaultData(),
      users: [
        { id: 1, name: 'ユーザー1', email: 'u1@example.com', roomId: null },
        { id: 2, name: 'ユーザー2', email: 'u2@example.com', roomId: null },
      ],
    });

    render(<BrowserRouter><AddUserPage /></BrowserRouter>);

    expect(screen.getByText('2人のユーザーが見つかりました')).toBeInTheDocument();
  });
});
