// MessageBubble.jsx
import { useState } from 'react';

export default function MessageBubble({
  isSender,
  type = 'text',
  content,
  id,
  senderName,
  createdAt,
  onDelete,
  isDeleted = false,
}) {
  const [showDelete, setShowDelete] = useState(false);

  const baseStyle =
    'max-w-[70%] px-4 py-3 rounded-2xl text-sm whitespace-pre-wrap break-words shadow-md relative';

  const alignment = isSender
    ? 'self-end bg-gradient-primary text-white rounded-br-none'
    : 'self-start bg-gray-100 text-gray-800 rounded-bl-none border-l-4 border-primary-400';

  const deletedStyle = 'bg-gray-200 text-gray-500 italic';

  const formatTime = (dateString) => {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleTimeString('ja-JP', {
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div
      className={`flex ${
        isSender ? 'justify-end' : 'justify-start'
      } my-3 animate-fade-in group`}
      onMouseEnter={() => isSender && !isDeleted && setShowDelete(true)}
      onMouseLeave={() => setShowDelete(false)}
    >
      <div className="flex flex-col max-w-[70%]">
        {!isSender && senderName && (
          <span className="text-xs text-gray-500 mb-1 ml-1">{senderName}</span>
        )}

        <div className={`${baseStyle} ${isDeleted ? deletedStyle : alignment}`}>
          {isDeleted ? (
            <p>メッセージを削除しました</p>
          ) : (
            <>
              {type === 'text' && <p>{content}</p>}
              {type === 'image' && (
                <img
                  src={content}
                  alt="画像"
                  className="max-w-full rounded-lg shadow-md"
                />
              )}
              {type === 'bot' && <p className="italic opacity-80">{content}</p>}
            </>
          )}

          {!isDeleted && createdAt && (
            <span
              className={`absolute bottom-1 right-2 text-[10px] ${
                isSender ? 'text-white/70' : 'text-gray-400'
              }`}
            >
              {formatTime(createdAt)}
            </span>
          )}

          {isSender && showDelete && onDelete && !isDeleted && (
            <button
              onClick={() => onDelete(id)}
              className="absolute -top-2 -right-2 bg-red-500 hover:bg-red-700 text-white rounded-full p-1 shadow-md transition-all"
              title="削除"
            >
              <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fillRule="evenodd"
                  d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9z"
                  clipRule="evenodd"
                />
              </svg>
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
