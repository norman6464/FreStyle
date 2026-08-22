import { forwardRef, useEffect, useImperativeHandle, useState } from 'react';
import type { EditorCommand } from './editorCommands';

export interface SlashMenuListProps {
  items: EditorCommand[];
  /** 項目確定時に呼ばれる（Enter / クリック）。 */
  onSelect: (item: EditorCommand) => void;
}

/** SlashMenuListHandle は Suggestion の onKeyDown からキー操作を流し込むためのハンドル。 */
export interface SlashMenuListHandle {
  onKeyDown: (event: KeyboardEvent) => boolean;
}

/**
 * SlashMenuList は '/' メニューの候補リスト（presentational）。
 * 表示はレジストリの日本語ラベル＋英語トリガ（id）。矢印キーで選択、Enter で確定する。
 * 位置決め・表示切替は呼び出し側（Suggestion の mount）が担う。
 */
const SlashMenuList = forwardRef<SlashMenuListHandle, SlashMenuListProps>(
  function SlashMenuList({ items, onSelect }, ref) {
    const [selectedIndex, setSelectedIndex] = useState(0);

    // 絞り込みで件数が変わったら先頭へ戻す（範囲外選択を防ぐ）。
    useEffect(() => {
      setSelectedIndex(0);
    }, [items]);

    useImperativeHandle(ref, () => ({
      onKeyDown: (event: KeyboardEvent): boolean => {
        if (items.length === 0) return false;
        if (event.key === 'ArrowDown') {
          setSelectedIndex((prev) => (prev + 1) % items.length);
          return true;
        }
        if (event.key === 'ArrowUp') {
          setSelectedIndex((prev) => (prev - 1 + items.length) % items.length);
          return true;
        }
        if (event.key === 'Enter') {
          const item = items[selectedIndex];
          if (item) onSelect(item);
          return true;
        }
        return false;
      },
    }));

    if (items.length === 0) {
      return <p className="rte-slash-empty">該当するコマンドがありません</p>;
    }

    return (
      <ul role="listbox" aria-label="ブロックの挿入" className="rte-slash-list">
        {items.map((item, index) => (
          <li key={item.id} role="option" aria-selected={index === selectedIndex}>
            <button
              type="button"
              className={`rte-slash-item ${index === selectedIndex ? 'is-active' : ''}`}
              // mousedown での blur によりメニューが先に閉じないようにする。
              onMouseDown={(mouseEvent) => mouseEvent.preventDefault()}
              onClick={() => onSelect(item)}
              onMouseEnter={() => setSelectedIndex(index)}
            >
              <span className="rte-slash-glyph" aria-hidden="true">
                {item.glyph}
              </span>
              <span className="rte-slash-label">{item.label}</span>
              <span className="rte-slash-trigger">/{item.id.toLowerCase()}</span>
            </button>
          </li>
        ))}
      </ul>
    );
  },
);

export default SlashMenuList;
