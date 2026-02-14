import { XMarkIcon } from '@heroicons/react/24/solid';

interface MessageSearchBarProps {
  onSearch: (query: string) => void;
  onClear: () => void;
  matchCount: number;
}

export default function MessageSearchBar({ onSearch, onClear, matchCount }: MessageSearchBarProps) {
  return (
    <div className="flex items-center gap-2 px-4 py-2 bg-surface-2 border-b border-surface-3" role="search">
      <input
        type="text"
        placeholder="メッセージを検索..."
        aria-label="メッセージを検索"
        onChange={(e) => onSearch(e.target.value)}
        className="flex-1 text-sm bg-surface-1 border border-surface-3 rounded-lg px-3 py-1.5 outline-none focus:ring-1 focus:ring-primary-400 focus:border-primary-400"
      />
      {matchCount > 0 && (
        <span className="text-xs text-[var(--color-text-muted)] flex-shrink-0" aria-live="polite">{matchCount}件</span>
      )}
      <button
        onClick={onClear}
        aria-label="検索をクリア"
        className="text-xs text-[var(--color-text-faint)] hover:text-[var(--color-text-tertiary)] p-1 rounded hover:bg-surface-3 transition-colors"
      >
        <XMarkIcon className="w-4 h-4" />
      </button>
    </div>
  );
}
