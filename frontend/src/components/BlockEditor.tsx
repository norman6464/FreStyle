import { useEffect, useMemo } from 'react';
import { useEditor, EditorContent } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Placeholder from '@tiptap/extension-placeholder';
import Image from '@tiptap/extension-image';
import { SlashCommandExtension } from '../extensions/SlashCommandExtension';
import { slashCommandRenderer } from '../extensions/slashCommandRenderer';
import { isLegacyMarkdown } from '../utils/isLegacyMarkdown';
import { markdownToTiptap } from '../utils/markdownToTiptap';

interface BlockEditorProps {
  content: string;
  onChange: (jsonString: string) => void;
}

export default function BlockEditor({ content, onChange }: BlockEditorProps) {
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
    ],
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

  return (
    <div className="block-editor flex-1 overflow-y-auto" data-testid="block-editor">
      <EditorContent editor={editor} aria-label="ノートの内容" />
    </div>
  );
}
