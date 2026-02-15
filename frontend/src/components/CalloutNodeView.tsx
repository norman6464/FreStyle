import { useState, useRef, useEffect } from 'react';
import { NodeViewWrapper, NodeViewContent } from '@tiptap/react';
import type { CalloutType } from '../extensions/CalloutExtension';

const CALLOUT_STYLES: Record<CalloutType, { bg: string; border: string }> = {
  info: { bg: 'bg-blue-50 dark:bg-blue-950/30', border: 'border-blue-200 dark:border-blue-800' },
  warning: { bg: 'bg-amber-50 dark:bg-amber-950/30', border: 'border-amber-200 dark:border-amber-800' },
  error: { bg: 'bg-red-50 dark:bg-red-950/30', border: 'border-red-200 dark:border-red-800' },
  success: { bg: 'bg-green-50 dark:bg-green-950/30', border: 'border-green-200 dark:border-green-800' },
};

const CALLOUT_TYPES: { type: CalloutType; emoji: string; label: string }[] = [
  { type: 'info', emoji: '💡', label: '情報' },
  { type: 'warning', emoji: '⚠️', label: '警告' },
  { type: 'error', emoji: '🚫', label: 'エラー' },
  { type: 'success', emoji: '✅', label: '成功' },
];

interface CalloutNodeViewProps {
  node: { attrs: { type: CalloutType; emoji: string } };
  updateAttributes: (attrs: Partial<{ type: CalloutType; emoji: string }>) => void;
}

export default function CalloutNodeView({ node, updateAttributes }: CalloutNodeViewProps) {
  const [showMenu, setShowMenu] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const calloutType = node.attrs.type as CalloutType;
  const emoji = node.attrs.emoji;
  const style = CALLOUT_STYLES[calloutType] || CALLOUT_STYLES.info;

  useEffect(() => {
    if (!showMenu) return;
    const handleClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setShowMenu(false);
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [showMenu]);

  return (
    <NodeViewWrapper className={`callout my-2 rounded-lg border-l-4 p-3 ${style.bg} ${style.border}`}>
      <div className="flex items-start gap-2">
        <div className="relative" contentEditable={false}>
          <button
            type="button"
            className="text-lg leading-none hover:scale-110 transition-transform cursor-pointer"
            aria-label="コールアウトタイプを変更"
            onClick={() => setShowMenu(!showMenu)}
          >
            {emoji}
          </button>
          {showMenu && (
            <div
              ref={menuRef}
              className="absolute left-0 top-8 z-50 bg-[var(--color-surface-1)] border border-[var(--color-border)] rounded-lg shadow-lg p-1 min-w-[120px]"
            >
              {CALLOUT_TYPES.map((ct) => (
                <button
                  key={ct.type}
                  type="button"
                  className={`w-full text-left px-2 py-1 rounded text-sm flex items-center gap-2 hover:bg-[var(--color-surface-2)] ${
                    calloutType === ct.type ? 'bg-[var(--color-surface-2)]' : ''
                  }`}
                  onClick={() => {
                    updateAttributes({ type: ct.type, emoji: ct.emoji });
                    setShowMenu(false);
                  }}
                >
                  <span>{ct.emoji}</span>
                  <span>{ct.label}</span>
                </button>
              ))}
            </div>
          )}
        </div>
        <div className="flex-1 min-w-0">
          <NodeViewContent />
        </div>
      </div>
    </NodeViewWrapper>
  );
}
