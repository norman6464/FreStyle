import MessageBubbleAi from '../components/MessageBubbleAi';
import MessageInput from '../components/MessageInput';
import ConfirmModal from '../components/ConfirmModal';
import ScoreCardComponent from '../components/ScoreCard';
import PracticeResultSummary from '../components/PracticeResultSummary';
import SecondaryPanel from '../components/layout/SecondaryPanel';
import PracticeTimer from '../components/PracticeTimer';
import SessionNoteEditor from '../components/SessionNoteEditor';
import ExportSessionButton from '../components/ExportSessionButton';
import AiSessionListItem from '../components/AiSessionListItem';
import EmptyState from '../components/EmptyState';
import ConversationTemplates from '../components/ConversationTemplates';
import Loading from '../components/Loading';
import { PlusIcon, Bars3Icon, SparklesIcon, MagnifyingGlassIcon } from '@heroicons/react/24/outline';
import { useAskAi } from '../hooks/useAskAi';
import { useMobilePanelState } from '../hooks/useMobilePanelState';
import { useCopyToClipboard } from '../hooks/useCopyToClipboard';

export default function AskAiPage() {
  const { isOpen: mobilePanelOpen, open: openMobilePanel, close: closeMobilePanel } = useMobilePanelState();
  const { copiedId, copyToClipboard } = useCopyToClipboard();
  const {
    sessions,
    filteredSessions,
    messages,
    scoreCard,
    loading,
    messagesEndRef,
    isPracticeMode,
    scenarioId,
    scenarioName,
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
    handleDeleteMessage,
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
              onSelect={(id: number) => { handleSelectSession(id); closeMobilePanel(); }}
              onStartEdit={handleStartEditTitle}
              onDelete={handleDeleteSession}
              onSaveTitle={handleSaveTitle}
              onCancelEdit={handleCancelEditTitle}
              onEditingTitleChange={setEditingTitle}
            />
          ))}
        </div>
      </SecondaryPanel>

      {/* メインコンテンツ */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* モバイルヘッダー */}
        <div className="md:hidden bg-surface-1 border-b border-surface-3 px-4 py-2 flex items-center">
          <button
            onClick={openMobilePanel}
            className="p-1.5 hover:bg-surface-2 rounded transition-colors"
            aria-label="セッション一覧を開く"
          >
            <Bars3Icon className="w-5 h-5 text-[var(--color-text-muted)]" />
          </button>
          <span className="ml-2 text-xs text-[var(--color-text-muted)]">セッション一覧</span>
        </div>

        {/* 練習モードヘッダー */}
        {isPracticeMode && (
          <div className="bg-surface-1 border-b border-surface-3 px-4 py-3 flex items-center justify-between">
            <div>
              <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">
                {scenarioName || '練習モード'}
              </h2>
              <p className="text-xs text-[var(--color-text-muted)]">AIが相手役を演じます</p>
            </div>
            <div className="flex items-center gap-3">
              <PracticeTimer />
              <button
                onClick={() => handleSend('練習を終了して、今回の会話全体に対するフィードバックとスコアカードをお願いします。')}
                className="bg-rose-500 text-white text-xs px-3 py-1.5 rounded-lg hover:bg-rose-600 transition-colors"
              >
                練習終了
              </button>
            </div>
          </div>
        )}

        {/* メッセージエリア */}
        <div className="flex-1 overflow-y-auto px-4 py-4 space-y-3">
          {messages.length === 0 && (
            <div className="flex flex-col items-center justify-center h-full gap-6">
              <EmptyState
                icon={SparklesIcon}
                title="AIアシスタントへようこそ"
                description="質問や相談を何でも聞いてください"
              />
              <ConversationTemplates onSelect={handleSend} />
            </div>
          )}
          {messages.map((msg) => (
            <div key={msg.id} className="max-w-3xl mx-auto w-full">
              <MessageBubbleAi
                {...msg}
                isSender={msg.role === 'user'}
                type={msg.role === 'user' ? 'text' : 'bot'}
                onDelete={handleDeleteMessage}
                onCopy={copyToClipboard}
                isCopied={copiedId === msg.id}
              />
            </div>
          ))}

          {scoreCard && (
            <div className="max-w-3xl mx-auto w-full space-y-3">
              <ScoreCardComponent scoreCard={scoreCard} />
              {isPracticeMode && (
                <PracticeResultSummary scoreCard={scoreCard} scenarioName={scenarioName || '練習'} scenarioId={scenarioId ?? undefined} />
              )}
              {currentSessionId && (
                <SessionNoteEditor sessionId={currentSessionId} />
              )}
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* 入力欄 */}
        <div className="bg-surface-1 border-t border-surface-3 p-4">
          <div className="max-w-3xl mx-auto w-full flex items-end gap-2">
            <div className="flex-1">
              <MessageInput onSend={handleSend} />
            </div>
            <ExportSessionButton messages={messages} />
          </div>
        </div>
      </div>

      {/* セッション削除確認モーダル */}
      <ConfirmModal
        isOpen={deleteModal.isOpen}
        title="セッションを削除"
        message="このセッションを削除しますか？チャット履歴もすべて削除されます。この操作は取り消せません。"
        confirmText="削除する"
        cancelText="キャンセル"
        onConfirm={confirmDeleteSession}
        onCancel={cancelDeleteSession}
        isDanger={true}
      />
    </div>
  );
}
