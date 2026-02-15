import { useCallback, useEffect, useState } from 'react';
import type { Editor } from '@tiptap/core';

export interface LinkBubbleState {
  url: string;
  top: number;
  left: number;
}

export function useLinkEditor(
  editor: Editor | null,
  containerRef: React.RefObject<HTMLDivElement | null>,
) {
  const [linkBubble, setLinkBubble] = useState<LinkBubbleState | null>(null);

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

  // リンククリック検知
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
  }, [editor, containerRef]);

  // リンク編集
  const handleEditLink = useCallback(() => {
    if (!editor || !linkBubble) return;
    const url = window.prompt('URLを入力', linkBubble.url);
    if (url === null) return;
    if (url === '') {
      editor.chain().focus().extendMarkRange('link').unsetLink().run();
    } else {
      editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run();
    }
    setLinkBubble(null);
  }, [editor, linkBubble]);

  // リンク削除
  const handleRemoveLink = useCallback(() => {
    if (!editor) return;
    editor.chain().focus().extendMarkRange('link').unsetLink().run();
    setLinkBubble(null);
  }, [editor]);

  const dismissLinkBubble = useCallback(() => {
    setLinkBubble(null);
  }, []);

  return {
    linkBubble,
    handleEditorClick,
    handleEditLink,
    handleRemoveLink,
    dismissLinkBubble,
  };
}
