import { useCallback, useRef, useState } from 'react';
import { EditorContent } from '@tiptap/react';
import 'tippy.js/dist/tippy.css';
import { executeCommand } from '../extensions/SlashCommandExtension';
import { useBlockEditor } from '../hooks/useBlockEditor';
import { useImageUpload } from '../hooks/useImageUpload';
import { useLinkEditor } from '../hooks/useLinkEditor';
import { useEditorFormat } from '../hooks/useEditorFormat';
import BlockInserterButton from './BlockInserterButton';
import LinkBubbleMenu from './LinkBubbleMenu';
import EditorToolbar from './EditorToolbar';
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
  const { linkBubble, handleEditorClick, handleEditLink, handleRemoveLink } = useLinkEditor(editor, containerRef);
  const { handleBold, handleItalic, handleUnderline, handleStrike, handleAlign, handleSelectColor, handleHighlight, handleSuperscript, handleSubscript, handleUndo, handleRedo, handleClearFormatting, handleIndent, handleOutdent, handleBlockquote } = useEditorFormat(editor);

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
      <EditorToolbar
        onBold={handleBold}
        onItalic={handleItalic}
        onUnderline={handleUnderline}
        onStrike={handleStrike}
        onSuperscript={handleSuperscript}
        onSubscript={handleSubscript}
        onSelectColor={handleSelectColor}
        onHighlight={handleHighlight}
        onAlign={handleAlign}
        onUndo={handleUndo}
        onRedo={handleRedo}
        onClearFormatting={handleClearFormatting}
        onIndent={handleIndent}
        onOutdent={handleOutdent}
        onBlockquote={handleBlockquote}
      />
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
      />
      <div className="pl-8">
        <EditorContent editor={editor} aria-label="ノートの内容" />
      </div>
    </div>
  );
}
