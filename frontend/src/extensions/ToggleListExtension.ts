import { Node, mergeAttributes, type RawCommands, type CommandProps } from '@tiptap/core';
import { ReactNodeViewRenderer, type ReactNodeViewProps } from '@tiptap/react';
import type { ComponentType } from 'react';
import ToggleListNodeView from '../components/ToggleListNodeView';

export const ToggleSummary = Node.create({
  name: 'toggleSummary',
  content: 'inline*',
  defining: true,
  selectable: false,

  parseHTML() {
    return [{ tag: 'summary' }];
  },

  renderHTML({ HTMLAttributes }) {
    return ['summary', mergeAttributes(HTMLAttributes), 0];
  },
});

export const ToggleContent = Node.create({
  name: 'toggleContent',
  content: 'block+',
  defining: true,
  selectable: false,

  parseHTML() {
    return [{ tag: 'div[data-toggle-content]' }];
  },

  renderHTML({ HTMLAttributes }) {
    return ['div', mergeAttributes(HTMLAttributes, { 'data-toggle-content': '' }), 0];
  },
});

export const ToggleList = Node.create({
  name: 'toggleList',
  group: 'block',
  content: 'toggleSummary toggleContent',
  defining: true,

  addAttributes() {
    return {
      open: {
        default: true,
        parseHTML: (element) => element.hasAttribute('open'),
        renderHTML: (attributes) => {
          if (!attributes.open) return {};
          return { open: '' };
        },
      },
    };
  },

  parseHTML() {
    return [{ tag: 'details' }];
  },

  renderHTML({ HTMLAttributes }) {
    return ['details', mergeAttributes(HTMLAttributes, { class: 'toggle-list' }), 0];
  },

  addNodeView() {
    // ToggleListNodeView は ReactNodeViewProps<HTMLElement> 拡張型なので、
    // ReactNodeViewRenderer の厳密シグネチャに合わせて cast する。
    return ReactNodeViewRenderer(
      ToggleListNodeView as ComponentType<ReactNodeViewProps<HTMLElement>>,
    );
  },

  addCommands(): Partial<RawCommands> {
    return {
      setToggleList:
        () =>
        ({ commands }: CommandProps) => {
          return commands.insertContent({
            type: 'toggleList',
            attrs: { open: true },
            content: [
              { type: 'toggleSummary', content: [{ type: 'text', text: 'トグル' }] },
              { type: 'toggleContent', content: [{ type: 'paragraph' }] },
            ],
          });
        },
    };
  },
});
