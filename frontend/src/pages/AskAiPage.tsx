import { useState } from 'react';
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
import { useAskAi } from '../hooks/useAskAi';

export default function AskAiPage() {
  const [mobilePanelOpen, setMobilePanelOpen] = useState(false);
  const {
    sessions,
    messages,
    scoreCard,
    messagesEndRef,
    isPracticeMode,
    scenarioName,
    currentSessionId,
    deleteModal,
    editingSessionId,
    editingTitle,
    setEditingTitle,
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

  return (
    <div className="flex h-full">
      {/* セカンダリパネル: セッション一覧 */}
      <SecondaryPanel
        title="セッション"
        mobileOpen={mobilePanelOpen}
        onMobileClose={() => setMobilePanelOpen(false)}
        headerContent={
          <button
            onClick={handleNewSession}
            className="w-full bg-primary-500 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-primary-600 transition-colors flex items-center justify-center gap-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            新しいチャット
          </button>
        }
      >
        <div className="p-2 space-y-0.5">
          {sessions.map((session) => (
            <AiSessionListItem
              key={session.id}
              id={session.id}
              title={session.title}
              createdAt={session.createdAt}
              isActive={currentSessionId === session.id}
              isEditing={editingSessionId === session.id}
              editingTitle={editingTitle}
              onSelect={(id: number) => { handleSelectSession(id); setMobilePanelOpen(false); }}
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
            onClick={() => setMobilePanelOpen(true)}
            className="p-1.5 hover:bg-surface-2 rounded transition-colors"
            aria-label="セッション一覧を開く"
          >
            <svg className="w-5 h-5 text-[var(--color-text-muted)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
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
                className="bg-rose-900/300 text-white text-xs px-3 py-1.5 rounded-lg hover:bg-rose-600 transition-colors"
              >
                練習終了
              </button>
            </div>
          </div>
        )}

        {/* メッセージエリア */}
        <div className="flex-1 overflow-y-auto px-4 py-4 space-y-3">
          {messages.length === 0 && (
            <div className="flex flex-col items-center justify-center h-full text-center">
              <div className="bg-surface-3 rounded-full p-4 mb-4">
                <svg className="w-8 h-8 text-[var(--color-text-faint)]" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                </svg>
              </div>
              <h3 className="text-base font-semibold text-[var(--color-text-secondary)] mb-1">
                AIアシスタントへようこそ
              </h3>
              <p className="text-sm text-[var(--color-text-muted)] max-w-xs">
                質問や相談を何でも聞いてください
              </p>
            </div>
          )}
          {messages.map((msg) => (
            <div key={msg.id} className="max-w-3xl mx-auto w-full">
              <MessageBubbleAi
                {...msg}
                type={msg.isSender ? 'text' : 'bot'}
                onDelete={handleDeleteMessage}
              />
            </div>
          ))}

          {scoreCard && (
            <div className="max-w-3xl mx-auto w-full space-y-3">
              <ScoreCardComponent scoreCard={scoreCard} />
              {isPracticeMode && (
                <PracticeResultSummary scoreCard={scoreCard} scenarioName={scenarioName || '練習'} />
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
