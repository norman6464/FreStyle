// src/components/MessageInput.jsx
import { useState } from 'react';
import {
  PaperAirplaneIcon,
  PlusIcon,
  MicrophoneIcon,
} from '@heroicons/react/24/solid';

export default function MessageInput({ onSend }) {
  const [text, setText] = useState('');

  const handleSend = () => {
    if (!text.trim()) return;
    onSend(text);
    setText('');
  };

  return (
    <div className="w-full bg-[#1e1f23] px-4 py-3 flex items-center justify-center border-t border-gray-700">
      <div className="flex items-center gap-2 w-full max-w-3xl bg-[#2a2b31] text-gray-200 rounded-2xl px-4 py-2">
        {/* 左側の + ボタン */}
        <button className="text-gray-400 hover:text-white transition">
          <PlusIcon className="h-5 w-5" />
        </button>

        {/* 入力欄 */}
        <input
          type="text"
          className="flex-1 bg-transparent outline-none placeholder-gray-500 text-white"
          placeholder="質問する"
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSend()}
        />

        {/* 右側のマイク or 送信ボタン */}
        <button
          onClick={handleSend}
          className="text-gray-400 hover:text-white transition"
        >
          {text.trim() ? (
            <PaperAirplaneIcon className="h-5 w-5 rotate-90" />
          ) : (
            <MicrophoneIcon className="h-5 w-5" />
          )}
        </button>
      </div>
    </div>
  );
}
