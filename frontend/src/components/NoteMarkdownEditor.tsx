import { useCallback, useMemo, useRef, useState, type ReactNode } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import rehypeHighlight from 'rehype-highlight';
import {
  ArrowDownTrayIcon,
  ArrowUturnLeftIcon,
  ArrowUturnRightIcon,
  ClipboardDocumentIcon,
  PhotoIcon,
} from '@heroicons/react/24/outline';
import type { SaveStatus } from '../hooks/useNoteEditor';
import { useToast } from '../hooks/useToast';
import { getNoteStats } from '../utils/noteStats';
import WordCount from '@/shared/ui/WordCount';
import LineCount from '@/shared/ui/LineCount';
import ReadingTime from '@/shared/ui/ReadingTime';

interface NoteMarkdownEditorProps {
  title: string;
  content: string;
  saveStatus: SaveStatus;
  onTitleChange: (title: string) => void;
  onContentChange: (content: string) => void;
  // 画像アップロード（任意）。File を受け取り公開 URL を返す。 指定すると textarea への
  // ドラッグ&ドロップ / 貼り付け / ボタンで画像を挿入できる（教材で draw.io 図を埋め込む用途）。
  onImageUpload?: (file: File) => Promise<string>;
}

const SAVE_STATUS_CONFIG: Record<Exclude<SaveStatus, 'idle'>, { label: string; color: string }> = {
  unsaved: { label: '未保存', color: 'text-amber-500' },
  saving: { label: '保存中...', color: 'text-[var(--color-text-muted)]' },
  saved: { label: '保存済み', color: 'text-emerald-500' },
};

/**
 * NoteMarkdownEditor — GitHub README 風の Edit / Preview タブ付き Markdown エディタ。
 *
 * ノートは生 Markdown 文字列として保存され、 Preview タブで react-markdown レンダリングする。
 * 完全な双方向同期はしない（Preview の WYSIWYG 編集は不可）。
 */
export default function NoteMarkdownEditor({
  title,
  content,
  saveStatus,
  onTitleChange,
  onContentChange,
  onImageUpload,
}: NoteMarkdownEditorProps) {
  const [tab, setTab] = useState<'edit' | 'preview'>('edit');
  const [uploading, setUploading] = useState(false);
  const titleRef = useRef<HTMLInputElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const { showToast } = useToast();

  const stats = useMemo(() => getNoteStats(content), [content]);

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
      await navigator.clipboard.writeText(content);
      showToast('success', 'Markdown をコピーしました');
    } catch {
      showToast('error', 'コピーに失敗しました');
    }
  }, [content, showToast]);

  // textarea にフォーカスを戻してから document.execCommand('undo'/'redo')。
  //
  // controlled <textarea> でも、 ブラウザはキー入力ベースの undo 履歴を
  // textarea 自身に保持している（外から value をリセットしない限り）。
  // execCommand は deprecated だが、 textarea の native undo / redo 操作を
  // プログラム的にトリガーする最も互換性の高い手段で、 主要ブラウザで動作する。
  const handleUndo = useCallback(() => {
    textareaRef.current?.focus();
    document.execCommand('undo');
  }, []);

  const handleRedo = useCallback(() => {
    textareaRef.current?.focus();
    document.execCommand('redo');
  }, []);

  const handleDownload = useCallback(() => {
    const blob = new Blob([content], { type: 'text/markdown;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${title || '無題'}.md`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, [content, title]);

  // textarea のカーソル位置に text を挿入する（controlled value を更新）。
  const insertAtCursor = useCallback(
    (text: string) => {
      const ta = textareaRef.current;
      if (!ta) {
        onContentChange(content + text);
        return;
      }
      const start = ta.selectionStart ?? content.length;
      const end = ta.selectionEnd ?? content.length;
      onContentChange(content.slice(0, start) + text + content.slice(end));
      requestAnimationFrame(() => {
        ta.focus();
        const pos = start + text.length;
        ta.setSelectionRange(pos, pos);
      });
    },
    [content, onContentChange],
  );

  // 画像をアップロードして ![alt](url) を挿入する。
  const uploadImage = useCallback(
    async (file: File) => {
      if (!onImageUpload) return;
      if (!file.type.startsWith('image/')) {
        showToast('error', '画像ファイルのみアップロードできます');
        return;
      }
      setUploading(true);
      try {
        const url = await onImageUpload(file);
        const alt = file.name.replace(/\.[^.]+$/, '') || 'image';
        insertAtCursor(`\n![${alt}](${url})\n`);
        showToast('success', '画像を挿入しました');
      } catch {
        showToast('error', '画像のアップロードに失敗しました');
      } finally {
        setUploading(false);
      }
    },
    [onImageUpload, showToast, insertAtCursor],
  );

  const handleDrop = useCallback(
    (e: React.DragEvent<HTMLTextAreaElement>) => {
      if (!onImageUpload) return;
      const file = Array.from(e.dataTransfer.files).find((f) => f.type.startsWith('image/'));
      if (file) {
        e.preventDefault();
        uploadImage(file);
      }
    },
    [onImageUpload, uploadImage],
  );

  const handlePaste = useCallback(
    (e: React.ClipboardEvent<HTMLTextAreaElement>) => {
      if (!onImageUpload) return;
      const file = Array.from(e.clipboardData.items)
        .find((i) => i.type.startsWith('image/'))
        ?.getAsFile();
      if (file) {
        e.preventDefault();
        uploadImage(file);
      }
    },
    [onImageUpload, uploadImage],
  );

  return (
    <div className="flex flex-col h-full p-6 max-w-6xl mx-auto w-full">
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
          編集
        </TabButton>
        <TabButton active={tab === 'preview'} onClick={() => setTab('preview')}>
          プレビュー
        </TabButton>
      </div>

      <div className="flex-1 min-h-0 overflow-hidden">
        {tab === 'edit' ? (
          <textarea
            ref={textareaRef}
            value={content}
            onChange={(e) => onContentChange(e.target.value)}
            onDrop={handleDrop}
            onPaste={handlePaste}
            onDragOver={onImageUpload ? (e) => e.preventDefault() : undefined}
            placeholder="# 見出し&#10;&#10;Markdown でノートを書きましょう..."
            spellCheck={false}
            className="w-full h-full resize-none p-4 rounded-lg bg-surface-1 border border-surface-3 text-sm font-mono text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:outline-none focus:border-brand-400/50 transition-colors"
          />
        ) : (
          <div className="w-full h-full overflow-y-auto p-4 rounded-lg bg-surface-1 border border-surface-3 prose prose-sm max-w-none">
            {content.trim() ? (
              <MarkdownView content={content} />
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
        {uploading && (
          <span className="text-xs text-[var(--color-text-muted)]">画像アップロード中...</span>
        )}
        <div className="ml-auto flex items-center gap-1">
          {onImageUpload && (
            <>
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                disabled={uploading}
                aria-label="画像を挿入"
                title="画像を挿入（ドラッグ&ドロップ / 貼り付けも可）"
                className="p-1.5 rounded hover:bg-surface-3 text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] transition-colors disabled:opacity-50"
              >
                <PhotoIcon className="w-4 h-4" />
              </button>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                className="hidden"
                onChange={(e) => {
                  const f = e.target.files?.[0];
                  if (f) uploadImage(f);
                  e.target.value = '';
                }}
              />
              <span className="w-px h-4 bg-surface-3 mx-1" aria-hidden="true" />
            </>
          )}
          <button
            type="button"
            onClick={handleUndo}
            aria-label="元に戻す (Cmd+Z / Ctrl+Z)"
            className="p-1.5 rounded hover:bg-surface-3 text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] transition-colors"
            title="元に戻す (Cmd+Z / Ctrl+Z)"
          >
            <ArrowUturnLeftIcon className="w-4 h-4" />
          </button>
          <button
            type="button"
            onClick={handleRedo}
            aria-label="やり直す (Cmd+Shift+Z / Ctrl+Y)"
            className="p-1.5 rounded hover:bg-surface-3 text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] transition-colors"
            title="やり直す (Cmd+Shift+Z / Ctrl+Y)"
          >
            <ArrowUturnRightIcon className="w-4 h-4" />
          </button>
          <span className="w-px h-4 bg-surface-3 mx-1" aria-hidden="true" />
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
            className="text-brand-400 underline-offset-2 hover:underline"
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
