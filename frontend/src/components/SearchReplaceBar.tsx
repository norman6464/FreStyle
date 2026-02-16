import { useState, useEffect, useRef, useCallback } from 'react';
import type { Editor } from '@tiptap/react';
import {
  MagnifyingGlassIcon,
  XMarkIcon,
  ChevronUpIcon,
  ChevronDownIcon,
  ArrowsRightLeftIcon,
} from '@heroicons/react/24/outline';

interface SearchReplaceBarProps {
  editor: Editor | null;
  isOpen: boolean;
  onClose: () => void;
}

export default function SearchReplaceBar({ editor, isOpen, onClose }: SearchReplaceBarProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [replaceTerm, setReplaceTerm] = useState('');
  const [showReplace, setShowReplace] = useState(false);
  const searchInputRef = useRef<HTMLInputElement>(null);

  const resultCount = editor?.storage.searchReplace?.results?.length ?? 0;
  const currentIndex = editor?.storage.searchReplace?.currentIndex ?? 0;

  useEffect(() => {
    if (isOpen) {
      setSearchTerm('');
      setReplaceTerm('');
      setShowReplace(false);
      editor?.commands.clearSearch();
      setTimeout(() => searchInputRef.current?.focus(), 0);
    } else {
      editor?.commands.clearSearch();
    }
  }, [isOpen, editor]);

  const handleSearchChange = useCallback((value: string) => {
    setSearchTerm(value);
    editor?.commands.setSearchTerm(value);
  }, [editor]);

  const handleReplaceChange = useCallback((value: string) => {
    setReplaceTerm(value);
    editor?.commands.setReplaceTerm(value);
  }, [editor]);

  const handleFindNext = useCallback(() => {
    editor?.commands.findNext();
  }, [editor]);

  const handleFindPrev = useCallback(() => {
    editor?.commands.findPrev();
  }, [editor]);

  const handleReplaceCurrent = useCallback(() => {
    editor?.commands.replaceCurrent();
  }, [editor]);

  const handleReplaceAll = useCallback(() => {
    editor?.commands.replaceAll();
  }, [editor]);

  const handleClose = useCallback(() => {
    editor?.commands.clearSearch();
    onClose();
  }, [editor, onClose]);

  const handleSearchKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      e.preventDefault();
      handleClose();
      return;
    }
    if (e.key === 'Enter') {
      e.preventDefault();
      if (e.shiftKey) {
        handleFindPrev();
      } else {
        handleFindNext();
      }
    }
  }, [handleClose, handleFindNext, handleFindPrev]);

  const handleReplaceKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      e.preventDefault();
      handleClose();
      return;
    }
    if (e.key === 'Enter') {
      e.preventDefault();
      handleReplaceCurrent();
    }
  }, [handleClose, handleReplaceCurrent]);

  if (!isOpen) return null;

  return (
    <div
      className="absolute top-0 right-0 z-30 bg-[var(--color-surface-1)] border border-[var(--color-surface-3)] rounded-bl-lg shadow-lg"
      data-testid="search-replace-bar"
    >
      {/* 検索行 */}
      <div className="flex items-center gap-1.5 px-3 py-2">
        <button
          type="button"
          aria-label="置換を表示"
          className="p-0.5 rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-muted)]"
          onClick={() => setShowReplace(!showReplace)}
        >
          <ArrowsRightLeftIcon className="w-4 h-4" />
        </button>
        <div className="relative flex-1">
          <MagnifyingGlassIcon className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-[var(--color-text-muted)]" />
          <input
            ref={searchInputRef}
            type="text"
            placeholder="検索..."
            value={searchTerm}
            onChange={e => handleSearchChange(e.target.value)}
            onKeyDown={handleSearchKeyDown}
            className="w-48 pl-7 pr-2 py-1 text-xs bg-[var(--color-surface-2)] border border-[var(--color-surface-3)] rounded text-[var(--color-text-primary)] placeholder-[var(--color-text-muted)] outline-none focus:border-[var(--color-accent)]"
          />
        </div>
        <span className="text-[10px] text-[var(--color-text-muted)] min-w-[3.5rem] text-center">
          {searchTerm ? `${resultCount > 0 ? currentIndex + 1 : 0} / ${resultCount}` : ''}
        </span>
        <button
          type="button"
          aria-label="前を検索"
          className="p-1 rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-muted)] disabled:opacity-30"
          onClick={handleFindPrev}
          disabled={resultCount === 0}
        >
          <ChevronUpIcon className="w-3.5 h-3.5" />
        </button>
        <button
          type="button"
          aria-label="次を検索"
          className="p-1 rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-muted)] disabled:opacity-30"
          onClick={handleFindNext}
          disabled={resultCount === 0}
        >
          <ChevronDownIcon className="w-3.5 h-3.5" />
        </button>
        <button
          type="button"
          aria-label="検索を閉じる"
          className="p-1 rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-muted)]"
          onClick={handleClose}
        >
          <XMarkIcon className="w-3.5 h-3.5" />
        </button>
      </div>

      {/* 置換行 */}
      {showReplace && (
        <div className="flex items-center gap-1.5 px-3 py-2 border-t border-[var(--color-surface-3)]">
          <div className="w-5" /> {/* スペーサー（アイコン幅に合わせる） */}
          <input
            type="text"
            placeholder="置換..."
            value={replaceTerm}
            onChange={e => handleReplaceChange(e.target.value)}
            onKeyDown={handleReplaceKeyDown}
            className="w-48 px-2 py-1 text-xs bg-[var(--color-surface-2)] border border-[var(--color-surface-3)] rounded text-[var(--color-text-primary)] placeholder-[var(--color-text-muted)] outline-none focus:border-[var(--color-accent)]"
          />
          <button
            type="button"
            aria-label="置換"
            className="px-2 py-1 text-[10px] rounded bg-[var(--color-surface-3)] hover:bg-[var(--color-surface-2)] text-[var(--color-text-secondary)] disabled:opacity-30"
            onClick={handleReplaceCurrent}
            disabled={resultCount === 0}
          >
            置換
          </button>
          <button
            type="button"
            aria-label="全て置換"
            className="px-2 py-1 text-[10px] rounded bg-[var(--color-surface-3)] hover:bg-[var(--color-surface-2)] text-[var(--color-text-secondary)] disabled:opacity-30"
            onClick={handleReplaceAll}
            disabled={resultCount === 0}
          >
            全置換
          </button>
        </div>
      )}
    </div>
  );
}
