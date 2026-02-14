import { useNavigate } from 'react-router-dom';
import SecondaryPanel from '../components/layout/SecondaryPanel';
import EmptyState from '../components/EmptyState';
import SearchBox from '../components/SearchBox';
import { ChatBubbleLeftRightIcon } from '@heroicons/react/24/outline';
import Avatar from '../components/Avatar';
import Loading from '../components/Loading';
import { useChatList } from '../hooks/useChatList';
import { formatTime, truncateMessage } from '../utils/formatters';
import { useMobilePanelState } from '../hooks/useMobilePanelState';

export default function ChatListPage() {
  const { isOpen: mobilePanelOpen, open: openMobilePanel, close: closeMobilePanel } = useMobilePanelState();
  const navigate = useNavigate();
  const { chatUsers, loading, searchQuery, setSearchQuery } = useChatList();

  return (
    <div className="flex h-full">
      <SecondaryPanel
        title="チャット"
        mobileOpen={mobilePanelOpen}
        onMobileClose={closeMobilePanel}
        headerContent={
          <SearchBox
            value={searchQuery}
            onChange={setSearchQuery}
            placeholder="名前やメールで検索..."
          />
        }
      >
        {loading ? (
          <Loading className="py-8" />
        ) : chatUsers.length === 0 ? (
          <div className="p-4 text-center text-sm text-[var(--color-text-muted)]">
            チャット履歴がありません
          </div>
        ) : (
          chatUsers.map((user) => (
            <button
              key={user.roomId}
              onClick={() => { navigate(`/chat/users/${user.roomId}`); closeMobilePanel(); }}
              className="w-full px-4 py-3 flex items-center gap-3 hover:bg-surface-2 transition-colors border-b border-surface-3 text-left"
            >
              <Avatar name={user.name} src={user.profileImage} />

              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-semibold text-[var(--color-text-primary)] truncate">
                    {user.name || 'Unknown User'}
                  </span>
                  <span className="text-[11px] text-[var(--color-text-faint)] flex-shrink-0 ml-2">
                    {formatTime(user.lastMessageAt!)}
                  </span>
                </div>
                <div className="flex items-center justify-between mt-0.5">
                  <p className="text-xs text-[var(--color-text-muted)] truncate pr-2">
                    {user.lastMessageSenderId === user.userId
                      ? truncateMessage(user.lastMessage)
                      : `あなた: ${truncateMessage(user.lastMessage)}`}
                  </p>
                  {user.unreadCount > 0 && (
                    <span className="bg-rose-900/300 text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full flex-shrink-0">
                      {user.unreadCount > 99 ? '99+' : user.unreadCount}
                    </span>
                  )}
                </div>
              </div>
            </button>
          ))
        )}
      </SecondaryPanel>

      <div className="flex-1 flex flex-col">
        {/* モバイルヘッダー */}
        <div className="md:hidden bg-surface-1 border-b border-surface-3 px-4 py-2 flex items-center">
          <button
            onClick={openMobilePanel}
            className="p-1.5 hover:bg-surface-2 rounded transition-colors"
            aria-label="チャット一覧を開く"
          >
            <svg className="w-5 h-5 text-[var(--color-text-muted)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          <span className="ml-2 text-xs text-[var(--color-text-muted)]">チャット一覧</span>
        </div>
        <EmptyState
          icon={ChatBubbleLeftRightIcon}
          title="チャットを選択してください"
          description="左のリストからチャット相手を選択するか、新しいチャットを始めましょう。"
          action={{ label: 'ユーザーを検索', onClick: () => navigate('/chat/users') }}
        />
      </div>
    </div>
  );
}
