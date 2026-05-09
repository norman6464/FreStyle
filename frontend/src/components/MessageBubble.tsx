import { memo, useState, useMemo, ReactNode, isValidElement } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import {
  ClipboardDocumentIcon,
  ClipboardDocumentCheckIcon,
  TrashIcon,
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

  // アシスタント / 他者のメッセージ: フラット、本文に Markdown。
  // 本文が空のときは SSE で最初の token が来るまでの「考え中」状態とみなし、
  // favicon を回しながら "考え中..." ラベルを出す（Claude / ChatGPT 同様の UX）。
  const isThinking = type === 'text' && content.trim() === '';
  return (
    <div
      className="my-6 group flex gap-3"
      onMouseEnter={() => setShowActions(true)}
      onMouseLeave={() => setShowActions(false)}
      role="article"
      aria-label={senderName ? `${senderName}のメッセージ` : 'AIのメッセージ'}
    >
      <div className="flex-shrink-0 w-7 h-7 rounded-full bg-[var(--color-surface-2)] flex items-center justify-center">
        <img
          src="/favicon.svg"
          alt=""
          aria-hidden="true"
          className={`w-4 h-4 ${isThinking ? 'animate-thinking' : ''}`}
        />
      </div>
      <div className="flex-1 min-w-0">
        {senderName && (
          <p className="text-xs text-[var(--color-text-muted)] mb-1">{senderName}</p>
        )}
        {isThinking ? (
          <p
            className="text-sm text-[var(--color-text-muted)] italic"
            aria-live="polite"
          >
            考え中...
          </p>
        ) : (
          <>
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
            {/* 完了マーカー: ストリーミング placeholder ではない（= done 確定済 / 履歴ロード済）
                AI 応答の末尾に favicon を 1 つ置いて「ここで応答が締まった」ことを視覚化する。
                Claude 等のメジャー AI チャットで採用されている bookend 表現に倣う。
                id は型定義上 string だが、 旧テスト互換のため number が来ても動くよう String 化して判定。 */}
            {!String(id).startsWith('streaming-') && (
              <div className="mt-8 flex" aria-hidden="true">
                <img
                  src="/favicon.svg"
                  alt=""
                  className="w-4 h-4 opacity-60"
                />
              </div>
            )}
          </>
        )}
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
        pre: ({ children }) => <CodeBlock>{children as ReactNode}</CodeBlock>,
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

/**
 * AI 応答の Markdown 内コードブロックに、言語名表示 + コピーボタン付きのヘッダを足す。
 *
 * ChatGPT / Claude.ai と同じ UX で、ユーザーは AI 応答のコードを 1 クリックでクリップボードに
 * 取り込める。「メッセージ全体コピー」ボタンとは別に、コードブロック単位のコピー UI を提供する。
 *
 * react-markdown が rehype-highlight 後に渡してくる構造:
 *
 *   <pre>            ← この `pre` コンポーネント置換が CodeBlock のエントリ
 *     <code class="language-py">
 *       <span class="hljs-keyword">def</span>
 *       ...
 *     </code>
 *   </pre>
 *
 * `children` の最も内側の文字列を再帰的に集めて生コードを復元する（rehype-highlight が
 * 挿入した <span> はクラスだけ付けて textContent は変えないため、テキスト連結で正確に
 * 元のコードが復元できる）。
 */
function CodeBlock({ children }: { children: ReactNode }) {
  const [copied, setCopied] = useState(false);

  const language = useMemo(() => extractLanguage(children), [children]);
  const rawCode = useMemo(
    () => extractTextContent(children).replace(/\n$/, ''),
    [children]
  );

  const handleCopy = async () => {
    if (!rawCode) return;
    try {
      await navigator.clipboard.writeText(rawCode);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // clipboard API はブラウザ設定 / HTTP 環境で拒否されることがある。
      // エラーは握りつぶし、ユーザーには反応なしのままにする（壊れた挙動より無反応の方が安全）。
    }
  };

  return (
    <div className="my-2 rounded-md overflow-hidden bg-[var(--color-surface-3)]">
      <div className="flex items-center justify-between px-3 py-1 bg-[var(--color-surface-2)] border-b border-[var(--color-surface-3)] text-[11px]">
        <span className="text-[var(--color-text-muted)] font-mono">
          {language || 'code'}
        </span>
        <button
          type="button"
          onClick={handleCopy}
          aria-label={copied ? 'コードをコピーしました' : 'コードをコピー'}
          className="flex items-center gap-1 px-1.5 py-0.5 rounded text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-surface-3)] transition-colors"
        >
          {copied ? (
            <>
              <ClipboardDocumentCheckIcon className="w-3.5 h-3.5 text-green-400" />
              <span>コピー済み</span>
            </>
          ) : (
            <>
              <ClipboardDocumentIcon className="w-3.5 h-3.5" />
              <span>コピー</span>
            </>
          )}
        </button>
      </div>
      <pre className="p-3 overflow-x-auto text-xs">{children as ReactNode}</pre>
    </div>
  );
}

/**
 * 子要素の className から `language-xxx` を抽出する。複数行コードのときだけ言語が付く。
 * 言語不明（インラインコード相当 / 未知の言語）は空文字を返す。
 */
function extractLanguage(node: ReactNode): string {
  if (!isValidElement(node)) return '';
  const el = node as React.ReactElement<{ className?: string }>;
  const match = el.props.className?.match(/language-(\w+)/);
  return match?.[1] ?? '';
}

/**
 * React 要素ツリーから text node だけを連結して返す。
 * rehype-highlight 後の `<span class="hljs-...">code</span>` 構造から元のコードを復元する用途。
 */
function extractTextContent(node: ReactNode): string {
  if (typeof node === 'string') return node;
  if (typeof node === 'number') return String(node);
  if (Array.isArray(node)) return node.map(extractTextContent).join('');
  if (isValidElement(node)) {
    const props = node.props as { children?: ReactNode };
    return extractTextContent(props.children);
  }
  return '';
}
