import { Node, mergeAttributes, type RawCommands, type CommandProps } from '@tiptap/core';
import { ReactNodeViewRenderer, type ReactNodeViewProps } from '@tiptap/react';
import type { ComponentType } from 'react';
import CalloutNodeView from '../components/CalloutNodeView';

export type CalloutType = 'info' | 'warning' | 'error' | 'success';

export const Callout = Node.create({
  name: 'callout',
  group: 'block',
  content: 'block+',
  defining: true,

  addAttributes() {
    return {
      type: {
        default: 'info' as CalloutType,
        parseHTML: (element) => element.getAttribute('data-callout-type') || 'info',
        renderHTML: (attributes) => ({ 'data-callout-type': attributes.type }),
      },
      emoji: {
        default: '💡',
        parseHTML: (element) => element.getAttribute('data-callout-emoji') || '💡',
        renderHTML: (attributes) => ({ 'data-callout-emoji': attributes.emoji }),
      },
    };
  },

  parseHTML() {
    return [{ tag: 'div[data-callout]' }];
  },

  renderHTML({ HTMLAttributes }) {
    return ['div', mergeAttributes(HTMLAttributes, { 'data-callout': '', class: 'callout' }), 0];
  },

  addNodeView() {
    // CalloutNodeView は ReactNodeViewProps<HTMLElement> を拡張した形状で型付けされているが、
    // ReactNodeViewRenderer は厳密な ComponentType を要求するためここで cast する。
    return ReactNodeViewRenderer(
      CalloutNodeView as ComponentType<ReactNodeViewProps<HTMLElement>>,
    );
  },

  addCommands(): Partial<RawCommands> {
    return {
      setCallout:
        () =>
        ({ commands }: CommandProps) => {
          return commands.insertContent({
            type: 'callout',
            attrs: { type: 'info', emoji: '💡' },
            content: [{ type: 'paragraph' }],
          });
        },
      setCalloutWithType:
        (calloutType: string, emoji: string) =>
        ({ commands }: CommandProps) => {
          return commands.insertContent({
            type: 'callout',
            attrs: { type: calloutType, emoji },
            content: [{ type: 'paragraph' }],
          });
        },
    };
  },
});
