interface MessageSelectionPanelProps {
  selectedCount: number;
  onQuickSelect: (n: number) => void;
  onSelectAll: () => void;
  onDeselectAll: () => void;
  onCancel: () => void;
  onSend: () => void;
}

const QUICK_SELECT_OPTIONS = [5, 10, 20];

export default function MessageSelectionPanel({
  selectedCount,
  onQuickSelect,
  onSelectAll,
  onDeselectAll,
  onCancel,
  onSend,
}: MessageSelectionPanelProps) {
  return (
    <div className="space-y-3">
      <div className="bg-surface-2 border border-[var(--color-border-hover)] rounded-lg p-3">
        <p className="text-sm text-primary-300">
          {selectedCount > 0
            ? `${selectedCount}件のメッセージを選択しました`
            : '開始位置のメッセージをタップしてください'
          }
        </p>
      </div>

      <div className="flex gap-2 flex-wrap">
        <span className="text-xs text-[var(--color-text-muted)] self-center">クイック選択:</span>
        {QUICK_SELECT_OPTIONS.map((n) => (
          <button
            key={n}
            onClick={() => onQuickSelect(n)}
            className="px-3 py-1 text-xs bg-surface-3 hover:bg-surface-3 text-[var(--color-text-secondary)] rounded-full transition-colors"
          >
            直近{n}件
          </button>
        ))}
        <button
          onClick={onSelectAll}
          className="px-3 py-1 text-xs bg-surface-3 hover:bg-surface-3 text-[var(--color-text-secondary)] rounded-full transition-colors"
        >
          すべて
        </button>
        <button
          onClick={onDeselectAll}
          className="px-3 py-1 text-xs text-rose-500 hover:bg-rose-900/30 rounded-full transition-colors"
        >
          リセット
        </button>
      </div>

      <div className="flex gap-2">
        <button
          onClick={onCancel}
          className="flex-1 bg-surface-3 hover:bg-surface-3 text-[var(--color-text-secondary)] font-medium py-2.5 px-4 rounded-lg text-sm transition-colors"
        >
          キャンセル
        </button>
        <button
          onClick={onSend}
          disabled={selectedCount === 0}
          className={`flex-1 font-medium py-2.5 px-4 rounded-lg text-sm transition-colors ${
            selectedCount > 0
              ? 'bg-primary-500 hover:bg-primary-600 text-white'
              : 'bg-surface-3 text-[var(--color-text-muted)] cursor-not-allowed'
          }`}
        >
          {selectedCount > 0
            ? `${selectedCount}件をAIに送信`
            : '範囲を選択してください'
          }
        </button>
      </div>
    </div>
  );
}
