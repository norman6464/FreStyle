import { useState, useRef, useEffect, KeyboardEvent } from 'react';
import { PaperAirplaneIcon, PlusIcon } from '@heroicons/react/24/solid';

interface MessageInputProps {
  onSend: (text: string) => void;
  isSending?: boolean;
}

export default function MessageInput({ onSend, isSending = false }: MessageInputProps) {
  const [text, setText] = useState('');
  const [isComposing, setIsComposing] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const minRows = 1;
  const maxRows = 8;

  useEffect(() => {
    const textarea = textareaRef.current;
    if (textarea) {
      textarea.style.height = 'auto';
      const lineHeight = parseFloat(getComputedStyle(textarea).lineHeight);
      const paddingTop = parseFloat(getComputedStyle(textarea).paddingTop);
      const paddingBottom = parseFloat(
        getComputedStyle(textarea).paddingBottom
      );

      const maxHeight = maxRows * lineHeight + paddingTop + paddingBottom;
      const newScrollHeight = textarea.scrollHeight;
      const newHeight = Math.min(newScrollHeight, maxHeight);

      if (newHeight > 0) {
        textarea.style.height = `${newHeight}px`;
      } else {
        textarea.style.height = `${
          minRows * lineHeight + paddingTop + paddingBottom
        }px`;
      }

      textarea.style.overflowY =
        newScrollHeight > maxHeight ? 'scroll' : 'hidden';
    }
  }, [text]);

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
    <div className="w-full bg-white flex items-end p-3 border-t border-slate-200">
      <button
        className="text-slate-500 hover:text-slate-700 p-2.5 rounded-full transition-colors duration-150 flex-shrink-0 mb-1"
        aria-label="添付ファイル"
      >
        <PlusIcon className="h-6 w-6" />
      </button>

      <div className="flex-1 min-w-0">
        <textarea
          ref={textareaRef}
          rows={minRows}
          className="w-full bg-transparent text-slate-800 outline-none resize-none px-3 py-2 placeholder-slate-400 leading-6"
          placeholder="メッセージを入力..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={handleKeyDown}
          onCompositionStart={() => setIsComposing(true)}
          onCompositionEnd={() => setIsComposing(false)}
          disabled={isSending}
        />
      </div>

      <button
        onClick={handleSend}
        className="text-white bg-primary-500 p-2.5 rounded-full hover:bg-primary-600 transition-colors duration-150 flex-shrink-0 disabled:bg-slate-300 mb-1 ml-2"
        disabled={!text.trim() || isSending}
        aria-label="送信"
      >
        <PaperAirplaneIcon className="h-6 w-6 rotate-90" />
      </button>

      {isSending && (
        <span className="text-xs text-slate-400 ml-2 mb-2 flex-shrink-0">送信中...</span>
      )}
    </div>
  );
}
