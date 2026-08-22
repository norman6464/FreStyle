import type { ReactNode } from 'react';
import { type Editor, useEditorState } from '@tiptap/react';

/**
 * ToolbarButton はツールバーの 1 ボタン。
 * トグル系（太字〜コードブロック）は pressed を渡し、aria-pressed で押下状態を表す。
 * 非トグル系（水平線・undo・redo）は pressed を渡さず、aria-pressed を付けない
 * （スクリーンリーダに「押下状態を持つボタン」と誤認させないため）。
 * label はスクリーンリーダ用（title と同じ日本語）で、children にアイコン文字を置く。
 */
function ToolbarButton({
  label,
  pressed,
  disabled = false,
  onClick,
  children,
}: {
  label: string;
  pressed?: boolean;
  disabled?: boolean;
  onClick: () => void;
  children: ReactNode;
}) {
  const active = pressed === true;
  return (
    <button
      type="button"
      title={label}
      aria-label={label}
      // pressed が undefined のとき React は aria-pressed 属性自体を出力しない（非トグル）。
      aria-pressed={pressed}
      disabled={disabled}
      // onMouseDown で preventDefault し、ボタン押下でエディタからフォーカスが外れない（＝選択が消えない）ようにする。
      onMouseDown={(mouseEvent) => mouseEvent.preventDefault()}
      onClick={onClick}
      className={[
        'inline-flex h-8 min-w-8 items-center justify-center rounded px-2 text-sm font-medium',
        'transition-colors disabled:cursor-not-allowed disabled:opacity-40',
        active
          ? 'bg-[var(--color-surface-3)] text-[var(--color-text-primary)]'
          : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-2)]',
      ].join(' ')}
    >
      {children}
    </button>
  );
}

function Divider() {
  return <span aria-hidden="true" className="mx-1 h-5 w-px self-center bg-[var(--color-surface-3)]" />;
}

/**
 * RichTextEditorToolbar は RichTextEditor の書式ツールバー。editor の現在状態を
 * useEditorState で購読し、選択位置に応じて active / disabled を更新する。
 */
export default function RichTextEditorToolbar({ editor }: { editor: Editor }) {
  // 選択・内容の変化で必要な派生状態だけを取り出して再描画する（過剰な再描画を避ける）。
  const editorState = useEditorState({
    editor,
    selector: ({ editor: currentEditor }) => ({
      bold: currentEditor.isActive('bold'),
      italic: currentEditor.isActive('italic'),
      underline: currentEditor.isActive('underline'),
      strike: currentEditor.isActive('strike'),
      code: currentEditor.isActive('code'),
      h1: currentEditor.isActive('heading', { level: 1 }),
      h2: currentEditor.isActive('heading', { level: 2 }),
      h3: currentEditor.isActive('heading', { level: 3 }),
      bulletList: currentEditor.isActive('bulletList'),
      orderedList: currentEditor.isActive('orderedList'),
      blockquote: currentEditor.isActive('blockquote'),
      codeBlock: currentEditor.isActive('codeBlock'),
      canUndo: currentEditor.can().undo(),
      canRedo: currentEditor.can().redo(),
    }),
  });

  const chain = () => editor.chain().focus();

  return (
    <div
      role="toolbar"
      aria-label="書式ツールバー"
      className="flex flex-wrap items-center gap-0.5 border-b border-[var(--color-surface-3)] px-2 py-1.5"
    >
      <ToolbarButton label="太字" pressed={editorState.bold} onClick={() => chain().toggleBold().run()}>
        <span className="font-bold">B</span>
      </ToolbarButton>
      <ToolbarButton label="斜体" pressed={editorState.italic} onClick={() => chain().toggleItalic().run()}>
        <span className="italic">I</span>
      </ToolbarButton>
      <ToolbarButton label="下線" pressed={editorState.underline} onClick={() => chain().toggleUnderline().run()}>
        <span className="underline">U</span>
      </ToolbarButton>
      <ToolbarButton label="打ち消し線" pressed={editorState.strike} onClick={() => chain().toggleStrike().run()}>
        <span className="line-through">S</span>
      </ToolbarButton>
      <ToolbarButton label="インラインコード" pressed={editorState.code} onClick={() => chain().toggleCode().run()}>
        <span className="font-mono">{'</>'}</span>
      </ToolbarButton>

      <Divider />

      <ToolbarButton label="見出し1" pressed={editorState.h1} onClick={() => chain().toggleHeading({ level: 1 }).run()}>
        H1
      </ToolbarButton>
      <ToolbarButton label="見出し2" pressed={editorState.h2} onClick={() => chain().toggleHeading({ level: 2 }).run()}>
        H2
      </ToolbarButton>
      <ToolbarButton label="見出し3" pressed={editorState.h3} onClick={() => chain().toggleHeading({ level: 3 }).run()}>
        H3
      </ToolbarButton>

      <Divider />

      <ToolbarButton label="箇条書き" pressed={editorState.bulletList} onClick={() => chain().toggleBulletList().run()}>
        •
      </ToolbarButton>
      <ToolbarButton
        label="番号付きリスト"
        pressed={editorState.orderedList}
        onClick={() => chain().toggleOrderedList().run()}
      >
        1.
      </ToolbarButton>
      <ToolbarButton label="引用" pressed={editorState.blockquote} onClick={() => chain().toggleBlockquote().run()}>
        <span className="font-serif">&ldquo;</span>
      </ToolbarButton>
      <ToolbarButton
        label="コードブロック"
        pressed={editorState.codeBlock}
        onClick={() => chain().toggleCodeBlock().run()}
      >
        <span className="font-mono text-xs">{'{ }'}</span>
      </ToolbarButton>
      <ToolbarButton label="水平線" onClick={() => chain().setHorizontalRule().run()}>
        —
      </ToolbarButton>

      <Divider />

      <ToolbarButton label="元に戻す" disabled={!editorState.canUndo} onClick={() => chain().undo().run()}>
        ↺
      </ToolbarButton>
      <ToolbarButton label="やり直す" disabled={!editorState.canRedo} onClick={() => chain().redo().run()}>
        ↻
      </ToolbarButton>
    </div>
  );
}
