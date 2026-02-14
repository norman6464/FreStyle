import { useState } from 'react';
import { ClipboardDocumentIcon, ClipboardDocumentCheckIcon, TrashIcon } from '@heroicons/react/24/outline';

interface MessageBubbleAiProps {
  isSender: boolean;
  type?: 'text' | 'image' | 'bot';
  content: string;
  id: number;
  onDelete?: (id: number) => void;
  onCopy?: ((id: number, content: string) => void) | null;
  isCopied?: boolean;
  isDeleted?: boolean;
}

export default function MessageBubbleAi({
  isSender,
  type = 'text',
  content,
  id,
  onDelete,
  onCopy,
  isCopied = false,
  isDeleted = false,
}: MessageBubbleAiProps) {
  const [showDelete, setShowDelete] = useState(false);

  const baseStyle =
    'max-w-[85%] px-4 py-2.5 rounded-2xl text-sm whitespace-pre-wrap break-words relative';

  const alignment = isSender
    ? 'self-end bg-primary-500 text-white rounded-br-sm'
    : 'self-start bg-surface-3 text-[var(--color-text-primary)] rounded-bl-sm';

  const deletedStyle = 'bg-surface-3 text-[var(--color-text-muted)] italic';

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
            <TrashIcon className="w-3 h-3" />
          </button>
        )}
      </div>

      {!isDeleted && onCopy && (
        <div className={`flex items-center mt-1 ${isSender ? 'justify-end' : 'justify-start'}`}>
          <button
            onClick={() => onCopy(id, content)}
            title={isCopied ? 'コピー済み' : 'コピー'}
            aria-label={isCopied ? 'コピー済み' : 'メッセージをコピー'}
            className="text-[var(--color-text-faint)] hover:text-[var(--color-text-secondary)] transition-colors opacity-0 group-hover:opacity-100"
          >
            {isCopied ? (
              <ClipboardDocumentCheckIcon className="w-3.5 h-3.5 text-green-500" />
            ) : (
              <ClipboardDocumentIcon className="w-3.5 h-3.5" />
            )}
          </button>
        </div>
      )}
    </div>
  );
}
