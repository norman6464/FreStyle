import { useCallback } from 'react';
import type { Editor } from '@tiptap/react';

export function useEditorFormat(editor: Editor | null) {
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

  return { handleBold, handleItalic, handleUnderline, handleStrike, handleAlign, handleSelectColor, handleHighlight, handleSuperscript, handleSubscript, handleUndo, handleRedo };
}
