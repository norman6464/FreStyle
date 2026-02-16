import { useEffect, useMemo, useRef } from 'react';
import { useEditor } from '@tiptap/react';
import { createEditorExtensions } from '../utils/editorExtensions';
import { isLegacyMarkdown } from '../utils/isLegacyMarkdown';
import { markdownToTiptap } from '../utils/markdownToTiptap';

interface UseBlockEditorOptions {
  content: string;
  onChange: (jsonString: string) => void;
}

export function useBlockEditor({ content, onChange }: UseBlockEditorOptions) {
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  // エディタ自身が発行した最後のJSON文字列を記録し、同期ループを防止
  const lastEmittedJson = useRef<string>('');

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
      const json = JSON.stringify(editor.getJSON());
      lastEmittedJson.current = json;
      onChangeRef.current(json);
    },
  });

  useEffect(() => {
    if (!editor) return;

    // エディタのonUpdateから発行されたコンテンツが戻ってきた場合はスキップ
    if (content === lastEmittedJson.current) return;

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

    const currentJson = JSON.stringify(editor.getJSON());
    const newJson = JSON.stringify(newContent);
    if (newJson === currentJson) return;

    editor.commands.setContent(newContent);
  }, [content, editor]);

  return { editor };
}
