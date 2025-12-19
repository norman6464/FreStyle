import { useState, useRef, useEffect } from 'react';
import { PaperAirplaneIcon, PlusIcon } from '@heroicons/react/24/solid';

export default function MessageInput({ onSend }) {
  const [text, setText] = useState('');
  const textareaRef = useRef(null);
  const minRows = 1;
  const maxRows = 8; // 最大8行で高さを制限

  // テキストエリアの自動高さ調整
  useEffect(() => {
    const textarea = textareaRef.current;
    if (textarea) {
      // 一旦最小行に戻してから、スクロールに合わせて高さを再計算
      textarea.style.height = 'auto';
      const lineHeight = parseFloat(getComputedStyle(textarea).lineHeight);
      const paddingTop = parseFloat(getComputedStyle(textarea).paddingTop);
      const paddingBottom = parseFloat(
        getComputedStyle(textarea).paddingBottom
      );

      const maxHeight = maxRows * lineHeight + paddingTop + paddingBottom;

      const newScrollHeight = textarea.scrollHeight;

      // 新しい高さを計算し、最大高さを超えないように制限
      const newHeight = Math.min(newScrollHeight, maxHeight);

      // 高さを設定。ただし、最低限の高さ（1行）を確保
      if (newHeight > 0) {
        textarea.style.height = `${newHeight}px`;
      } else {
        textarea.style.height = `${
          minRows * lineHeight + paddingTop + paddingBottom
        }px`;
      }

      // 最大高さを超えた場合はスクロールバーを表示
      textarea.style.overflowY =
        newScrollHeight > maxHeight ? 'scroll' : 'hidden';
    }
  }, [text]);

  // 送信処理
  const handleSend = () => {
    if (!text.trim()) return;
    onSend(text);
    setText('');
  };

  // Enterキーでの送信処理（Shift+Enterで改行を可能にする）
  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault(); // Enterによるフォーム送信を防ぐ
      handleSend();
    }
  };

  return (
    <div className="w-full bg-white flex items-end p-3 border-t-2 border-gray-100 shadow-2xl transition-all duration-300">
      {/* ＋アイコンボタン（左端） */}
      <button
        className="text-primary-400 hover:text-primary-600 p-2.5 rounded-full transition-all duration-150 flex-shrink-0 mb-1 hover:bg-primary-50"
        aria-label="添付ファイル"
      >
        <PlusIcon className="h-6 w-6" />
      </button>

      {/* テキストエリアのコンテナ */}
      <div className="flex-1 min-w-0">
        <textarea
          ref={textareaRef}
          rows={minRows}
          className="w-full bg-transparent text-gray-800 outline-none resize-none px-3 py-2 placeholder-gray-400 leading-6 font-medium"
          placeholder="メッセージを入力..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={handleKeyDown}
        />
      </div>

      {/* 送信ボタン（右端） */}
      <button
        onClick={handleSend}
        className="text-white bg-gradient-primary p-2.5 rounded-full hover:shadow-lg transition-all duration-150 flex-shrink-0 disabled:bg-gray-300 disabled:shadow-none mb-1 ml-2 transform hover:scale-110"
        disabled={!text.trim()}
        aria-label="送信"
      >
        <PaperAirplaneIcon className="h-6 w-6 rotate-90" />
      </button>
    </div>
  );
}
