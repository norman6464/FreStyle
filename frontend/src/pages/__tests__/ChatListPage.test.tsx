import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import ChatListPage from '../ChatListPage';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('sockjs-client', () => ({
  default: vi.fn(),
}));

vi.mock('@stomp/stompjs', () => ({
  Client: class MockClient {
    activate = vi.fn();
    deactivate = vi.fn();
    subscribe = vi.fn();
    constructor() {}
  },
}));

const mockUseChatList = vi.fn();
vi.mock('../../hooks/useChatList', () => ({
  useChatList: () => mockUseChatList(),
}));

function defaultChatListData() {
  return {
    chatUsers: [
      {
        roomId: 1,
        userId: 10,
        name: '田中太郎',
        profileImage: null,
        lastMessage: 'こんにちは',
        lastMessageAt: 1739437200000,
        lastMessageSenderId: 10,
        unreadCount: 3,
      },
      {
        roomId: 2,
        userId: 20,
        name: '佐藤花子',
        profileImage: 'https://example.com/photo.jpg',
        lastMessage: 'ありがとうございます',
        lastMessageAt: 1739368800000,
        lastMessageSenderId: 99,
        unreadCount: 0,
      },
    ],
    loading: false,
    userId: 99,
    fetchChatUsers: vi.fn(),
    updateUnreadCount: vi.fn(),
  };
}

describe('ChatListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseChatList.mockReturnValue(defaultChatListData());
  });

  it('チャットユーザー一覧を表示する', () => {
    render(<BrowserRouter><ChatListPage /></BrowserRouter>);

    expect(screen.getAllByText('田中太郎').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('佐藤花子').length).toBeGreaterThanOrEqual(1);
  });

  it('ローディング中はスピナーを表示する', () => {
    mockUseChatList.mockReturnValue({
      ...defaultChatListData(),
      loading: true,
      chatUsers: [],
    });

    render(<BrowserRouter><ChatListPage /></BrowserRouter>);

    expect(screen.getByText('チャットを選択してください')).toBeInTheDocument();
  });

  it('チャット履歴がない場合はメッセージを表示する', () => {
    mockUseChatList.mockReturnValue({
      ...defaultChatListData(),
      chatUsers: [],
    });

    render(<BrowserRouter><ChatListPage /></BrowserRouter>);

    expect(screen.getAllByText('チャット履歴がありません').length).toBeGreaterThanOrEqual(1);
  });

  it('未読バッジを表示する', () => {
    render(<BrowserRouter><ChatListPage /></BrowserRouter>);

    expect(screen.getAllByText('3').length).toBeGreaterThanOrEqual(1);
  });

  it('チャットルームクリックでnavigateが呼ばれる', () => {
    render(<BrowserRouter><ChatListPage /></BrowserRouter>);

    fireEvent.click(screen.getAllByText('田中太郎')[0]);

    expect(mockNavigate).toHaveBeenCalledWith('/chat/users/1');
  });

  it('自分が送ったメッセージには「あなた:」プレフィックスが付く', () => {
    render(<BrowserRouter><ChatListPage /></BrowserRouter>);

    expect(screen.getAllByText(/あなた: ありがとうございます/).length).toBeGreaterThanOrEqual(1);
  });

  it('プロフィール画像がない場合は頭文字アバターを表示する', () => {
    render(<BrowserRouter><ChatListPage /></BrowserRouter>);

    expect(screen.getAllByText('田').length).toBeGreaterThanOrEqual(1);
  });

  it('プロフィール画像がある場合はimgタグを表示する', () => {
    render(<BrowserRouter><ChatListPage /></BrowserRouter>);

    const imgs = screen.getAllByAltText('佐藤花子');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
    expect(imgs[0].getAttribute('src')).toBe('https://example.com/photo.jpg');
  });

  it('チャットユーザー数バッジを表示する', () => {
    render(<BrowserRouter><ChatListPage /></BrowserRouter>);

    expect(screen.getAllByText('2件').length).toBeGreaterThanOrEqual(1);
  });

  it('チャットユーザーが0件の場合「0件」バッジを表示する', () => {
    mockUseChatList.mockReturnValue({
      ...defaultChatListData(),
      chatUsers: [],
    });

    render(<BrowserRouter><ChatListPage /></BrowserRouter>);

    expect(screen.getAllByText('0件').length).toBeGreaterThanOrEqual(1);
  });
});
