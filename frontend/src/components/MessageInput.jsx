// src/components/MessageInput.jsx
import { useState } from 'react';
import {
  PaperAirplaneIcon,
  PaperClipIcon,
  GlobeAltIcon,
} from '@heroicons/react/24/solid';

export default function MessageInput({ onSend }) {
  const [text, setText] = useState('');

  const handleSend = () => {
    if (!text.trim()) return;
    onSend(text);
    setText('');
  };

  return (
    <div className="w-full max-w-[70%] mx-auto bg-white px-4 py-3 flex items-center gap-2 border-t border-gray-300 rounded-full">
      <button className="text-gray-500 hover:text-black">
        <PaperClipIcon className="h-5 w-5" />
      </button>
      <button className="text-gray-500 hover:text-black">
        <GlobeAltIcon className="h-5 w-5" />
      </button>

      <input
        type="text"
        className="flex-1 bg-gray-200 text-black rounded-full px-4 py-2 outline-none"
        placeholder="メッセージを入力"
        value={text}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={(e) => e.key === 'Enter' && handleSend()}
      />

      <button
        onClick={handleSend}
        className="text-white bg-blue-500 p-2 rounded-full hover:bg-blue-600 transition"
      >
        <PaperAirplaneIcon className="h-5 w-5 rotate-90" />
      </button>
    </div>
  );
}
