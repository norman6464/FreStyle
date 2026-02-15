import { useState } from 'react';
import { QuestionMarkCircleIcon, XMarkIcon } from '@heroicons/react/24/outline';

const SHORTCUTS = [
  { keys: 'Ctrl+B', label: '太字' },
  { keys: 'Ctrl+I', label: '斜体' },
  { keys: 'Ctrl+U', label: '下線' },
  { keys: 'Ctrl+Shift+S', label: '取り消し線' },
  { keys: 'Ctrl+Z', label: '元に戻す' },
  { keys: 'Ctrl+Shift+Z', label: 'やり直す' },
  { keys: 'Ctrl+Shift+7', label: '番号付きリスト' },
  { keys: 'Ctrl+Shift+8', label: '箇条書き' },
  { keys: '/', label: 'スラッシュコマンド' },
];

export default function KeyboardShortcutsHelp() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <button
        type="button"
        aria-label="ショートカット一覧"
        className="w-6 h-6 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-faint)] transition-colors"
        onClick={() => setOpen(true)}
      >
        <QuestionMarkCircleIcon className="w-4 h-4" />
      </button>
      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-[var(--color-surface-1)] rounded-lg shadow-xl p-5 w-80 max-h-[80vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">キーボードショートカット</h3>
              <button
                type="button"
                aria-label="閉じる"
                className="w-6 h-6 flex items-center justify-center rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-faint)]"
                onClick={() => setOpen(false)}
              >
                <XMarkIcon className="w-4 h-4" />
              </button>
            </div>
            <div className="space-y-2">
              {SHORTCUTS.map((shortcut) => (
                <div key={shortcut.keys} className="flex items-center justify-between py-1">
                  <span className="text-xs text-[var(--color-text-secondary)]">{shortcut.label}</span>
                  <kbd className="text-[10px] bg-[var(--color-surface-3)] text-[var(--color-text-faint)] px-1.5 py-0.5 rounded font-mono">
                    {shortcut.keys}
                  </kbd>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
