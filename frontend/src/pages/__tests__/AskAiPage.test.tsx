import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import AskAiPage from '../AskAiPage';
import { useAskAi } from '../../hooks/useAskAi';
import { useCopyToClipboard } from '../../hooks/useCopyToClipboard';

vi.mock('../../hooks/useAskAi');
vi.mock('../../hooks/useCopyToClipboard');
vi.mock('../../hooks/useBlockEditor', () => ({
  useBlockEditor: () => ({ editor: null }),
}));

const mockUseAskAi = {
  sessions: [],
  filteredSessions: [],
  messages: [],
  scoreCard: null,
  messagesEndRef: { current: null },
  isPracticeMode: false,
  scenarioName: null,
  scenarioId: null,
  currentSessionId: null,
  deleteModal: { isOpen: false, sessionId: null },
  editingSessionId: null,
  editingTitle: '',
  setEditingTitle: vi.fn(),
  sessionSearchQuery: '',
  setSessionSearchQuery: vi.fn(),
  handleNewSession: vi.fn(),
  handleSelectSession: vi.fn(),
  handleDeleteSession: vi.fn(),
  confirmDeleteSession: vi.fn(),
  cancelDeleteSession: vi.fn(),
  handleStartEditTitle: vi.fn(),
  handleSaveTitle: vi.fn(),
  handleCancelEditTitle: vi.fn(),
  handleSend: vi.fn(),
  handleDeleteMessage: vi.fn(),
  loading: false,
};

describe('AskAiPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useAskAi).mockReturnValue({ ...mockUseAskAi });
    vi.mocked(useCopyToClipboard).mockReturnValue({ copiedId: null, copyToClipboard: vi.fn() });
  });

  it('メッセージが空の場合にEmptyStateが表示される', () => {
    render(<AskAiPage />);
    expect(screen.getByText('AIアシスタントへようこそ')).toBeInTheDocument();
  });

  it('新しいチャットボタンが表示される', () => {
    render(<AskAiPage />);
    expect(screen.getAllByText('新しいチャット').length).toBeGreaterThanOrEqual(1);
  });

  it('新しいチャットボタンクリックでhandleNewSessionが呼ばれる', () => {
    render(<AskAiPage />);
    fireEvent.click(screen.getAllByText('新しいチャット')[0]);
    expect(mockUseAskAi.handleNewSession).toHaveBeenCalled();
  });

  it('セッション一覧が表示される', () => {
    const sessionData = [
      { id: 1, title: 'セッション1', createdAt: '2026-02-10T10:00:00Z' },
      { id: 2, title: 'セッション2', createdAt: '2026-02-11T10:00:00Z' },
    ];
    vi.mocked(useAskAi).mockReturnValue({
      ...mockUseAskAi,
      sessions: sessionData as any,
      filteredSessions: sessionData as any,
    });
    render(<AskAiPage />);
    expect(screen.getAllByText('セッション1').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('セッション2').length).toBeGreaterThanOrEqual(1);
  });

  it('メッセージが表示される', () => {
    vi.mocked(useAskAi).mockReturnValue({
      ...mockUseAskAi,
      messages: [
        { id: 'msg-1', content: 'こんにちは', isSender: true, createdAt: 1707523200000 },
        { id: 'msg-2', content: 'お手伝いします', isSender: false, createdAt: 1707523200000 },
      ] as any,
    });
    render(<AskAiPage />);
    expect(screen.getByText('こんにちは')).toBeInTheDocument();
    expect(screen.getByText('お手伝いします')).toBeInTheDocument();
  });

  it('練習モード時に練習ヘッダーが表示される', () => {
    vi.mocked(useAskAi).mockReturnValue({
      ...mockUseAskAi,
      isPracticeMode: true,
      scenarioName: 'テストシナリオ',
    });
    render(<AskAiPage />);
    expect(screen.getByText('テストシナリオ')).toBeInTheDocument();
    expect(screen.getByText('AIが相手役を演じます')).toBeInTheDocument();
    expect(screen.getByText('練習終了')).toBeInTheDocument();
  });

  it('練習終了ボタンクリックでhandleSendが呼ばれる', () => {
    vi.mocked(useAskAi).mockReturnValue({
      ...mockUseAskAi,
      isPracticeMode: true,
      scenarioName: 'テストシナリオ',
    });
    render(<AskAiPage />);
    fireEvent.click(screen.getByText('練習終了'));
    expect(mockUseAskAi.handleSend).toHaveBeenCalledWith(
      expect.stringContaining('練習を終了')
    );
  });

  it('セッション削除確認モーダルが表示される', () => {
    vi.mocked(useAskAi).mockReturnValue({
      ...mockUseAskAi,
      deleteModal: { isOpen: true, sessionId: 1 },
    });
    render(<AskAiPage />);
    expect(screen.getByText('このセッションを削除しますか？チャット履歴もすべて削除されます。この操作は取り消せません。')).toBeInTheDocument();
  });

  it('削除確認でconfirmDeleteSessionが呼ばれる', () => {
    vi.mocked(useAskAi).mockReturnValue({
      ...mockUseAskAi,
      deleteModal: { isOpen: true, sessionId: 1 },
    });
    render(<AskAiPage />);
    fireEvent.click(screen.getByText('削除する'));
    expect(mockUseAskAi.confirmDeleteSession).toHaveBeenCalled();
  });

  it('削除キャンセルでcancelDeleteSessionが呼ばれる', () => {
    vi.mocked(useAskAi).mockReturnValue({
      ...mockUseAskAi,
      deleteModal: { isOpen: true, sessionId: 1 },
    });
    render(<AskAiPage />);
    fireEvent.click(screen.getByText('キャンセル'));
    expect(mockUseAskAi.cancelDeleteSession).toHaveBeenCalled();
  });

  it('モバイルヘッダーにセッション一覧を開くボタンが表示される', () => {
    render(<AskAiPage />);
    expect(screen.getByLabelText('セッション一覧を開く')).toBeInTheDocument();
  });

  it('スコアカードが表示される', () => {
    vi.mocked(useAskAi).mockReturnValue({
      ...mockUseAskAi,
      scoreCard: {
        overallScore: 8.5,
        scores: [
          { axis: '論理的構成力', score: 9, comment: '素晴らしい' },
        ],
      } as any,
    });
    render(<AskAiPage />);
    expect(screen.getByText('8.5')).toBeInTheDocument();
  });

  it('練習モードでないときは練習ヘッダーが表示されない', () => {
    render(<AskAiPage />);
    expect(screen.queryByText('AIが相手役を演じます')).not.toBeInTheDocument();
    expect(screen.queryByText('練習終了')).not.toBeInTheDocument();
  });

  it('セッション件数が表示される', () => {
    const sessionData = [
      { id: 1, title: 'セッション1', createdAt: '2026-02-10T10:00:00Z' },
      { id: 2, title: 'セッション2', createdAt: '2026-02-11T10:00:00Z' },
    ];
    vi.mocked(useAskAi).mockReturnValue({
      ...mockUseAskAi,
      sessions: sessionData as any,
      filteredSessions: sessionData as any,
    });
    render(<AskAiPage />);
    expect(screen.getAllByText('2件').length).toBeGreaterThanOrEqual(1);
  });

  it('セッション0件の場合、0件と表示される', () => {
    render(<AskAiPage />);
    expect(screen.getAllByText('0件').length).toBeGreaterThanOrEqual(1);
  });

  it('ローディング中はLoadingコンポーネントが表示される', () => {
    vi.mocked(useAskAi).mockReturnValue({
      ...mockUseAskAi,
      loading: true,
    });
    render(<AskAiPage />);
    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.queryByText('AIアシスタントへようこそ')).not.toBeInTheDocument();
  });
});
