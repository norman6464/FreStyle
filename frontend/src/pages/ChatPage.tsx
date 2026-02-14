import MessageBubble from '../components/MessageBubble';
import MessageInput from '../components/MessageInput';
import ConfirmModal from '../components/ConfirmModal';
import SceneSelector from '../components/SceneSelector';
import RephraseModal from '../components/RephraseModal';
import { useChat } from '../hooks/useChat';

export default function ChatPage() {
  const {
    messages,
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

  return (
    <div className="flex flex-col h-full">
      {/* メッセージエリア */}
      <div className="flex-1 overflow-y-auto px-4 py-4 space-y-3">
        {messages.length === 0 && (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <div className="bg-surface-3 rounded-full p-4 mb-4">
              <svg className="w-8 h-8 text-[var(--color-text-faint)]" fill="currentColor" viewBox="0 0 20 20">
                <path d="M2 5a2 2 0 012-2h12a2 2 0 012 2v10a2 2 0 01-2 2H4a2 2 0 01-2-2V5z" />
              </svg>
            </div>
            <h3 className="text-base font-semibold text-[var(--color-text-secondary)] mb-1">
              チャットへようこそ
            </h3>
            <p className="text-sm text-[var(--color-text-muted)]">
              相手とのチャットをここで行えます
            </p>
          </div>
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
                    <svg className="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
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
            <div className="space-y-3">
              <div className="bg-surface-2 border border-[var(--color-border-hover)] rounded-lg p-3">
                <p className="text-sm text-primary-300">
                  {selectedMessages.size > 0
                    ? `${selectedMessages.size}件のメッセージを選択しました`
                    : '開始位置のメッセージをタップしてください'
                  }
                </p>
              </div>

              <div className="flex gap-2 flex-wrap">
                <span className="text-xs text-[var(--color-text-muted)] self-center">クイック選択:</span>
                {[5, 10, 20].map((n) => (
                  <button
                    key={n}
                    onClick={() => handleQuickSelect(n)}
                    className="px-3 py-1 text-xs bg-surface-3 hover:bg-surface-3 text-[var(--color-text-secondary)] rounded-full transition-colors"
                  >
                    直近{n}件
                  </button>
                ))}
                <button
                  onClick={handleSelectAll}
                  className="px-3 py-1 text-xs bg-surface-3 hover:bg-surface-3 text-[var(--color-text-secondary)] rounded-full transition-colors"
                >
                  すべて
                </button>
                <button
                  onClick={handleDeselectAll}
                  className="px-3 py-1 text-xs text-rose-500 hover:bg-rose-900/30 rounded-full transition-colors"
                >
                  リセット
                </button>
              </div>

              <div className="flex gap-2">
                <button
                  onClick={handleCancelSelection}
                  className="flex-1 bg-surface-3 hover:bg-surface-3 text-[var(--color-text-secondary)] font-medium py-2.5 px-4 rounded-lg text-sm transition-colors"
                >
                  キャンセル
                </button>
                <button
                  onClick={handleSendToAi}
                  disabled={selectedMessages.size === 0}
                  className={`flex-1 font-medium py-2.5 px-4 rounded-lg text-sm transition-colors ${
                    selectedMessages.size > 0
                      ? 'bg-primary-500 hover:bg-primary-600 text-white'
                      : 'bg-surface-3 text-[var(--color-text-muted)] cursor-not-allowed'
                  }`}
                >
                  {selectedMessages.size > 0
                    ? `${selectedMessages.size}件をAIに送信`
                    : '範囲を選択してください'
                  }
                </button>
              </div>
            </div>
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
