/**
 * Zenn 仕様の URL 単独行 → サービス専用埋め込み変換。
 *
 * 対応:
 *   - YouTube      ( youtu.be / youtube.com/watch?v= )
 *   - Twitter (X)  ( twitter.com|x.com/<user>/status/<id> )
 *   - GitHub permalink ( github.com/.../blob/.../path[#L1-L3] )
 *
 * markdown-it の block ruler で「行内に URL 1 つだけ含む段落」を検出し、
 * service_embed トークンに置換する。renderer は data-service-embed 付きの
 * div を出力し、Tiptap が parseHTML で serviceEmbed node に取り込む。
 *
 * 本 PR 段階では NodeView は最小描画 (PR G の ServiceEmbedNodeView)。
 */
import { Node, mergeAttributes, type RawCommands, type CommandProps } from '@tiptap/core';
import { ReactNodeViewRenderer, type ReactNodeViewProps } from '@tiptap/react';
import type { ComponentType } from 'react';
import type { MarkdownSerializerState } from 'prosemirror-markdown';
import type { Node as PMNode } from 'prosemirror-model';
import ServiceEmbedNodeView from '../components/ServiceEmbedNodeView';

interface MdState {
  src: string;
  bMarks: number[];
  eMarks: number[];
  tShift: number[];
  push: (type: string, tag: string, nesting: number) => MdToken;
}

interface MdToken {
  type: string;
  meta?: { kind?: string; url?: string };
  block?: boolean;
  map?: [number, number];
}

interface MarkdownItType {
  block: { ruler: { before: (existing: string, name: string, fn: unknown, opts?: unknown) => void } };
  renderer: { rules: Record<string, (tokens: MdToken[], idx: number) => string> };
}

function classifyServiceUrl(url: string): 'youtube' | 'twitter' | 'github' | null {
  try {
    const u = new URL(url);
    const host = u.hostname;
    if (host === 'youtu.be' || host.endsWith('youtube.com')) return 'youtube';
    if (host === 'twitter.com' || host === 'x.com') return 'twitter';
    if (host === 'github.com') return 'github';
    return null;
  } catch {
    return null;
  }
}

export const ServiceEmbed = Node.create({
  name: 'serviceEmbed',
  group: 'block',
  atom: true,
  selectable: true,

  addAttributes() {
    return {
      kind: {
        default: 'youtube',
        parseHTML: (el) => el.getAttribute('data-service-kind') || 'youtube',
        renderHTML: (attrs) => ({ 'data-service-kind': attrs.kind }),
      },
      url: {
        default: '',
        parseHTML: (el) => el.getAttribute('data-service-url') || '',
        renderHTML: (attrs) => ({ 'data-service-url': attrs.url }),
      },
    };
  },

  parseHTML() {
    return [{ tag: 'div[data-service-embed]' }];
  },

  renderHTML({ HTMLAttributes }) {
    return [
      'div',
      mergeAttributes(HTMLAttributes, {
        'data-service-embed': '',
        class: 'service-embed',
      }),
    ];
  },

  addNodeView() {
    return ReactNodeViewRenderer(
      ServiceEmbedNodeView as ComponentType<ReactNodeViewProps<HTMLElement>>,
    );
  },

  addCommands(): Partial<RawCommands> {
    return {
      insertServiceEmbed:
        (kind: 'youtube' | 'twitter' | 'github', url: string) =>
        ({ commands }: CommandProps) =>
          commands.insertContent({
            type: 'serviceEmbed',
            attrs: { kind, url },
          }),
    };
  },

  addStorage() {
    return {
      markdown: {
        serialize(state: MarkdownSerializerState, node: PMNode) {
          const url = (node.attrs.url as string) || '';
          // Zenn の表記に合わせて URL 単独行で書き戻す。
          state.write(url);
          state.closeBlock(node);
        },
        parse: {
          setup(this: unknown, md: MarkdownItType) {
            // block ruler に "URL only line" 検出を入れる。paragraph 検出より前に走らせる。
            md.block.ruler.before(
              'paragraph',
              'service_embed',
              (state: MdState, startLine: number, endLine: number, silent: boolean) => {
                const start = state.bMarks[startLine] + state.tShift[startLine];
                const max = state.eMarks[startLine];
                const line = state.src.slice(start, max).trim();
                // URL 単独行 (前後空白除く)
                if (!/^https?:\/\/\S+$/.test(line)) return false;
                const kind = classifyServiceUrl(line);
                if (!kind) return false;
                if (silent) return true;
                const token = state.push('service_embed', '', 0);
                token.block = true;
                token.meta = { kind, url: line };
                token.map = [startLine, startLine + 1];
                // 1 行進める
                return true;
              },
            );
            md.renderer.rules.service_embed = (tokens: MdToken[], idx: number) => {
              const t = tokens[idx];
              const kind = t.meta?.kind ?? 'youtube';
              const url = t.meta?.url ?? '';
              return `<div data-service-embed="" data-service-kind="${kind}" data-service-url="${url}" class="service-embed"></div>`;
            };
          },
        },
      },
    };
  },
});
