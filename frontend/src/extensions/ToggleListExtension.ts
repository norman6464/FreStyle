import { Node, mergeAttributes, type RawCommands, type CommandProps } from '@tiptap/core';
import { ReactNodeViewRenderer, type ReactNodeViewProps } from '@tiptap/react';
import type { ComponentType } from 'react';
import container from 'markdown-it-container';

// markdown-it 14.x の namespace アクセスを避けるため、必要なフィールドだけ local type で定義。
interface MdToken {
  nesting: 1 | 0 | -1;
  info: string;
  type: string;
}

interface MarkdownItType {
  use: (plugin: unknown, ...params: unknown[]) => MarkdownItType;
}
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

  addStorage() {
    return {
      markdown: {
        // toggleSummary の中身は Zenn では `:::details タイトル` の "タイトル" 部に直接書くため、
        // markdown-serialize としては個別に何も書かない（toggleList 側で attrs.title として吸収する）。
        serialize() {
          // no-op: serialize is handled by parent toggleList
        },
      },
    };
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

  addStorage() {
    return {
      markdown: {
        serialize(
          state: import('prosemirror-markdown').MarkdownSerializerState,
          node: import('prosemirror-model').Node,
        ) {
          state.renderContent(node);
        },
      },
    };
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

  /**
   * Zenn `:::details タイトル` 記法 ↔ Tiptap toggleList の相互変換。
   *
   * - serialize:
   *     :::details {summary のテキスト}
   *     {toggleContent の中身を markdown 化}
   *     :::
   * - parse:
   *     markdown-it-container で `details` をパースし、開きタグの info 部分を
   *     summary テキストとして使う。
   */
  addStorage() {
    return {
      markdown: {
        serialize(
          state: import('prosemirror-markdown').MarkdownSerializerState,
          node: import('prosemirror-model').Node,
        ) {
          let summaryText = '';
          let contentNode: import('prosemirror-model').Node | null = null;
          node.forEach((child) => {
            if (child.type.name === 'toggleSummary') summaryText = child.textContent;
            else if (child.type.name === 'toggleContent') contentNode = child;
          });
          state.write(`:::details ${summaryText}\n`);
          if (contentNode) state.renderContent(contentNode);
          state.ensureNewLine();
          state.write(':::');
          state.closeBlock(node);
        },
        parse: {
          setup(_md: MarkdownItType) {
            // details container は Callout 側と独立に登録する。
            // markdown-it は同じ名前の container を二重登録すると例外を出すので、
            // ここでは details というユニーク名を使う。
            _md.use(container, 'details', {
              validate: (params: string) => /^details\s+(.*)$/.test(params.trim()),
              render: (tokens: MdToken[], idx: number) => {
                const token = tokens[idx];
                if (token.nesting === 1) {
                  const m = token.info.trim().match(/^details\s+(.*)$/);
                  const title = m ? m[1] : '';
                  return `<details class="toggle-list" open><summary>${title}</summary><div data-toggle-content="">\n`;
                }
                return '</div></details>\n';
              },
            });
          },
        },
      },
    };
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
