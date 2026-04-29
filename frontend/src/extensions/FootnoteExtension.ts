/**
 * Zenn 互換の脚注 extension。
 *
 * markdown 入力:
 *   本文[^1] / 本文^[インライン脚注]
 *   [^1]: 脚注の内容
 *
 * 設計:
 *   - markdown-it-footnote プラグインを setup() で登録し、HTML 出力を得る
 *   - footnote_ref 〜 footnote_block の HTML を Tiptap が parseHTML できるよう、
 *     updateDOM で sup.footnote-ref / section.footnotes を data 属性付きの軽量要素に置換
 *   - 編集時は inline atom (footnoteRef) と block atom (footnoteList) として扱う
 *   - serialize は ProseMirror node から markdown を再構成
 */
import { Node, mergeAttributes } from '@tiptap/core';
import type { MarkdownSerializerState } from 'prosemirror-markdown';
import type { Node as PMNode } from 'prosemirror-model';
// @ts-expect-error  no type declarations
import mdFootnote from 'markdown-it-footnote';

interface MarkdownItType {
  use: (plugin: unknown) => MarkdownItType;
}

/**
 * 脚注参照 (本文中の `[^1]` 部分)。inline atom。
 */
export const FootnoteRef = Node.create({
  name: 'footnoteRef',
  group: 'inline',
  inline: true,
  atom: true,
  selectable: true,

  addAttributes() {
    return {
      label: {
        default: '',
        parseHTML: (el) => el.getAttribute('data-footnote-ref') || '',
        renderHTML: (attrs) => ({ 'data-footnote-ref': attrs.label }),
      },
    };
  },

  parseHTML() {
    return [{ tag: 'sup[data-footnote-ref]' }];
  },

  renderHTML({ HTMLAttributes, node }) {
    return [
      'sup',
      mergeAttributes(HTMLAttributes, { class: 'footnote-ref' }),
      `[^${(node.attrs.label as string) || ''}]`,
    ];
  },

  addStorage() {
    return {
      markdown: {
        serialize(state: MarkdownSerializerState, node: PMNode) {
          const label = (node.attrs.label as string) || '';
          state.write(`[^${label}]`);
        },
        parse: {
          setup(this: unknown, md: MarkdownItType) {
            md.use(mdFootnote);
          },
          updateDOM(this: unknown, element: HTMLElement) {
            element.querySelectorAll('sup.footnote-ref a, sup .footnote-ref').forEach((el) => {
              const label = el.textContent?.replace(/[[\]]/g, '') ?? '';
              const sup = document.createElement('sup');
              sup.setAttribute('data-footnote-ref', label);
              sup.classList.add('footnote-ref');
              sup.textContent = `[^${label}]`;
              const parent = el.closest('sup');
              (parent || el).replaceWith(sup);
            });
          },
        },
      },
    };
  },
});

/**
 * 脚注定義リスト ( <section class="footnotes"> ... </section> )。block atom。
 *
 * 文書末尾に 1 つだけ存在する想定で、Tiptap node として保持する。
 * markdown serialize 時は `[^N]: 内容` 形式で書き戻す。
 */
export const FootnoteList = Node.create({
  name: 'footnoteList',
  group: 'block',
  atom: true,
  selectable: false,

  addAttributes() {
    return {
      items: {
        default: [] as { label: string; content: string }[],
        parseHTML: (el) => {
          const raw = el.getAttribute('data-footnote-items');
          if (!raw) return [];
          try {
            return JSON.parse(raw);
          } catch {
            return [];
          }
        },
        renderHTML: (attrs) => ({
          'data-footnote-items': JSON.stringify(attrs.items ?? []),
        }),
      },
    };
  },

  parseHTML() {
    return [{ tag: 'section[data-footnote-list]' }];
  },

  renderHTML({ HTMLAttributes, node }) {
    const items = (node.attrs.items as { label: string; content: string }[]) || [];
    return [
      'section',
      mergeAttributes(HTMLAttributes, {
        class: 'footnotes',
        'data-footnote-list': '',
      }),
      [
        'ol',
        {},
        ...items.map((it) => [
          'li',
          { id: `fn-${it.label}` },
          [
            'span',
            {},
            it.content,
            ` `,
            ['a', { href: `#fnref-${it.label}` }, '↩'],
          ],
        ]),
      ],
    ];
  },

  addStorage() {
    return {
      markdown: {
        serialize(state: MarkdownSerializerState, node: PMNode) {
          const items = (node.attrs.items as { label: string; content: string }[]) || [];
          if (items.length === 0) return;
          state.ensureNewLine();
          for (const it of items) {
            state.write(`[^${it.label}]: ${it.content}`);
            state.ensureNewLine();
          }
          state.closeBlock(node);
        },
        parse: {
          updateDOM(this: unknown, element: HTMLElement) {
            element.querySelectorAll('section.footnotes').forEach((sectionEl) => {
              const items: { label: string; content: string }[] = [];
              sectionEl.querySelectorAll('ol > li').forEach((li) => {
                const id = li.getAttribute('id') ?? '';
                const label = id.replace(/^fn-?/, '');
                // 末尾の "↩" 戻りリンクは内容から除外
                const clone = li.cloneNode(true) as HTMLElement;
                clone.querySelectorAll('a.footnote-backref, .footnote-backref').forEach((a) => a.remove());
                const content = clone.textContent?.trim() ?? '';
                if (label) items.push({ label, content });
              });
              const section = document.createElement('section');
              section.setAttribute('data-footnote-list', '');
              section.setAttribute('data-footnote-items', JSON.stringify(items));
              section.classList.add('footnotes');
              sectionEl.replaceWith(section);
            });
          },
        },
      },
    };
  },
});
