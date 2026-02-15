import { Node, mergeAttributes } from '@tiptap/core';
import { ReactNodeViewRenderer } from '@tiptap/react';
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
    return ReactNodeViewRenderer(CalloutNodeView);
  },

  addCommands() {
    return {
      setCallout: () => ({ commands }) => {
        return commands.insertContent({
          type: 'callout',
          attrs: { type: 'info', emoji: '💡' },
          content: [{ type: 'paragraph' }],
        });
      },
      setCalloutWithType: (calloutType: CalloutType, emoji: string) => ({ commands }) => {
        return commands.insertContent({
          type: 'callout',
          attrs: { type: calloutType, emoji },
          content: [{ type: 'paragraph' }],
        });
      },
    };
  },
});
