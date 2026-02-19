import StarterKit from '@tiptap/starter-kit';
import Heading from '@tiptap/extension-heading';
import { textblockTypeInputRule } from '@tiptap/core';
import Placeholder from '@tiptap/extension-placeholder';
import Image from '@tiptap/extension-image';
import TaskList from '@tiptap/extension-task-list';
import TaskItem from '@tiptap/extension-task-item';
import Highlight from '@tiptap/extension-highlight';
import { Table } from '@tiptap/extension-table';
import { TableRow } from '@tiptap/extension-table-row';
import { TableCell } from '@tiptap/extension-table-cell';
import { TableHeader } from '@tiptap/extension-table-header';
import { common, createLowlight } from 'lowlight';
import { SlashCommandExtension } from '../extensions/SlashCommandExtension';
import { slashCommandRenderer } from '../extensions/slashCommandRenderer';
import { ToggleList, ToggleSummary, ToggleContent } from '../extensions/ToggleListExtension';
import { Callout } from '../extensions/CalloutExtension';
import { CodeBlock } from '../extensions/CodeBlockExtension';
import Link from '@tiptap/extension-link';
import { TextStyle } from '@tiptap/extension-text-style';
import Color from '@tiptap/extension-color';
import TextAlign from '@tiptap/extension-text-align';
import Underline from '@tiptap/extension-underline';
import Superscript from '@tiptap/extension-superscript';
import Subscript from '@tiptap/extension-subscript';
import Youtube from '@tiptap/extension-youtube';
import { SearchReplaceExtension } from '../extensions/SearchReplaceExtension';
import { FullWidthHeadingEnter } from '../extensions/FullWidthHeadingEnter';

export function createEditorExtensions() {
  return [
    StarterKit.configure({
      heading: false,
      codeBlock: false,
    }),
    Heading.configure({ levels: [1, 2, 3] }).extend({
      addInputRules() {
        return (this.options.levels as number[]).map((level) => {
          return textblockTypeInputRule({
            find: new RegExp(`^([#＃]{${Math.min(...(this.options.levels as number[]))},${level}})\\s$`),
            type: this.type,
            getAttributes: { level },
          });
        });
      },
    }),
    CodeBlock.configure({
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
    SearchReplaceExtension,
    FullWidthHeadingEnter,
  ];
}
