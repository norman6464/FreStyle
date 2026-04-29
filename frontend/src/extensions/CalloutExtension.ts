import { Node, mergeAttributes, type RawCommands, type CommandProps } from '@tiptap/core';
import { ReactNodeViewRenderer, type ReactNodeViewProps } from '@tiptap/react';
import type { ComponentType } from 'react';
import container from 'markdown-it-container';

// tiptap-markdown が markdown-it インスタンスを setup() に渡す。
// markdown-it 14.x の型は CommonJS 由来で namespace アクセスが面倒なため、
// このファイルで使うのに必要最小限の構造を local type で定義する。
interface MdToken {
  nesting: 1 | 0 | -1;
  info: string;
  type: string;
}

interface MarkdownItType {
  use: (plugin: unknown, ...params: unknown[]) => MarkdownItType;
}
import CalloutNodeView from '../components/CalloutNodeView';

export type CalloutType = 'info' | 'warning' | 'error' | 'success';

/**
 * Zenn の `:::message` / `:::message alert` を CalloutExtension にマッピングする。
 * info  ↔ :::message
 * error ↔ :::message alert
 * warning / success は info にフォールバック（Zenn 標準が info / alert の 2 種のため）。
 */
type ZennCalloutVariant = 'info' | 'alert';

function zennVariantFromCalloutType(t: CalloutType): ZennCalloutVariant {
  return t === 'error' || t === 'warning' ? 'alert' : 'info';
}

function calloutTypeFromZennVariant(v: ZennCalloutVariant): CalloutType {
  return v === 'alert' ? 'error' : 'info';
}

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

  /**
   * tiptap-markdown 連携。Zenn `:::message` / `:::message alert` 記法を
   * Callout ノードと相互変換する。
   *
   * - serialize: Tiptap callout → markdown `:::message [alert]\n...content...\n:::`
   * - parse:     markdown-it-container plugin で `:::message` をパースし、Tiptap の
   *              callout node に変換する HTML token を発行
   */
  addStorage() {
    return {
      markdown: {
        serialize(state: import('prosemirror-markdown').MarkdownSerializerState, node: import('prosemirror-model').Node) {
          const variant = zennVariantFromCalloutType((node.attrs.type as CalloutType) || 'info');
          state.write(variant === 'alert' ? ':::message alert\n' : ':::message\n');
          state.renderContent(node);
          state.ensureNewLine();
          state.write(':::');
          state.closeBlock(node);
        },
        parse: {
          setup(this: { editor: { schema: { nodes: { callout?: import('prosemirror-model').NodeType } } } }, md: MarkdownItType) {
            md.use(container, 'message', {
              validate: (params: string) => /^message(\s+alert)?\s*$/.test(params.trim()),
              render: (tokens: MdToken[], idx: number) => {
                const token = tokens[idx];
                if (token.nesting === 1) {
                  const params = token.info.trim();
                  const variant: ZennCalloutVariant = /\balert\b/.test(params) ? 'alert' : 'info';
                  const calloutType = calloutTypeFromZennVariant(variant);
                  const emoji = variant === 'alert' ? '⚠️' : '💡';
                  return `<div data-callout="" data-callout-type="${calloutType}" data-callout-emoji="${emoji}" class="callout">\n`;
                }
                return '</div>\n';
              },
            });
          },
        },
      },
    };
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
