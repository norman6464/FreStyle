// src/pages/ChatPage.jsx
import { useState } from 'react';
import MessageBubble from '../components/MessageBubble';
import MessageInput from '../components/MessageInput';

export default function AskAiPage() {
  const [messages, setMessages] = useState([
    { id: 1, content: '始めましょうか。', type: 'bot', isSender: false },
  ]);

  const handleSend = (text) => {
    setMessages((prev) => [
      ...prev,
      { id: Date.now(), content: text, type: 'text', isSender: true },
    ]);

    // ダミーのボット応答
    setTimeout(() => {
      setMessages((prev) => [
        ...prev,
        {
          id: Date.now() + 1,
          content: 'AI応答です。',
          type: 'bot',
          isSender: false,
        },
      ]);
    }, 1000);
  };

  return (
    <div className="flex flex-col h-screen bg-gray-100 text-black">
      <div className="flex-1 overflow-y-auto px-4 py-6 space-y-2 max-w-3xl mx-auto w-full">
        {messages.map((msg) => (
          <MessageBubble key={msg.id} {...msg} />
        ))}
      </div>

      <MessageInput onSend={handleSend} />
    </div>
  );
}
