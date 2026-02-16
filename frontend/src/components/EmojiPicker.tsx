import { useState, useEffect, useRef, useMemo } from 'react';
import { EMOJI_CATEGORIES } from '../constants/emojiData';

interface EmojiPickerProps {
  isOpen: boolean;
  onSelect: (emoji: string) => void;
  onClose: () => void;
}

export default function EmojiPicker({ isOpen, onSelect, onClose }: EmojiPickerProps) {
  const [search, setSearch] = useState('');
  const [activeCategory, setActiveCategory] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const gridRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (isOpen) {
      setSearch('');
      setActiveCategory(0);
      setTimeout(() => inputRef.current?.focus(), 0);
    }
  }, [isOpen]);

  const filteredCategories = useMemo(() => {
    if (!search.trim()) return EMOJI_CATEGORIES;
    const q = search.toLowerCase();
    return EMOJI_CATEGORIES.map(cat => ({
      ...cat,
      emojis: cat.emojis.filter(e => e.includes(q)),
    })).filter(cat => cat.emojis.length > 0);
  }, [search]);

  const handleSelect = (emoji: string) => {
    onSelect(emoji);
    onClose();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      e.preventDefault();
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div
      data-testid="emoji-picker"
      className="w-72 bg-[var(--color-surface-1)] border border-[var(--color-border)] rounded-xl shadow-2xl overflow-hidden"
    >
      {/* 検索 */}
      <div className="px-3 pt-3 pb-2">
        <input
          ref={inputRef}
          type="text"
          placeholder="絵文字を検索..."
          value={search}
          onChange={e => setSearch(e.target.value)}
          onKeyDown={handleKeyDown}
          className="w-full px-3 py-1.5 text-xs bg-[var(--color-surface-2)] border border-[var(--color-border)] rounded-lg text-[var(--color-text-primary)] placeholder-[var(--color-text-muted)] outline-none focus:border-[var(--color-accent)]"
        />
      </div>

      {/* カテゴリタブ */}
      {!search && (
        <div className="flex items-center gap-0.5 px-2 pb-1 overflow-x-auto">
          {EMOJI_CATEGORIES.map((cat, i) => (
            <button
              key={cat.name}
              type="button"
              aria-label={cat.name}
              className={`flex-shrink-0 w-7 h-7 flex items-center justify-center rounded text-sm transition-colors ${
                activeCategory === i
                  ? 'bg-[var(--color-surface-3)]'
                  : 'hover:bg-[var(--color-surface-2)]'
              }`}
              onClick={() => {
                setActiveCategory(i);
                gridRef.current?.scrollTo(0, 0);
              }}
            >
              {cat.icon}
            </button>
          ))}
        </div>
      )}

      {/* 絵文字グリッド */}
      <div ref={gridRef} className="h-52 overflow-y-auto px-2 pb-2">
        {search ? (
          // 検索結果
          filteredCategories.length === 0 ? (
            <div className="flex items-center justify-center h-full text-xs text-[var(--color-text-muted)]">
              該当する絵文字がありません
            </div>
          ) : (
            filteredCategories.map(cat => (
              <div key={cat.name}>
                <div className="text-[10px] font-medium text-[var(--color-text-muted)] px-1 py-1">
                  {cat.name}
                </div>
                <div className="grid grid-cols-8 gap-0.5">
                  {cat.emojis.map((emoji, i) => (
                    <button
                      key={`${cat.name}-${i}`}
                      type="button"
                      className="w-8 h-8 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] text-lg transition-colors"
                      onClick={() => handleSelect(emoji)}
                    >
                      {emoji}
                    </button>
                  ))}
                </div>
              </div>
            ))
          )
        ) : (
          // カテゴリ表示
          <>
            <div className="text-[10px] font-medium text-[var(--color-text-muted)] px-1 py-1">
              {EMOJI_CATEGORIES[activeCategory]?.name}
            </div>
            <div className="grid grid-cols-8 gap-0.5">
              {EMOJI_CATEGORIES[activeCategory]?.emojis.map((emoji, i) => (
                <button
                  key={i}
                  type="button"
                  className="w-8 h-8 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] text-lg transition-colors"
                  onClick={() => handleSelect(emoji)}
                >
                  {emoji}
                </button>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
