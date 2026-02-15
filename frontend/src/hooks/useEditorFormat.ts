import { useCallback } from 'react';
import type { Editor } from '@tiptap/react';

export interface EditorFormatHandlers {
  handleBold: () => void;
  handleItalic: () => void;
  handleUnderline: () => void;
  handleStrike: () => void;
  handleAlign: (alignment: 'left' | 'center' | 'right') => void;
  handleSelectColor: (color: string) => void;
  handleHighlight: (color: string) => void;
  handleSuperscript: () => void;
  handleSubscript: () => void;
  handleUndo: () => void;
  handleRedo: () => void;
  handleClearFormatting: () => void;
  handleIndent: () => void;
  handleOutdent: () => void;
  handleBlockquote: () => void;
  handleHorizontalRule: () => void;
  handleCodeBlock: () => void;
  handleBulletList: () => void;
  handleOrderedList: () => void;
  handleInlineCode: () => void;
  handleHeading: (level: number) => void;
  handleTaskList: () => void;
  handleInsertTable: () => void;
}

export function useEditorFormat(editor: Editor | null): EditorFormatHandlers {
  const handleBold = useCallback(() => {
    editor?.chain().focus().toggleBold().run();
  }, [editor]);

  const handleItalic = useCallback(() => {
    editor?.chain().focus().toggleItalic().run();
  }, [editor]);

  const handleUnderline = useCallback(() => {
    editor?.chain().focus().toggleUnderline().run();
  }, [editor]);

  const handleStrike = useCallback(() => {
    editor?.chain().focus().toggleStrike().run();
  }, [editor]);

  const handleAlign = useCallback((alignment: 'left' | 'center' | 'right') => {
    if (!editor) return;
    editor.chain().focus().setTextAlign(alignment).run();
  }, [editor]);

  const handleSelectColor = useCallback((color: string) => {
    if (!editor) return;
    if (color) {
      editor.chain().focus().setColor(color).run();
    } else {
      editor.chain().focus().unsetColor().run();
    }
  }, [editor]);

  const handleHighlight = useCallback((color: string) => {
    if (!editor) return;
    if (color) {
      editor.chain().focus().setHighlight({ color }).run();
    } else {
      editor.chain().focus().unsetHighlight().run();
    }
  }, [editor]);

  const handleSuperscript = useCallback(() => {
    editor?.chain().focus().toggleSuperscript().run();
  }, [editor]);

  const handleSubscript = useCallback(() => {
    editor?.chain().focus().toggleSubscript().run();
  }, [editor]);

  const handleUndo = useCallback(() => {
    editor?.chain().focus().undo().run();
  }, [editor]);

  const handleRedo = useCallback(() => {
    editor?.chain().focus().redo().run();
  }, [editor]);

  const handleClearFormatting = useCallback(() => {
    editor?.chain().focus().unsetAllMarks().clearNodes().run();
  }, [editor]);

  const handleIndent = useCallback(() => {
    if (!editor) return;
    if (editor.can().sinkListItem('listItem')) {
      editor.chain().focus().sinkListItem('listItem').run();
    } else if (editor.can().sinkListItem('taskItem')) {
      editor.chain().focus().sinkListItem('taskItem').run();
    }
  }, [editor]);

  const handleOutdent = useCallback(() => {
    if (!editor) return;
    if (editor.can().liftListItem('listItem')) {
      editor.chain().focus().liftListItem('listItem').run();
    } else if (editor.can().liftListItem('taskItem')) {
      editor.chain().focus().liftListItem('taskItem').run();
    }
  }, [editor]);

  const handleBlockquote = useCallback(() => {
    editor?.chain().focus().toggleBlockquote().run();
  }, [editor]);

  const handleHorizontalRule = useCallback(() => {
    editor?.chain().focus().setHorizontalRule().run();
  }, [editor]);

  const handleCodeBlock = useCallback(() => {
    editor?.chain().focus().toggleCodeBlock().run();
  }, [editor]);

  const handleBulletList = useCallback(() => {
    editor?.chain().focus().toggleBulletList().run();
  }, [editor]);

  const handleOrderedList = useCallback(() => {
    editor?.chain().focus().toggleOrderedList().run();
  }, [editor]);

  const handleInlineCode = useCallback(() => {
    editor?.chain().focus().toggleCode().run();
  }, [editor]);

  const handleTaskList = useCallback(() => {
    editor?.chain().focus().toggleTaskList().run();
  }, [editor]);

  const handleInsertTable = useCallback(() => {
    editor?.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run();
  }, [editor]);

  const handleHeading = useCallback((level: number) => {
    if (!editor) return;
    if (level === 0) {
      editor.chain().focus().setParagraph().run();
    } else {
      editor.chain().focus().toggleHeading({ level: level as 1 | 2 | 3 }).run();
    }
  }, [editor]);

  return { handleBold, handleItalic, handleUnderline, handleStrike, handleAlign, handleSelectColor, handleHighlight, handleSuperscript, handleSubscript, handleUndo, handleRedo, handleClearFormatting, handleIndent, handleOutdent, handleBlockquote, handleHorizontalRule, handleCodeBlock, handleBulletList, handleOrderedList, handleInlineCode, handleHeading, handleTaskList, handleInsertTable };
}
