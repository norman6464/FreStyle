import { useState } from 'react';
import ChatRoomList from '../components/ChatRoomList';
import MessageBubble from '../components/MessageBubble';
import MessageInput from '../components/MessageInput';
import HamburgerMenu from '../components/HamburgerMenu';

export default function ChatPage() {
  const [messages, setMessages] = useState([
    { id: 1, content: 'こんにちは！', type: 'text', isSender: false },
  ]);

  const handleSend = (text) => {
    setMessages((prev) => [
      ...prev,
      { id: Date.now(), content: text, type: 'text', isSender: true },
    ]);

    // 仮のボット応答
    setTimeout(() => {
      setMessages((prev) => [
        ...prev,
        {
          id: Date.now() + 1,
          content: 'ボットの応答です。',
          type: 'bot',
          isSender: false,
        },
      ]);
    }, 800);
  };

  return (
    <>
      <HamburgerMenu />
      <div className="flex h-screen bg-gray-100 text-black">
        {/* 左サイドバー（チャットルーム一覧） */}
        <aside className="hidden lg:block w-64 border-r border-gray-300 bg-white overflow-y-auto">
          <ChatRoomList />
        </aside>

        {/* チャットエリア */}
        <main className="flex flex-col flex-1">
          <div
            className="flex-1 overflow-y-auto px-4 py-6 space-y-2
          w-full max-w-[800px] lg:max-w-[70%] lg:mx-auto
        "
          >
            {messages.map((msg) => (
              <MessageBubble key={msg.id} {...msg} />
            ))}
          </div>

          <MessageInput onSend={handleSend} />
        </main>
      </div>
    </>
  );
}
