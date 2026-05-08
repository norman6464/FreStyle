import { memo, useState, ReactNode } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import {
  ClipboardDocumentIcon,
  ClipboardDocumentCheckIcon,
  TrashIcon,
  SparklesIcon,
} from '@heroicons/react/24/outline';
import 'highlight.js/styles/github-dark.css';
import { formatHourMinute } from '../utils/formatters';

/**
 * 自分のメッセージに添付された画像 / ドキュメントの参照（送信中の Object URL も許容）。
 * MessageBubble 内では画像のみインラインプレビュー、それ以外はファイル名カードに落とす。
 *
 * `kind` は backend `domain.AttachmentKind*` と整合する `'image' | 'document'`。
 */
export interface MessageAttachmentView {
  key: string;
  filename: string;
  contentType: string;
  kind: 'image' | 'document';
  sizeBytes: number;
  /** 送信中チップから引き継ぐローカル Object URL */
  previewUrl?: string;
  /** 送信完了後 backend が返す CDN / S3 URL（あれば優先） */
  url?: string;
}

interface MessageBubbleProps {
  isSender: boolean;
  type?: 'text' | 'image' | 'bot';
  content: string;
  id: string;
  senderName?: string;
  createdAt?: string;
  attachments?: MessageAttachmentView[];
  onDelete?: ((id: string) => void) | null;
  onCopy?: ((id: string, content: string) => void) | null;
  isCopied?: boolean;
  isDeleted?: boolean;
  /** 旧 API 互換のため受けるが本コンポーネントでは使わない */
  onRephrase?: ((content: string) => void) | null;
}

/**
 * メッセージ表示コンポーネント。
 *
 * 設計:
 *   - 自分のメッセージは右寄せのコンパクトな塊（軽いハイライト背景）
 *   - アシスタント / 他者のメッセージは左寄せで **背景なしのフラット表示**。
 *     見出し / リスト / コード / 表など Markdown 要素を本文として描画する
 *   - カードや角丸の装飾はあえて入れない（一般的な汎用 AI チャット UI に寄せる）
 *
 * Markdown:
 *   - GFM（表 / タスクリスト / strikethrough）対応
 *   - コードブロックは highlight.js でシンタックスハイライト
 */
export default memo(function MessageBubble({
  isSender,
  type = 'text',
  content,
  id,
  senderName,
  createdAt,
  attachments,
  onDelete,
  onCopy,
  isCopied = false,
  isDeleted = false,
}: MessageBubbleProps) {
  const [showActions, setShowActions] = useState(false);

  if (type === 'image') {
    return (
      <div
        className={`my-3 flex ${isSender ? 'justify-end' : 'justify-start'}`}
        role="article"
        aria-label="画像メッセージ"
      >
        <img src={content} alt="画像" className="max-w-[85%] rounded-lg shadow-md" />
      </div>
    );
  }

  if (isDeleted) {
    return (
      <div
        className={`my-3 text-xs text-[var(--color-text-muted)] italic ${
          isSender ? 'text-right' : 'text-left'
        }`}
      >
        メッセージを削除しました
      </div>
    );
  }

  if (isSender) {
    return (
      <div
        className="my-4 group flex justify-end"
        onMouseEnter={() => setShowActions(true)}
        onMouseLeave={() => setShowActions(false)}
        role="article"
        aria-label="自分のメッセージ"
      >
        <div className="max-w-[85%] flex flex-col items-end gap-1">
          {attachments && attachments.length > 0 && (
            <AttachmentList attachments={attachments} />
          )}
          {content && (
            <div className="px-4 py-2 rounded-2xl bg-[var(--color-surface-3)] text-[var(--color-text-primary)] text-sm whitespace-pre-wrap break-words">
              {content}
            </div>
          )}
          <MessageActionRow
            isSender
            id={id}
            content={content}
            createdAt={createdAt}
            isCopied={isCopied}
            onCopy={onCopy}
            onDelete={onDelete}
            visible={showActions}
          />
        </div>
      </div>
    );
  }

  // アシスタント / 他者のメッセージ: フラット、本文に Markdown
  return (
    <div
      className="my-6 group flex gap-3"
      onMouseEnter={() => setShowActions(true)}
      onMouseLeave={() => setShowActions(false)}
      role="article"
      aria-label={senderName ? `${senderName}のメッセージ` : 'AIのメッセージ'}
    >
      <div className="flex-shrink-0 w-7 h-7 rounded-full bg-[var(--color-surface-3)] flex items-center justify-center text-[var(--color-text-muted)]">
        <SparklesIcon className="w-4 h-4" aria-hidden="true" />
      </div>
      <div className="flex-1 min-w-0">
        {senderName && (
          <p className="text-xs text-[var(--color-text-muted)] mb-1">{senderName}</p>
        )}
        <div className="prose prose-sm max-w-none text-[var(--color-text-primary)] leading-relaxed">
          {type === 'bot' ? (
            <p className="italic opacity-80">{content}</p>
          ) : (
            <MarkdownView content={content} />
          )}
        </div>
        <MessageActionRow
          isSender={false}
          id={id}
          content={content}
          createdAt={createdAt}
          isCopied={isCopied}
          onCopy={onCopy}
          visible={showActions}
        />
      </div>
    </div>
  );
});

interface MessageActionRowProps {
  isSender: boolean;
  id: string;
  content: string;
  createdAt?: string;
  isCopied: boolean;
  onCopy?: ((id: string, content: string) => void) | null;
  onDelete?: ((id: string) => void) | null;
  visible: boolean;
}

function MessageActionRow({
  isSender,
  id,
  content,
  createdAt,
  isCopied,
  onCopy,
  onDelete,
  visible,
}: MessageActionRowProps) {
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

/**
 * 自分のメッセージに紐付く添付の表示。
 *
 * 画像は max-w 240px のサムネで横並び（Object URL でローカル送信中も表示できる）。
 * 画像以外（PR-G2 で増える PDF / CSV）は filename カードのフォールバック。
 */
function AttachmentList({ attachments }: { attachments: MessageAttachmentView[] }) {
  return (
    <div className="flex flex-wrap gap-2 justify-end" aria-label="添付ファイル">
      {attachments.map((a) => {
        const src = a.url ?? a.previewUrl;
        if (a.kind === 'image' && src) {
          return (
            <img
              key={a.key}
              src={src}
              alt={a.filename}
              className="max-w-[240px] max-h-[240px] rounded-lg object-cover border border-[var(--color-surface-3)]"
            />
          );
        }
        return (
          <div
            key={a.key}
            className="px-3 py-2 rounded-lg border border-[var(--color-surface-3)] bg-[var(--color-surface-2)] text-xs text-[var(--color-text-primary)]"
          >
            {a.filename}
          </div>
        );
      })}
    </div>
  );
}

// react-markdown のラッパ。コンポーネントマップで pre / code / a / table などを
// プロジェクトのトーンに揃える。`prose` クラス（Tailwind Typography）が当たって
// いる前提なので、要素ごとにクラス上書きは最小限。
function MarkdownView({ content }: { content: string }) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      rehypePlugins={[rehypeHighlight]}
      components={{
        a: ({ href, children }) => (
          <a
            href={href}
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary-400 underline-offset-2 hover:underline"
          >
            {children as ReactNode}
          </a>
        ),
        // インラインコード（pre 直下でない code）にだけスタイルを当てる。
        // ブロックコードは rehype-highlight + highlight.js テーマで装飾済。
        code: ({ className, children, ...props }) => {
          const isBlock = className?.includes('language-');
          if (isBlock) {
            return (
              <code className={className} {...props}>
                {children as ReactNode}
              </code>
            );
          }
          return (
            <code className="px-1 py-0.5 rounded bg-[var(--color-surface-3)] text-[0.85em]">
              {children as ReactNode}
            </code>
          );
        },
        pre: ({ children }) => (
          <pre className="p-3 rounded-md bg-[var(--color-surface-3)] overflow-x-auto text-xs">
            {children as ReactNode}
          </pre>
        ),
        table: ({ children }) => (
          <div className="overflow-x-auto">
            <table className="text-sm border-collapse">{children as ReactNode}</table>
          </div>
        ),
        th: ({ children }) => (
          <th className="border border-[var(--color-surface-3)] px-2 py-1 bg-[var(--color-surface-2)] text-left">
            {children as ReactNode}
          </th>
        ),
        td: ({ children }) => (
          <td className="border border-[var(--color-surface-3)] px-2 py-1">
            {children as ReactNode}
          </td>
        ),
      }}
    >
      {content}
    </ReactMarkdown>
  );
}
