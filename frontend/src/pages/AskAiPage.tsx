import MessageBubbleAi from '../components/MessageBubbleAi';
import MessageInput from '../components/MessageInput';
import ConfirmModal from '../components/ConfirmModal';
import SecondaryPanel from '../components/layout/SecondaryPanel';
import AiSessionListItem from '../components/AiSessionListItem';
import EmptyState from '../components/EmptyState';
import Loading from '../components/Loading';
import {
  PlusIcon,
  Bars3Icon,
  SparklesIcon,
  MagnifyingGlassIcon,
} from '@heroicons/react/24/outline';
import { useAskAi } from '../hooks/useAskAi';
import { useMobilePanelState } from '../hooks/useMobilePanelState';
import { useCopyToClipboard } from '../hooks/useCopyToClipboard';

/**
 * 汎用 AI チャット画面。
 *
 * 旧版で並んでいた練習モード / スコアカード / シナリオ受け渡し / セッションノート /
 * 言い換え提案などはすべて削除し、純粋な「セッション一覧 + メッセージ表示 + 入力」だけを残す。
 *
 * Markdown 表示・カード装飾の除去・SSE ストリーミングは PR-B / PR-C で追加。
 */
export default function AskAiPage() {
  const { isOpen: mobilePanelOpen, open: openMobilePanel, close: closeMobilePanel } =
    useMobilePanelState();
  const { copiedId, copyToClipboard } = useCopyToClipboard();
  const {
    sessions,
    filteredSessions,
    messages,
    loading,
    messagesEndRef,
    currentSessionId,
    deleteModal,
    editingSessionId,
    editingTitle,
    setEditingTitle,
    sessionSearchQuery,
    setSessionSearchQuery,
    handleNewSession,
    handleSelectSession,
    handleDeleteSession,
    confirmDeleteSession,
    cancelDeleteSession,
    handleStartEditTitle,
    handleSaveTitle,
    handleCancelEditTitle,
    handleSend,
  } = useAskAi();

  if (loading && sessions.length === 0) {
    return <Loading message="読み込み中..." className="min-h-[calc(100vh-3.5rem)]" />;
  }

  return (
    <div className="flex h-full">
      {/* セカンダリパネル: セッション一覧 */}
      <SecondaryPanel
        title="セッション"
        badge={`${sessions.length}件`}
        mobileOpen={mobilePanelOpen}
        onMobileClose={closeMobilePanel}
        headerContent={
          <div className="space-y-2">
            <div className="relative">
              <MagnifyingGlassIcon className="w-4 h-4 absolute left-2.5 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" />
              <input
                type="text"
                placeholder="セッションを検索..."
                aria-label="セッションを検索"
                value={sessionSearchQuery}
                onChange={(e) => setSessionSearchQuery(e.target.value)}
                className="w-full pl-8 pr-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:outline-none focus:border-primary-500 transition-colors"
              />
            </div>
            <button
              onClick={handleNewSession}
              className="w-full bg-primary-500 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-primary-600 transition-colors flex items-center justify-center gap-2"
            >
              <PlusIcon className="w-4 h-4" />
              新しいチャット
            </button>
          </div>
        }
      >
        <div className="p-2 space-y-0.5">
          {filteredSessions.map((session) => (
            <AiSessionListItem
              key={session.id}
              id={session.id}
              title={session.title}
              createdAt={session.createdAt}
              isActive={currentSessionId === session.id}
              isEditing={editingSessionId === session.id}
              editingTitle={editingTitle}
              onSelect={(id: number) => {
                handleSelectSession(id);
                closeMobilePanel();
              }}
              onStartEdit={handleStartEditTitle}
              onDelete={handleDeleteSession}
              onSaveTitle={handleSaveTitle}
              onCancelEdit={handleCancelEditTitle}
              onEditingTitleChange={setEditingTitle}
            />
          ))}
          {filteredSessions.length === 0 && (
            <p className="text-center text-xs text-[var(--color-text-muted)] py-8">
              セッションがありません
            </p>
          )}
        </div>
      </SecondaryPanel>

      {/* メイン: チャット */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* モバイル用パネル開閉ボタン */}
        <div className="md:hidden p-2 border-b border-surface-3">
          <button
            onClick={openMobilePanel}
            className="p-2 rounded hover:bg-surface-2"
            aria-label="セッションを開く"
          >
            <Bars3Icon className="w-5 h-5" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-4 py-6">
          {messages.length === 0 ? (
            <EmptyState
              icon={SparklesIcon}
              title="質問してみましょう"
              description="自由に質問・要約・コードのレビュー依頼などができます。"
            />
          ) : (
            <div className="max-w-3xl mx-auto space-y-4">
              {messages.map((message) => (
                <MessageBubbleAi
                  key={message.id}
                  id={message.id}
                  type="text"
                  content={message.content}
                  attachments={message.attachments}
                  isSender={message.role === 'user'}
                  isCopied={copiedId === message.id}
                  onCopy={copyToClipboard}
                />
              ))}
              <div ref={messagesEndRef} />
            </div>
          )}
        </div>

        <div className="border-t border-surface-3 p-3 bg-[var(--color-surface-1)]">
          <div className="max-w-3xl mx-auto">
            <MessageInput onSend={handleSend} />
          </div>
        </div>
      </div>

      {/* 削除確認モーダル */}
      <ConfirmModal
        isOpen={deleteModal.isOpen}
        title="セッションを削除しますか？"
        message="この操作は取り消せません。"
        confirmText="削除"
        cancelText="キャンセル"
        onConfirm={confirmDeleteSession}
        onCancel={cancelDeleteSession}
      />
    </div>
  );
}
