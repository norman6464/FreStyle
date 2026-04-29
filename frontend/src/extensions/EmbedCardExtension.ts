/**
 * Zenn 互換の埋め込みカード extension。
 *
 * 仕様:
 *   - markdown 入力で `@[card](https://example.com)` を Tiptap embedCard node に変換
 *   - URL 単独行 (前後が空行) も embedCard 候補だが、Tiptap コア仕様との衝突を避けて
 *     Phase 1 では `@[card](...)` のみサポート (URL 単独行は Phase 2 で対応)
 *   - serialize: embedCard → `@[card](url)` に書き戻し
 *   - NodeView: EmbedCardNodeView (Go backend の /api/v2/embeds/oembed を呼ぶ)
 */
import { Node, mergeAttributes, type RawCommands, type CommandProps } from '@tiptap/core';
import { ReactNodeViewRenderer, type ReactNodeViewProps } from '@tiptap/react';
import type { ComponentType } from 'react';
import type { MarkdownSerializerState } from 'prosemirror-markdown';
import type { Node as PMNode } from 'prosemirror-model';
import EmbedCardNodeView from '../components/EmbedCardNodeView';

interface MarkdownItType {
  inline: { ruler: { before: (existing: string, name: string, fn: unknown) => void } };
}

interface MdToken {
  type: string;
  meta?: { url?: string };
  block?: boolean;
}

export const EmbedCard = Node.create({
  name: 'embedCard',
  group: 'block',
  atom: true,
  selectable: true,

  addAttributes() {
    return {
      url: {
        default: '',
        parseHTML: (element) => element.getAttribute('data-embed-url') || '',
        renderHTML: (attributes) => ({ 'data-embed-url': attributes.url }),
      },
    };
  },

  parseHTML() {
    return [{ tag: 'div[data-embed-card]' }];
  },

  renderHTML({ HTMLAttributes }) {
    return ['div', mergeAttributes(HTMLAttributes, { 'data-embed-card': '', class: 'embed-card' })];
  },

  addNodeView() {
    return ReactNodeViewRenderer(
      EmbedCardNodeView as ComponentType<ReactNodeViewProps<HTMLElement>>,
    );
  },

  addCommands(): Partial<RawCommands> {
    return {
      insertEmbedCard:
        (url: string) =>
        ({ commands }: CommandProps) =>
          commands.insertContent({
            type: 'embedCard',
            attrs: { url },
          }),
    };
  },

  addStorage() {
    return {
      markdown: {
        serialize(state: MarkdownSerializerState, node: PMNode) {
          const url = (node.attrs.url as string) || '';
          state.write(`@[card](${url})`);
          state.closeBlock(node);
        },
        parse: {
          // markdown-it の inline rule で `@[card](url)` をパースし、専用 token を発行する。
          // tiptap-markdown は token → DOM 変換を介して node spec の parseHTML に流すため、
          // ここでは render を data-embed-card 属性付きの div で出す。
          setup(this: unknown, md: MarkdownItType) {
            md.inline.ruler.before('link', 'embed_card', (state: unknown, silent: boolean) => {
              const s = state as {
                src: string;
                pos: number;
                posMax: number;
                push: (type: string, tag: string, nesting: number) => MdToken;
              };
              if (silent) return false;
              const start = s.pos;
              if (s.src.charCodeAt(start) !== 0x40 /* @ */) return false;
              const rest = s.src.slice(start);
              const m = rest.match(/^@\[card\]\(([^)\s]+)\)/);
              if (!m) return false;
              const token = s.push('embed_card', '', 0);
              token.meta = { url: m[1] };
              s.pos = start + m[0].length;
              return true;
            });
            // markdown-it の renderer に embed_card token をマップ。
            (md as unknown as { renderer: { rules: Record<string, (tokens: MdToken[], idx: number) => string> } }).renderer.rules.embed_card =
              (tokens: MdToken[], idx: number) => {
                const url = tokens[idx].meta?.url ?? '';
                return `<div data-embed-card="" data-embed-url="${url}" class="embed-card"></div>`;
              };
          },
        },
      },
    };
  },
});
