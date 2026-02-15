import { useCallback, useEffect, useRef, useState } from 'react';
import { EditorContent } from '@tiptap/react';
import 'tippy.js/dist/tippy.css';
import { executeCommand } from '../extensions/SlashCommandExtension';
import { useBlockEditor } from '../hooks/useBlockEditor';
import { useImageUpload } from '../hooks/useImageUpload';
import BlockInserterButton from './BlockInserterButton';
import LinkBubbleMenu from './LinkBubbleMenu';
import type { SlashCommand } from '../constants/slashCommands';

interface BlockEditorProps {
  content: string;
  onChange: (jsonString: string) => void;
  noteId: string | null;
}

export default function BlockEditor({ content, onChange, noteId }: BlockEditorProps) {
  const { editor } = useBlockEditor({ content, onChange });

  const containerRef = useRef<HTMLDivElement>(null);
  const [inserterVisible, setInserterVisible] = useState(false);
  const [inserterTop, setInserterTop] = useState(0);

  const { openFileDialog, handleDrop, handlePaste } = useImageUpload(noteId, editor);

  const [linkBubble, setLinkBubble] = useState<{ url: string; top: number; left: number } | null>(null);

  // Ctrl+K / Cmd+K でリンク挿入
  useEffect(() => {
    if (!editor) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        const previousUrl = editor.getAttributes('link').href || '';
        const url = window.prompt('URLを入力', previousUrl);
        if (url === null) return;
        if (url === '') {
          editor.chain().focus().extendMarkRange('link').unsetLink().run();
        } else {
          editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run();
        }
      }
    };
    const editorEl = editor.view.dom;
    editorEl.addEventListener('keydown', handleKeyDown);
    return () => editorEl.removeEventListener('keydown', handleKeyDown);
  }, [editor]);

  // リンクホバー検知
  const handleEditorClick = useCallback((e: React.MouseEvent) => {
    if (!editor || !containerRef.current) return;
    const target = e.target as HTMLElement;
    const linkEl = target.closest('a.note-link');
    if (linkEl) {
      const href = linkEl.getAttribute('href') || '';
      const containerRect = containerRef.current.getBoundingClientRect();
      const linkRect = linkEl.getBoundingClientRect();
      setLinkBubble({
        url: href,
        top: linkRect.bottom - containerRect.top + 4,
        left: linkRect.left - containerRect.left,
      });
    } else {
      setLinkBubble(null);
    }
  }, [editor]);

  const lastMoveTime = useRef(0);
  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    const now = Date.now();
    if (now - lastMoveTime.current < 50) return;
    lastMoveTime.current = now;

    if (!containerRef.current) return;
    const editorEl = containerRef.current.querySelector('.ProseMirror');
    if (!editorEl) return;

    const target = e.target as HTMLElement;
    const block = target.closest('p, h1, h2, h3, ul, ol, li, blockquote');
    if (!block || !editorEl.contains(block)) {
      setInserterVisible(false);
      return;
    }

    const containerRect = containerRef.current.getBoundingClientRect();
    const blockRect = block.getBoundingClientRect();
    setInserterTop(blockRect.top - containerRect.top);
    setInserterVisible(true);
  }, []);

  const handleMouseLeave = useCallback(() => {
    setInserterVisible(false);
  }, []);

  const handleCommand = useCallback((command: SlashCommand) => {
    if (!editor) return;
    if (command.action === 'image') {
      openFileDialog();
      return;
    }
    editor.chain().focus().run();
    executeCommand(editor, command);
  }, [editor, openFileDialog]);

  return (
    <div
      ref={containerRef}
      className="block-editor flex-1 overflow-y-auto relative"
      data-testid="block-editor"
      onMouseMove={handleMouseMove}
      onMouseLeave={handleMouseLeave}
      onDrop={handleDrop}
      onPaste={handlePaste}
      onDragOver={(e) => e.preventDefault()}
      onClick={handleEditorClick}
    >
      {linkBubble && (
        <div
          className="absolute z-50"
          style={{ top: linkBubble.top, left: linkBubble.left }}
        >
          <LinkBubbleMenu
            url={linkBubble.url}
            onEdit={() => {
              if (!editor) return;
              const url = window.prompt('URLを入力', linkBubble.url);
              if (url === null) return;
              if (url === '') {
                editor.chain().focus().extendMarkRange('link').unsetLink().run();
              } else {
                editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run();
              }
              setLinkBubble(null);
            }}
            onRemove={() => {
              if (!editor) return;
              editor.chain().focus().extendMarkRange('link').unsetLink().run();
              setLinkBubble(null);
            }}
          />
        </div>
      )}
      <BlockInserterButton
        visible={inserterVisible}
        top={inserterTop}
        onCommand={handleCommand}
      />
      <div className="pl-8">
        <EditorContent editor={editor} aria-label="ノートの内容" />
      </div>
    </div>
  );
}
