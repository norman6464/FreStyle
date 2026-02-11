import { useState } from 'react';

interface MessageBubbleAiProps {
  isSender: boolean;
  type?: 'text' | 'image' | 'bot';
  content: string;
  id: number;
  onDelete?: (id: number) => void;
  isDeleted?: boolean;
}

export default function MessageBubbleAi({
  isSender,
  type = 'text',
  content,
  id,
  onDelete,
  isDeleted = false,
}: MessageBubbleAiProps) {
  const [showDelete, setShowDelete] = useState(false);

  const baseStyle =
    'max-w-[85%] px-4 py-2.5 rounded-2xl text-sm whitespace-pre-wrap break-words relative';

  const alignment = isSender
    ? 'self-end bg-primary-500 text-white rounded-br-sm'
    : 'self-start bg-slate-100 text-slate-900 rounded-bl-sm';

  const deletedStyle = 'bg-slate-200 text-slate-500 italic';

  return (
    <div
      className={`flex ${
        isSender ? 'justify-end' : 'justify-start'
      } my-3 group`}
      onMouseEnter={() => isSender && !isDeleted && setShowDelete(true)}
      onMouseLeave={() => setShowDelete(false)}
    >
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

        {isSender && showDelete && onDelete && !isDeleted && (
          <button
            onClick={() => onDelete(id)}
            className="absolute -top-2 -right-2 bg-red-500 hover:bg-red-600 text-white rounded-full p-1 transition-colors duration-150"
            title="削除"
          >
            <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
              <path
                fillRule="evenodd"
                d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z"
                clipRule="evenodd"
              />
            </svg>
          </button>
        )}
      </div>
    </div>
  );
}
