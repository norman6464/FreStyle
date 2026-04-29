/**
 * Zenn 互換 Mermaid ダイアグラム拡張。
 *
 * 設計:
 * - markdown 入力では ```mermaid ... ``` のフェンスコードブロックとして書かれる
 * - parse: tiptap-markdown の updateDOM で <pre><code class="language-mermaid"> を検出し、
 *   <div data-mermaid-block data-code="..."> に置き換えて Tiptap が atom node として読み込む
 * - serialize: Tiptap mermaid node → ```mermaid\n{code}\n``` で書き戻し
 * - NodeView は MermaidNodeView (dynamic import で mermaid をロード)
 */
import { Node, mergeAttributes, type RawCommands, type CommandProps } from '@tiptap/core';
import { ReactNodeViewRenderer, type ReactNodeViewProps } from '@tiptap/react';
import type { ComponentType } from 'react';
import type { MarkdownSerializerState } from 'prosemirror-markdown';
import type { Node as PMNode } from 'prosemirror-model';
import MermaidNodeView from '../components/MermaidNodeView';

export const Mermaid = Node.create({
  name: 'mermaid',
  group: 'block',
  atom: true,
  selectable: true,

  addAttributes() {
    return {
      code: {
        default: '',
        parseHTML: (element) => element.getAttribute('data-code') || element.textContent || '',
        renderHTML: (attributes) => ({ 'data-code': attributes.code }),
      },
    };
  },

  parseHTML() {
    return [{ tag: 'div[data-mermaid-block]' }];
  },

  renderHTML({ HTMLAttributes }) {
    return ['div', mergeAttributes(HTMLAttributes, { 'data-mermaid-block': '', class: 'mermaid-block' })];
  },

  addNodeView() {
    return ReactNodeViewRenderer(
      MermaidNodeView as ComponentType<ReactNodeViewProps<HTMLElement>>,
    );
  },

  addCommands(): Partial<RawCommands> {
    return {
      insertMermaid:
        (code?: string) =>
        ({ commands }: CommandProps) =>
          commands.insertContent({
            type: 'mermaid',
            attrs: { code: code ?? '' },
          }),
    };
  },

  addStorage() {
    return {
      markdown: {
        serialize(state: MarkdownSerializerState, node: PMNode) {
          const code = (node.attrs.code as string) || '';
          state.write('```mermaid\n');
          state.text(code, false);
          state.ensureNewLine();
          state.write('```');
          state.closeBlock(node);
        },
        parse: {
          // markdown-it の標準 fence パーサーが ```mermaid``` を出した HTML を後処理で
          // mermaid block に置換する。tiptap-markdown が DOM 経由でパースする仕組みのため、
          // updateDOM で <pre><code class="language-mermaid"> を <div data-mermaid-block> に変換する。
          updateDOM(this: unknown, element: HTMLElement) {
            element.querySelectorAll('pre > code.language-mermaid').forEach((codeEl) => {
              const pre = codeEl.parentElement;
              if (!pre) return;
              const div = document.createElement('div');
              div.setAttribute('data-mermaid-block', '');
              div.setAttribute('data-code', codeEl.textContent ?? '');
              div.classList.add('mermaid-block');
              pre.replaceWith(div);
            });
          },
        },
      },
    };
  },
});
