import { useEffect, useMemo } from 'react';
import { useEditor } from '@tiptap/react';
import { createEditorExtensions } from '../utils/editorExtensions';
import { isLegacyMarkdown } from '../utils/isLegacyMarkdown';
import { markdownToTiptap } from '../utils/markdownToTiptap';

interface UseBlockEditorOptions {
  content: string;
  onChange: (jsonString: string) => void;
}

export function useBlockEditor({ content, onChange }: UseBlockEditorOptions) {
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

  const editor = useEditor({
    extensions: createEditorExtensions(),
    content: initialContent,
    onUpdate: ({ editor }) => {
      onChange(JSON.stringify(editor.getJSON()));
    },
  });

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

  return { editor };
}
