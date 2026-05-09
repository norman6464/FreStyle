import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import rehypeHighlight from 'rehype-highlight';
import { ArrowDownTrayIcon, ClipboardDocumentIcon } from '@heroicons/react/24/outline';
import type { SaveStatus } from '../hooks/useNoteEditor';
import { useToast } from '../hooks/useToast';
import { tiptapToMarkdown } from '../utils/tiptapToMarkdown';
import { getNoteStats } from '../utils/noteStats';
import WordCount from './WordCount';
import LineCount from './LineCount';
import ReadingTime from './ReadingTime';

interface NoteMarkdownEditorProps {
  title: string;
  content: string;
  saveStatus: SaveStatus;
  onTitleChange: (title: string) => void;
  onContentChange: (content: string) => void;
}

const SAVE_STATUS_CONFIG: Record<Exclude<SaveStatus, 'idle'>, { label: string; color: string }> = {
  unsaved: { label: '未保存', color: 'text-amber-500' },
  saving: { label: '保存中...', color: 'text-[var(--color-text-muted)]' },
  saved: { label: '保存済み', color: 'text-emerald-500' },
};

/**
 * NoteMarkdownEditor — GitHub README 風の Edit / Preview タブ付き Markdown エディタ。
 *
 * 旧実装（BlockEditor + TipTap）から切替: ノートは生 Markdown 文字列として保存され、
 * Preview タブで react-markdown レンダリングする。
 *
 * 互換性:
 *   - 旧データは TipTap JSON で保存されているので、 ロード時に
 *     `{"type":"doc"...` で始まる文字列を検出した場合は tiptapToMarkdown で
 *     変換した結果を初期表示にする（ユーザが保存し直すと Markdown 化される）
 *   - 完全な双方向同期はしない（Preview の WYSIWYG 編集は不可）
 */
export default function NoteMarkdownEditor({
  title,
  content,
  saveStatus,
  onTitleChange,
  onContentChange,
}: NoteMarkdownEditorProps) {
  const [tab, setTab] = useState<'edit' | 'preview'>('edit');
  const titleRef = useRef<HTMLInputElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const { showToast } = useToast();

  // 旧 TipTap JSON データの自動 markdown 化（初回ロード時のみ反映）。
  // ユーザが編集 → 保存すると content が markdown で上書きされ、 以降この分岐は通らない。
  const displayContent = useMemo(() => convertLegacyContent(content), [content]);
  const lastConvertedRef = useRef<string | null>(null);

  useEffect(() => {
    if (
      content !== displayContent &&
      lastConvertedRef.current !== content &&
      isLikelyTiptapJson(content)
    ) {
      // 一度だけ親に通知して store を markdown 化する。
      lastConvertedRef.current = content;
      onContentChange(displayContent);
    }
  }, [content, displayContent, onContentChange]);

  const stats = useMemo(() => getNoteStats(displayContent), [displayContent]);

  const handleTitleKeyDown = useCallback((e: React.KeyboardEvent<HTMLInputElement>) => {
    // IME 変換中の Enter（日本語入力の確定）はフォーカス移動しない。
    // ここを判定しないと「変換確定の Enter」で textarea にフォーカスが飛び、
    // 直前の未確定文字列（"ついて" など）がそのまま textarea に入力されてしまう。
    if (e.key !== 'Enter') return;
    if (e.nativeEvent.isComposing || e.keyCode === 229) return;
    e.preventDefault();
    textareaRef.current?.focus();
  }, []);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(displayContent);
      showToast('success', 'Markdown をコピーしました');
    } catch {
      showToast('error', 'コピーに失敗しました');
    }
  }, [displayContent, showToast]);

  const handleDownload = useCallback(() => {
    const blob = new Blob([displayContent], { type: 'text/markdown;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${title || '無題'}.md`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, [displayContent, title]);

  return (
    <div className="flex flex-col h-full p-6 max-w-3xl mx-auto w-full">
      <input
        ref={titleRef}
        type="text"
        value={title}
        onChange={(e) => onTitleChange(e.target.value)}
        onKeyDown={handleTitleKeyDown}
        placeholder="無題"
        aria-label="ノートのタイトル"
        className="text-2xl font-bold text-[var(--color-text-primary)] bg-transparent border-none outline-none w-full mb-4 placeholder:text-[var(--color-text-faint)]"
      />

      <div className="flex items-center gap-1 mb-3">
        <TabButton active={tab === 'edit'} onClick={() => setTab('edit')}>
          Edit
        </TabButton>
        <TabButton active={tab === 'preview'} onClick={() => setTab('preview')}>
          Preview
        </TabButton>
      </div>

      <div className="flex-1 min-h-0 overflow-hidden">
        {tab === 'edit' ? (
          <textarea
            ref={textareaRef}
            value={displayContent}
            onChange={(e) => onContentChange(e.target.value)}
            placeholder="# 見出し&#10;&#10;Markdown でノートを書きましょう..."
            spellCheck={false}
            className="w-full h-full resize-none p-4 rounded-md bg-surface-1 border border-surface-3 text-sm font-mono text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:outline-none focus:border-primary-500/50 transition-colors"
          />
        ) : (
          <div className="w-full h-full overflow-y-auto p-4 rounded-md bg-surface-1 border border-surface-3 prose prose-invert prose-sm max-w-none">
            {displayContent.trim() ? (
              <MarkdownView content={displayContent} />
            ) : (
              <p className="text-[var(--color-text-muted)] italic">プレビューするコンテンツがありません</p>
            )}
          </div>
        )}
      </div>

      <div className="flex items-center gap-3 pt-3 mt-3 border-t border-surface-3" aria-label="ノート統計">
        <WordCount charCount={stats.charCount} />
        <LineCount lineCount={stats.lineCount} />
        <ReadingTime charCount={stats.charCount} />
        {saveStatus !== 'idle' && (
          <span className={`text-xs ${SAVE_STATUS_CONFIG[saveStatus].color}`} aria-label="保存状態">
            {SAVE_STATUS_CONFIG[saveStatus].label}
          </span>
        )}
        <div className="ml-auto flex items-center gap-1">
          <button
            type="button"
            onClick={handleCopy}
            aria-label="Markdown をコピー"
            className="p-1.5 rounded hover:bg-surface-3 text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] transition-colors"
            title="Markdown をコピー"
          >
            <ClipboardDocumentIcon className="w-4 h-4" />
          </button>
          <button
            type="button"
            onClick={handleDownload}
            aria-label="Markdown でダウンロード"
            className="p-1.5 rounded hover:bg-surface-3 text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] transition-colors"
            title="Markdown でダウンロード"
          >
            <ArrowDownTrayIcon className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}

function TabButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={`px-3 py-1 text-xs rounded-md transition-colors ${
        active
          ? 'bg-surface-2 text-[var(--color-text-primary)]'
          : 'text-[var(--color-text-muted)] hover:bg-surface-2 hover:text-[var(--color-text-secondary)]'
      }`}
    >
      {children}
    </button>
  );
}

// 旧データが TipTap JSON で保存されているケースを検出する。
function isLikelyTiptapJson(s: string): boolean {
  if (!s) return false;
  const trimmed = s.trim();
  if (!trimmed.startsWith('{')) return false;
  try {
    const parsed = JSON.parse(trimmed);
    return parsed && typeof parsed === 'object' && parsed.type === 'doc';
  } catch {
    return false;
  }
}

function convertLegacyContent(s: string): string {
  return isLikelyTiptapJson(s) ? tiptapToMarkdown(s) : s;
}

function MarkdownView({ content }: { content: string }) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm, remarkBreaks]}
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
        table: ({ children }) => (
          <div className="overflow-x-auto">
            <table className="text-sm border-collapse">{children as ReactNode}</table>
          </div>
        ),
      }}
    >
      {content}
    </ReactMarkdown>
  );
}
