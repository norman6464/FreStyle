/**
 * KaTeX 数式 Tiptap extensions。
 *
 * Zenn 互換: `$$...$$` (block) / `$...$` (inline) を KaTeX で描画する。
 * 構成:
 *   - MathBlock: ProseMirror node (block)
 *   - MathInline: ProseMirror node (inline / atom)
 *
 * tiptap-markdown 連携:
 *   - serialize: ノードの latex content を `$$...$$` / `$...$` で書き戻す
 *   - parse:     markdown-it-katex プラグインを setup() で登録し、
 *                math_block / math_inline の token を Tiptap node に変換する DOM token を出力
 *
 * NodeView は katex.renderToString() で <span> / <div> に SVG/HTML 出力する。
 * KaTeX CSS は src/main.tsx などで一度だけ import する。
 */
import { Node, mergeAttributes, type RawCommands, type CommandProps } from '@tiptap/core';
import katex from 'katex';
// markdown-it-katex は型定義を持たないので @ts-expect-error で any 取り込み。
// @ts-expect-error  no type declarations
import mdKatex from 'markdown-it-katex';
import type { MarkdownSerializerState } from 'prosemirror-markdown';
import type { Node as PMNode } from 'prosemirror-model';

interface MarkdownItType {
  use: (plugin: unknown, ...params: unknown[]) => MarkdownItType;
}

function renderMath(latex: string, displayMode: boolean): string {
  try {
    return katex.renderToString(latex || '', {
      throwOnError: false,
      displayMode,
      output: 'html',
    });
  } catch {
    return `<span class="math-error">${latex}</span>`;
  }
}

/**
 * `$$...$$` のブロック数式。
 */
export const MathBlock = Node.create({
  name: 'mathBlock',
  group: 'block',
  atom: true,
  selectable: true,
  draggable: false,

  addAttributes() {
    return {
      latex: {
        default: '',
        parseHTML: (element) => element.getAttribute('data-latex') || element.textContent || '',
        renderHTML: (attributes) => ({ 'data-latex': attributes.latex }),
      },
    };
  },

  parseHTML() {
    return [
      { tag: 'div[data-math-block]' },
      { tag: 'div.math-block' },
    ];
  },

  renderHTML({ HTMLAttributes, node }) {
    const latex = (node.attrs.latex as string) || '';
    return [
      'div',
      mergeAttributes(HTMLAttributes, {
        'data-math-block': '',
        class: 'math-block',
        // KaTeX HTML output を innerHTML として持たせる必要があるが、tiptap の renderHTML は
        // children を返すかタグだけを返す形式。NodeView を使わず最低限としてはそのまま latex を残す。
        innerHTML: renderMath(latex, true),
      }),
    ];
  },

  addCommands(): Partial<RawCommands> {
    return {
      insertMathBlock:
        (latex?: string) =>
        ({ commands }: CommandProps) =>
          commands.insertContent({
            type: 'mathBlock',
            attrs: { latex: latex ?? '' },
          }),
    };
  },

  addStorage() {
    return {
      markdown: {
        serialize(state: MarkdownSerializerState, node: PMNode) {
          const latex = (node.attrs.latex as string) || '';
          state.write('$$\n');
          state.text(latex, false);
          state.ensureNewLine();
          state.write('$$');
          state.closeBlock(node);
        },
        parse: {
          setup(this: unknown, md: MarkdownItType) {
            md.use(mdKatex);
          },
          updateDOM(this: unknown, element: HTMLElement) {
            // markdown-it-katex は <span class="katex"> または <div class="katex-display"> を出すので、
            // それを mathBlock に変換するため data 属性を埋める。
            element.querySelectorAll('.katex-display').forEach((el) => {
              const annotation = el.querySelector('annotation[encoding="application/x-tex"]');
              const latex = annotation?.textContent ?? '';
              const div = document.createElement('div');
              div.setAttribute('data-math-block', '');
              div.setAttribute('class', 'math-block');
              div.setAttribute('data-latex', latex);
              el.replaceWith(div);
            });
          },
        },
      },
    };
  },
});

/**
 * `$...$` のインライン数式。
 */
export const MathInline = Node.create({
  name: 'mathInline',
  group: 'inline',
  inline: true,
  atom: true,
  selectable: true,

  addAttributes() {
    return {
      latex: {
        default: '',
        parseHTML: (element) => element.getAttribute('data-latex') || element.textContent || '',
        renderHTML: (attributes) => ({ 'data-latex': attributes.latex }),
      },
    };
  },

  parseHTML() {
    return [{ tag: 'span[data-math-inline]' }, { tag: 'span.math-inline' }];
  },

  renderHTML({ HTMLAttributes, node }) {
    const latex = (node.attrs.latex as string) || '';
    return [
      'span',
      mergeAttributes(HTMLAttributes, {
        'data-math-inline': '',
        class: 'math-inline',
        innerHTML: renderMath(latex, false),
      }),
    ];
  },

  addCommands(): Partial<RawCommands> {
    return {
      insertMathInline:
        (latex?: string) =>
        ({ commands }: CommandProps) =>
          commands.insertContent({
            type: 'mathInline',
            attrs: { latex: latex ?? '' },
          }),
    };
  },

  addStorage() {
    return {
      markdown: {
        serialize(state: MarkdownSerializerState, node: PMNode) {
          const latex = (node.attrs.latex as string) || '';
          state.write(`$${latex}$`);
        },
        parse: {
          // setup は MathBlock と同じ markdown-it-katex を共有するので no-op
          setup() {
            // intentionally empty
          },
          updateDOM(this: unknown, element: HTMLElement) {
            element.querySelectorAll('.katex:not(.katex-display .katex)').forEach((el) => {
              if (el.parentElement?.classList.contains('katex-display')) return;
              const annotation = el.querySelector('annotation[encoding="application/x-tex"]');
              const latex = annotation?.textContent ?? '';
              const span = document.createElement('span');
              span.setAttribute('data-math-inline', '');
              span.setAttribute('class', 'math-inline');
              span.setAttribute('data-latex', latex);
              el.replaceWith(span);
            });
          },
        },
      },
    };
  },
});
