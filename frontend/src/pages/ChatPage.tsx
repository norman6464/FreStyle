import MessageBubble from '../components/MessageBubble';
import MessageInput from '../components/MessageInput';
import ConfirmModal from '../components/ConfirmModal';
import SceneSelector from '../components/SceneSelector';
import RephraseModal from '../components/RephraseModal';
import MessageSelectionPanel from '../components/MessageSelectionPanel';
import EmptyState from '../components/EmptyState';
import Loading from '../components/Loading';
import { ChatBubbleLeftRightIcon, CheckIcon } from '@heroicons/react/24/outline';
import { useChat } from '../hooks/useChat';
import { useCopyToClipboard } from '../hooks/useCopyToClipboard';

export default function ChatPage() {
  const {
    messages,
    loading,
    deleteModal,
    selectionMode,
    selectedMessages,
    showSceneSelector,
    showRephraseModal,
    rephraseResult,
    rephraseOriginalText,
    messagesEndRef,
    handleSend,
    handleDeleteMessage,
    confirmDelete,
    cancelDelete,
    handleAiFeedback,
    handleRangeClick,
    handleQuickSelect,
    handleSelectAll,
    handleDeselectAll,
    handleCancelSelection,
    handleSendToAi,
    handleSceneSelect,
    handleRephrase,
    setShowRephraseModal,
    isInRange,
    getRangeLabel,
  } = useChat();

  const { copiedId, copyToClipboard } = useCopyToClipboard();

  if (loading) {
    return <Loading message="読み込み中..." className="min-h-[50vh]" />;
  }

  return (
    <div className="flex flex-col h-full">
      {/* メッセージエリア */}
      <div className="flex-1 overflow-y-auto px-4 py-4 space-y-3">
        {messages.length === 0 && (
          <EmptyState
            icon={ChatBubbleLeftRightIcon}
            title="チャットへようこそ"
            description="相手とのチャットをここで行えます"
          />
        )}
        {messages.map((msg, index) => (
          <div key={msg.id} className="flex items-start gap-2 max-w-3xl mx-auto w-full">
            {selectionMode && (
              <div className="flex-shrink-0 flex flex-col items-center">
                {getRangeLabel(index) && (
                  <span className={`text-xs font-bold mb-1 px-2 py-0.5 rounded ${
                    getRangeLabel(index) === '開始'
                      ? 'bg-green-900/30 text-green-400'
                      : 'bg-red-900/30 text-red-700'
                  }`}>
                    {getRangeLabel(index)}
                  </span>
                )}
                <button
                  onClick={() => handleRangeClick(msg.id)}
                  className={`mt-1 w-7 h-7 rounded-full border-2 flex items-center justify-center transition-all ${
                    isInRange(index)
                      ? 'bg-primary-500 border-primary-500 text-white'
                      : 'border-surface-3 hover:border-primary-400 hover:bg-surface-2'
                  }`}
                >
                  {isInRange(index) ? (
                    <CheckIcon className="w-3.5 h-3.5" />
                  ) : (
                    <span className="text-[10px] text-[var(--color-text-faint)]">{index + 1}</span>
                  )}
                </button>
              </div>
            )}
            <div className={`flex-1 transition-all ${
              selectionMode && isInRange(index) ? 'bg-surface-2 -mx-2 px-2 py-1 rounded-lg' : ''
            }`}>
              <MessageBubble
                {...msg}
                onDelete={selectionMode ? null : handleDeleteMessage}
                onRephrase={selectionMode ? null : handleRephrase}
                onCopy={selectionMode ? null : copyToClipboard}
                isCopied={copiedId === msg.id}
              />
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* 入力欄 */}
      <div className="bg-surface-1 border-t border-surface-3 p-4">
        <div className="max-w-3xl mx-auto w-full space-y-3">
          {selectionMode ? (
            <MessageSelectionPanel
              selectedCount={selectedMessages.size}
              onQuickSelect={handleQuickSelect}
              onSelectAll={handleSelectAll}
              onDeselectAll={handleDeselectAll}
              onCancel={handleCancelSelection}
              onSend={handleSendToAi}
            />
          ) : (
            <>
              {messages.length > 0 && (
                <button
                  onClick={handleAiFeedback}
                  className="w-full bg-primary-500 hover:bg-primary-600 text-white font-medium py-2.5 px-4 rounded-lg text-sm transition-colors"
                >
                  AIにフィードバックしてもらう
                </button>
              )}
              <MessageInput onSend={handleSend} />
            </>
          )}
        </div>
      </div>

      {showSceneSelector && (
        <SceneSelector
          onSelect={(sceneId) => handleSceneSelect(sceneId)}
          onCancel={() => handleSceneSelect(null)}
        />
      )}

      {showRephraseModal && (
        <RephraseModal
          result={rephraseResult}
          onClose={() => setShowRephraseModal(false)}
          originalText={rephraseOriginalText}
        />
      )}

      <ConfirmModal
        isOpen={deleteModal.isOpen}
        title="メッセージを削除"
        message="このメッセージを削除しますか？この操作は取り消せません。"
        confirmText="削除する"
        cancelText="キャンセル"
        onConfirm={confirmDelete}
        onCancel={cancelDelete}
        isDanger={true}
      />
    </div>
  );
}
