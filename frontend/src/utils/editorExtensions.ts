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
import Underline from '@tiptap/extension-underline';
import Superscript from '@tiptap/extension-superscript';
import Subscript from '@tiptap/extension-subscript';
import Youtube from '@tiptap/extension-youtube';

export function createEditorExtensions() {
  return [
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
    Underline,
    Superscript,
    Subscript,
    Youtube.configure({
      allowFullscreen: true,
      HTMLAttributes: { class: 'note-youtube' },
    }),
  ];
}
