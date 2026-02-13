import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import MemberPage from '../MemberPage';

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

describe('MemberPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseUserSearch.mockReturnValue(defaultData());
  });

  it('ヘッダーが表示される', () => {
    render(<BrowserRouter><MemberPage /></BrowserRouter>);

    expect(screen.getByText('チャットメンバー')).toBeInTheDocument();
    expect(screen.getByText('メンバーを検索または選択')).toBeInTheDocument();
  });

  it('検索ボックスが表示される', () => {
    render(<BrowserRouter><MemberPage /></BrowserRouter>);

    expect(screen.getByPlaceholderText('メンバーを検索...')).toBeInTheDocument();
  });

  it('メンバーがいない場合にメッセージを表示する', () => {
    render(<BrowserRouter><MemberPage /></BrowserRouter>);

    expect(screen.getByText('メンバーがまだいません')).toBeInTheDocument();
  });

  it('メンバー一覧を表示する', () => {
    mockUseUserSearch.mockReturnValue({
      ...defaultData(),
      users: [
        { id: 1, name: '鈴木一郎', email: 'suzuki@example.com', roomId: 10 },
      ],
    });

    render(<BrowserRouter><MemberPage /></BrowserRouter>);

    expect(screen.getByText('鈴木一郎')).toBeInTheDocument();
  });

  it('エラーメッセージを表示する', () => {
    mockUseUserSearch.mockReturnValue({
      ...defaultData(),
      error: 'ネットワークエラー',
    });

    render(<BrowserRouter><MemberPage /></BrowserRouter>);

    expect(screen.getByText('ネットワークエラー')).toBeInTheDocument();
  });

  it('エラー時にメンバーなしメッセージは表示しない', () => {
    mockUseUserSearch.mockReturnValue({
      ...defaultData(),
      error: 'エラー',
    });

    render(<BrowserRouter><MemberPage /></BrowserRouter>);

    expect(screen.queryByText('メンバーがまだいません')).not.toBeInTheDocument();
  });
});
