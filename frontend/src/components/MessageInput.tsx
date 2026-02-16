import { useState, useRef, useEffect, KeyboardEvent } from 'react';
import { PaperAirplaneIcon, PlusIcon } from '@heroicons/react/24/solid';
import { useAutoResizeTextarea } from '../hooks/useAutoResizeTextarea';

interface MessageInputProps {
  onSend: (text: string) => void;
  isSending?: boolean;
}

export default function MessageInput({ onSend, isSending = false }: MessageInputProps) {
  const [text, setText] = useState('');
  const [isComposing, setIsComposing] = useState(false);
  const textareaRef = useAutoResizeTextarea({ text });

  const prevIsSendingRef = useRef(isSending);
  useEffect(() => {
    if (prevIsSendingRef.current && !isSending) {
      textareaRef.current?.focus();
    }
    prevIsSendingRef.current = isSending;
  }, [isSending, textareaRef]);

  const handleSend = () => {
    if (!text.trim() || isSending) return;
    onSend(text);
    setText('');
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey && !isComposing) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="w-full bg-surface-1 flex items-end p-3 border-t border-surface-3">
      <button
        className="text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] p-2.5 rounded-full transition-colors duration-150 flex-shrink-0 mb-1"
        aria-label="添付ファイル"
      >
        <PlusIcon className="h-6 w-6" />
      </button>

      <div className="flex-1 min-w-0">
        <textarea
          ref={textareaRef}
          rows={1}
          className="w-full bg-transparent text-[var(--color-text-primary)] outline-none resize-none px-3 py-2 placeholder-slate-400 leading-6"
          placeholder="メッセージを入力..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={handleKeyDown}
          onCompositionStart={() => setIsComposing(true)}
          onCompositionEnd={() => setIsComposing(false)}
          disabled={isSending}
        />
        {text.length > 0 && (
          <span data-testid="char-count" className="block px-3 pb-1 text-[10px] text-[var(--color-text-muted)] text-right" aria-live="polite">
            {text.length}
          </span>
        )}
      </div>

      <button
        onClick={handleSend}
        className="text-white bg-primary-500 p-2.5 rounded-full hover:bg-primary-600 transition-colors duration-150 flex-shrink-0 disabled:bg-surface-3 mb-1 ml-2"
        disabled={!text.trim() || isSending}
        aria-label="送信"
      >
        <PaperAirplaneIcon className="h-6 w-6 rotate-90" />
      </button>

      {isSending && (
        <span className="text-xs text-[var(--color-text-faint)] ml-2 mb-2 flex-shrink-0">送信中...</span>
      )}
    </div>
  );
}
