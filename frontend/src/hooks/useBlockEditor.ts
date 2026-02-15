import { useEffect, useMemo } from 'react';
import { useEditor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Placeholder from '@tiptap/extension-placeholder';
import Image from '@tiptap/extension-image';
import TaskList from '@tiptap/extension-task-list';
import TaskItem from '@tiptap/extension-task-item';
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight';
import Highlight from '@tiptap/extension-highlight';
import Table from '@tiptap/extension-table';
import TableRow from '@tiptap/extension-table-row';
import TableCell from '@tiptap/extension-table-cell';
import TableHeader from '@tiptap/extension-table-header';
import { common, createLowlight } from 'lowlight';
import { SlashCommandExtension } from '../extensions/SlashCommandExtension';
import { slashCommandRenderer } from '../extensions/slashCommandRenderer';
import { ToggleList, ToggleSummary, ToggleContent } from '../extensions/ToggleListExtension';
import { Callout } from '../extensions/CalloutExtension';
import Link from '@tiptap/extension-link';
import TextStyle from '@tiptap/extension-text-style';
import Color from '@tiptap/extension-color';
import TextAlign from '@tiptap/extension-text-align';
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
    extensions: [
      StarterKit.configure({
        heading: { levels: [1, 2, 3] },
        codeBlock: false,
      }),
      CodeBlockLowlight.configure({
        lowlight: createLowlight(common),
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
      Highlight.configure({ multicolor: true }),
      TaskList,
      TaskItem.configure({ nested: true }),
      ToggleList,
      ToggleSummary,
      ToggleContent,
      Table.configure({ resizable: true }),
      TableRow,
      TableCell,
      TableHeader,
      Callout,
      Link.configure({
        openOnClick: false,
        HTMLAttributes: { class: 'note-link' },
      }),
      TextStyle,
      Color,
      TextAlign.configure({
        types: ['heading', 'paragraph'],
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

  return { editor };
}
