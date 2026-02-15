import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useEditor, EditorContent } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Placeholder from '@tiptap/extension-placeholder';
import Image from '@tiptap/extension-image';
import TaskList from '@tiptap/extension-task-list';
import TaskItem from '@tiptap/extension-task-item';
import 'tippy.js/dist/tippy.css';
import { SlashCommandExtension, executeCommand } from '../extensions/SlashCommandExtension';
import { slashCommandRenderer } from '../extensions/slashCommandRenderer';
import { ToggleList, ToggleSummary, ToggleContent } from '../extensions/ToggleListExtension';
import { isLegacyMarkdown } from '../utils/isLegacyMarkdown';
import { markdownToTiptap } from '../utils/markdownToTiptap';
import { useImageUpload } from '../hooks/useImageUpload';
import BlockInserterButton from './BlockInserterButton';
import type { SlashCommand } from '../constants/slashCommands';

interface BlockEditorProps {
  content: string;
  onChange: (jsonString: string) => void;
  noteId: string | null;
}

export default function BlockEditor({ content, onChange, noteId }: BlockEditorProps) {
  const initialContent = useMemo(() => {
    if (!content) return undefined;
    if (isLegacyMarkdown(content)) {
      return markdownToTiptap(content);
    }
    try {
      return JSON.parse(content);
    } catch {
      return undefined;
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const containerRef = useRef<HTMLDivElement>(null);
  const [inserterVisible, setInserterVisible] = useState(false);
  const [inserterTop, setInserterTop] = useState(0);

  const editor = useEditor({
    extensions: [
      StarterKit.configure({
        heading: { levels: [1, 2, 3] },
      }),
      Placeholder.configure({
        placeholder: 'ここに入力...',
      }),
      Image.configure({
        allowBase64: false,
        HTMLAttributes: { class: 'note-image' },
      }),
      SlashCommandExtension.configure({
        suggestion: {
          render: slashCommandRenderer,
        },
      }),
      TaskList,
      TaskItem.configure({ nested: true }),
      ToggleList,
      ToggleSummary,
      ToggleContent,
    ],
    content: initialContent,
    onUpdate: ({ editor }) => {
      onChange(JSON.stringify(editor.getJSON()));
    },
  });

  const { openFileDialog, handleDrop, handlePaste } = useImageUpload(noteId, editor);

  useEffect(() => {
    if (!editor) return;
    const currentJson = JSON.stringify(editor.getJSON());
    if (content === currentJson) return;

    if (!content) {
      editor.commands.clearContent();
      return;
    }

    let newContent;
    if (isLegacyMarkdown(content)) {
      newContent = markdownToTiptap(content);
    } else {
      try {
        newContent = JSON.parse(content);
      } catch {
        return;
      }
    }

    const newJson = JSON.stringify(newContent);
    if (newJson === currentJson) return;

    editor.commands.setContent(newContent);
  }, [content, editor]);

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
    >
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
