import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import SharedSessionsPage from '../SharedSessionsPage';
import { useSharedSessions } from '../../hooks/useSharedSessions';

vi.mock('../../hooks/useSharedSessions');
const mockedUseSharedSessions = vi.mocked(useSharedSessions);

function renderPage() {
  return render(<MemoryRouter><SharedSessionsPage /></MemoryRouter>);
}

describe('SharedSessionsPage', () => {
  const mockSessions = [
    { id: 1, sessionId: 10, sessionTitle: '会議の練習', userId: 1, username: 'TestUser', userIconUrl: null, description: '良いセッション', createdAt: '2024-01-01' },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    mockedUseSharedSessions.mockReturnValue({
      sessions: mockSessions,
      loading: false,
      error: null,
      shareSession: vi.fn(),
      unshareSession: vi.fn(),
    });
  });

  it('タイトルを表示する', () => {
    renderPage();
    expect(screen.getByText('みんなの会話')).toBeInTheDocument();
  });

  it('共有セッションを表示する', () => {
    renderPage();
    expect(screen.getByText('会議の練習')).toBeInTheDocument();
    expect(screen.getByText('TestUser')).toBeInTheDocument();
  });

  it('ローディング中はローディング表示する', () => {
    mockedUseSharedSessions.mockReturnValue({ sessions: [], loading: true, error: null, shareSession: vi.fn(), unshareSession: vi.fn() });
    renderPage();
    expect(screen.getByText('共有セッションを読み込み中...')).toBeInTheDocument();
  });

  it('セッションがない場合は空メッセージを表示する', () => {
    mockedUseSharedSessions.mockReturnValue({ sessions: [], loading: false, error: null, shareSession: vi.fn(), unshareSession: vi.fn() });
    renderPage();
    expect(screen.getByText('まだ共有されたセッションはありません')).toBeInTheDocument();
  });

  it('エラー時はエラーメッセージを表示する', () => {
    mockedUseSharedSessions.mockReturnValue({ sessions: [], loading: false, error: '共有セッションの取得に失敗しました', shareSession: vi.fn(), unshareSession: vi.fn() });
    renderPage();
    expect(screen.getByText('共有セッションの取得に失敗しました')).toBeInTheDocument();
  });
});
