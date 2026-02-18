import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ChatPage from '../ChatPage';
import { useChat } from '../../hooks/useChat';
import { useCopyToClipboard } from '../../hooks/useCopyToClipboard';

vi.mock('../../hooks/useChat');
vi.mock('../../hooks/useCopyToClipboard');

const mockUseChat = {
  messages: [],
  deleteModal: { isOpen: false, messageId: null },
  selectionMode: false,
  selectedMessages: new Set<string>(),
  showSceneSelector: false,
  showRephraseModal: false,
  rephraseResult: null,
  rephraseOriginalText: '',
  messagesEndRef: { current: null },
  handleSend: vi.fn(),
  handleDeleteMessage: vi.fn(),
  confirmDelete: vi.fn(),
  cancelDelete: vi.fn(),
  handleAiFeedback: vi.fn(),
  handleRangeClick: vi.fn(),
  handleQuickSelect: vi.fn(),
  handleSelectAll: vi.fn(),
  handleDeselectAll: vi.fn(),
  handleCancelSelection: vi.fn(),
  handleSendToAi: vi.fn(),
  handleSceneSelect: vi.fn(),
  handleRephrase: vi.fn(),
  setShowRephraseModal: vi.fn(),
  isInRange: vi.fn().mockReturnValue(false),
  getRangeLabel: vi.fn().mockReturnValue(null),
  loading: false,
};

describe('ChatPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useChat).mockReturnValue({ ...mockUseChat });
    vi.mocked(useCopyToClipboard).mockReturnValue({ copiedId: null, copyToClipboard: vi.fn() });
  });

  it('メッセージが空の場合にEmptyStateが表示される', () => {
    render(<ChatPage />);
    expect(screen.getByText('チャットへようこそ')).toBeInTheDocument();
  });

  it('メッセージが表示される', () => {
    vi.mocked(useChat).mockReturnValue({
      ...mockUseChat,
      messages: [
        { id: 'msg-1', roomId: 1, senderId: 1, senderName: 'テスト', content: 'メッセージ1', createdAt: 1707523200000, isSender: true },
        { id: 'msg-2', roomId: 1, senderId: 2, senderName: '相手', content: 'メッセージ2', createdAt: 1707523200000, isSender: false },
      ] as any,
    });
    render(<ChatPage />);
    expect(screen.getByText('メッセージ1')).toBeInTheDocument();
    expect(screen.getByText('メッセージ2')).toBeInTheDocument();
  });

  it('メッセージがある場合にAIフィードバックボタンが表示される', () => {
    vi.mocked(useChat).mockReturnValue({
      ...mockUseChat,
      messages: [
        { id: 'msg-1', roomId: 1, senderId: 1, senderName: 'テスト', content: 'テスト', createdAt: 1707523200000, isSender: true },
      ] as any,
    });
    render(<ChatPage />);
    expect(screen.getByText('AIにフィードバックしてもらう')).toBeInTheDocument();
  });

  it('AIフィードバックボタンクリックでhandleAiFeedbackが呼ばれる', () => {
    vi.mocked(useChat).mockReturnValue({
      ...mockUseChat,
      messages: [
        { id: 'msg-1', roomId: 1, senderId: 1, senderName: 'テスト', content: 'テスト', createdAt: 1707523200000, isSender: true },
      ] as any,
    });
    render(<ChatPage />);
    fireEvent.click(screen.getByText('AIにフィードバックしてもらう'));
    expect(mockUseChat.handleAiFeedback).toHaveBeenCalled();
  });

  it('選択モード時にMessageSelectionPanelが表示される', () => {
    vi.mocked(useChat).mockReturnValue({
      ...mockUseChat,
      selectionMode: true,
      selectedMessages: new Set(['msg-1']),
    });
    render(<ChatPage />);
    expect(screen.queryByText('AIにフィードバックしてもらう')).not.toBeInTheDocument();
  });

  it('メッセージが空の場合にAIフィードバックボタンが非表示', () => {
    render(<ChatPage />);
    expect(screen.queryByText('AIにフィードバックしてもらう')).not.toBeInTheDocument();
  });

  it('削除確認モーダルが表示される', () => {
    vi.mocked(useChat).mockReturnValue({
      ...mockUseChat,
      deleteModal: { isOpen: true, messageId: 'msg-1' },
    });
    render(<ChatPage />);
    expect(screen.getByText('このメッセージを削除しますか？この操作は取り消せません。')).toBeInTheDocument();
  });

  it('削除確認でconfirmDeleteが呼ばれる', () => {
    vi.mocked(useChat).mockReturnValue({
      ...mockUseChat,
      deleteModal: { isOpen: true, messageId: 'msg-1' },
    });
    render(<ChatPage />);
    fireEvent.click(screen.getByText('削除する'));
    expect(mockUseChat.confirmDelete).toHaveBeenCalled();
  });

  it('削除キャンセルでcancelDeleteが呼ばれる', () => {
    vi.mocked(useChat).mockReturnValue({
      ...mockUseChat,
      deleteModal: { isOpen: true, messageId: 'msg-1' },
    });
    render(<ChatPage />);
    fireEvent.click(screen.getByText('キャンセル'));
    expect(mockUseChat.cancelDelete).toHaveBeenCalled();
  });

  it('削除モーダルが閉じている場合はメッセージが表示されない', () => {
    render(<ChatPage />);
    expect(screen.queryByText('このメッセージを削除しますか？この操作は取り消せません。')).not.toBeInTheDocument();
  });

  it('ローディング中はLoadingコンポーネントが表示される', () => {
    vi.mocked(useChat).mockReturnValue({
      ...mockUseChat,
      loading: true,
    });
    render(<ChatPage />);
    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.queryByText('チャットへようこそ')).not.toBeInTheDocument();
  });
});
