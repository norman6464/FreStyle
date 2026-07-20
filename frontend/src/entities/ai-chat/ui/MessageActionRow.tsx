import {
  ClipboardDocumentCheckIcon,
  ClipboardDocumentIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';
import { formatHourMinute } from '@/shared/lib/formatters';

interface Props {
  isSender: boolean;
  id: string;
  content: string;
  createdAt?: string;
  isCopied: boolean;
  onCopy?: ((id: string, content: string) => void) | null;
  onDelete?: ((id: string) => void) | null;
  visible: boolean;
}

/**
 * メッセージ末尾 の アクション 行 (時刻 + コピー + 削除)。
 * 自分 のメッセージ では 削除 ボタン も 出す が、 他人 / AI では コピー だけ。
 * `visible` は hover で 表示 / focus で 表示 を フェード 制御。
 */
export default function MessageActionRow({
  isSender,
  id,
  content,
  createdAt,
  isCopied,
  onCopy,
  onDelete,
  visible,
}: Props) {
  return (
    <div
      className={`flex items-center gap-2 mt-1 ${
        isSender ? 'justify-end' : 'justify-start'
      } text-[var(--color-text-faint)]`}
    >
      {createdAt && (
        <span className="text-[10px]">{formatHourMinute(createdAt)}</span>
      )}
      {onCopy && (
        <button
          onClick={() => onCopy(id, content)}
          title={isCopied ? 'コピー済み' : 'コピー'}
          aria-label={isCopied ? 'コピー済み' : 'メッセージをコピー'}
          className={`hover:text-[var(--color-text-secondary)] transition-colors ${
            visible ? 'opacity-100' : 'opacity-0'
          } focus:opacity-100`}
        >
          {isCopied ? (
            <ClipboardDocumentCheckIcon className="w-3.5 h-3.5 text-green-500" />
          ) : (
            <ClipboardDocumentIcon className="w-3.5 h-3.5" />
          )}
        </button>
      )}
      {isSender && onDelete && (
        <button
          onClick={() => onDelete(id)}
          title="削除"
          aria-label="メッセージを削除"
          className={`hover:text-red-400 transition-colors ${
            visible ? 'opacity-100' : 'opacity-0'
          } focus:opacity-100`}
        >
          <TrashIcon className="w-3.5 h-3.5" />
        </button>
      )}
    </div>
  );
}
