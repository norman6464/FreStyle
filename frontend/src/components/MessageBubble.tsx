import { useState } from 'react';

interface MessageBubbleProps {
  isSender: boolean;
  type?: 'text' | 'image' | 'bot';
  content: string;
  id: number;
  senderName?: string;
  createdAt?: string;
  onDelete?: ((id: number) => void) | null;
  onRephrase?: ((content: string) => void) | null;
  isDeleted?: boolean;
}

export default function MessageBubble({
  isSender,
  type = 'text',
  content,
  id,
  senderName,
  createdAt,
  onDelete,
  onRephrase,
  isDeleted = false,
}: MessageBubbleProps) {
  const [showDelete, setShowDelete] = useState(false);

  const baseStyle =
    'max-w-[85%] px-4 py-2.5 rounded-2xl text-sm whitespace-pre-wrap break-words relative';

  const alignment = isSender
    ? 'self-end bg-primary-500 text-white rounded-br-sm'
    : 'self-start bg-slate-100 text-slate-900 rounded-bl-sm';

  const deletedAlignment = isSender
    ? 'self-end bg-slate-200 text-slate-500 italic rounded-br-sm'
    : 'self-start bg-slate-200 text-slate-500 italic rounded-bl-sm';

  const formatTime = (dateString?: string) => {
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
      } my-3 group`}
      onMouseEnter={() => isSender && !isDeleted && setShowDelete(true)}
      onMouseLeave={() => setShowDelete(false)}
    >
      <div className="flex flex-col w-full">
        {!isSender && senderName && (
          <span className="text-xs text-slate-500 mb-1 ml-1">{senderName}</span>
        )}

        <div className={`${baseStyle} ${isDeleted ? deletedAlignment : alignment}`}>
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
                  d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9z"
                  clipRule="evenodd"
                />
              </svg>
            </button>
          )}
        </div>

        {!isDeleted && (
          <div className={`flex items-center gap-2 mt-1 ${isSender ? 'justify-end' : 'justify-start'}`}>
            {createdAt && (
              <span className={`text-[10px] ${isSender ? 'mr-1 text-slate-400' : 'ml-1 text-slate-400'}`}>
                {formatTime(createdAt)}
              </span>
            )}
            {isSender && onRephrase && (
              <button
                onClick={() => onRephrase(content)}
                className="text-[10px] text-primary-500 hover:text-primary-700 transition-colors opacity-0 group-hover:opacity-100"
              >
                言い換え
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
