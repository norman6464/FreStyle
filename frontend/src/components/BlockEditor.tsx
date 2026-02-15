import { useCallback, useEffect, useRef, useState } from 'react';
import { EditorContent } from '@tiptap/react';
import 'tippy.js/dist/tippy.css';
import { executeCommand } from '../extensions/SlashCommandExtension';
import { useBlockEditor } from '../hooks/useBlockEditor';
import { useImageUpload } from '../hooks/useImageUpload';
import { useLinkEditor } from '../hooks/useLinkEditor';
import BlockInserterButton from './BlockInserterButton';
import LinkBubbleMenu from './LinkBubbleMenu';
import SelectionToolbar from './SelectionToolbar';
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
  const inserterMenuOpen = useRef(false);

  const { openFileDialog, handleDrop, handlePaste } = useImageUpload(noteId, editor);
  const { linkBubble, handleEditorClick, handleEditLink, handleRemoveLink } = useLinkEditor(editor, containerRef);

  // スラッシュコマンドから画像アップロードを呼べるようにする
  useEffect(() => {
    if (!editor) return;
    editor.storage.slashCommand.onImageUpload = openFileDialog;
    return () => { editor.storage.slashCommand.onImageUpload = null; };
  }, [editor, openFileDialog]);

  const lastMoveTime = useRef(0);
  const hideTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearHideTimeout = useCallback(() => {
    if (hideTimeout.current) {
      clearTimeout(hideTimeout.current);
      hideTimeout.current = null;
    }
  }, []);

  const scheduleHide = useCallback(() => {
    clearHideTimeout();
    hideTimeout.current = setTimeout(() => {
      if (!inserterMenuOpen.current) {
        setInserterVisible(false);
      }
    }, 200);
  }, [clearHideTimeout]);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    const now = Date.now();
    if (now - lastMoveTime.current < 50) return;
    lastMoveTime.current = now;

    if (!containerRef.current) return;
    const editorEl = containerRef.current.querySelector('.ProseMirror');
    if (!editorEl) return;

    const target = e.target as HTMLElement;

    // +ボタン上にマウスがある場合は表示状態を維持
    if (target.closest('[data-block-inserter]')) {
      clearHideTimeout();
      return;
    }

    const block = target.closest('p, h1, h2, h3, ul, ol, li, blockquote');
    if (!block || !editorEl.contains(block)) {
      scheduleHide();
      return;
    }

    clearHideTimeout();
    const containerRect = containerRef.current.getBoundingClientRect();
    const blockRect = block.getBoundingClientRect();
    setInserterTop(blockRect.top - containerRect.top);
    setInserterVisible(true);
  }, [clearHideTimeout, scheduleHide]);

  const handleMouseLeave = useCallback(() => {
    if (!inserterMenuOpen.current) {
      scheduleHide();
    }
  }, [scheduleHide]);

  const handleInserterMenuOpenChange = useCallback((open: boolean) => {
    inserterMenuOpen.current = open;
    if (!open) {
      setInserterVisible(false);
    }
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
      <SelectionToolbar editor={editor} containerRef={containerRef} />
      {linkBubble && (
        <div
          className="absolute z-50"
          style={{ top: linkBubble.top, left: linkBubble.left }}
        >
          <LinkBubbleMenu
            url={linkBubble.url}
            onEdit={handleEditLink}
            onRemove={handleRemoveLink}
          />
        </div>
      )}
      <BlockInserterButton
        visible={inserterVisible}
        top={inserterTop}
        onCommand={handleCommand}
        onMenuOpenChange={handleInserterMenuOpenChange}
      />
      <div className="pl-8">
        <EditorContent editor={editor} aria-label="ノートの内容" />
      </div>
    </div>
  );
}
