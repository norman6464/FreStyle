import { describe, it, expect } from 'vitest';
import { tiptapToMarkdown } from '../tiptapToMarkdown';

describe('tiptapToMarkdown', () => {
  it('空のdocを空文字列に変換する', () => {
    const doc = { type: 'doc', content: [] };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('');
  });

  it('空文字列を空文字列に変換する', () => {
    expect(tiptapToMarkdown('')).toBe('');
  });

  it('パラグラフをテキストに変換する', () => {
    const doc = {
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: 'こんにちは' }] },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('こんにちは');
  });

  it('複数のパラグラフを改行で区切る', () => {
    const doc = {
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: '段落1' }] },
        { type: 'paragraph', content: [{ type: 'text', text: '段落2' }] },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('段落1\n\n段落2');
  });

  it('見出しをMarkdownに変換する', () => {
    const doc = {
      type: 'doc',
      content: [
        { type: 'heading', attrs: { level: 1 }, content: [{ type: 'text', text: '見出し1' }] },
        { type: 'heading', attrs: { level: 2 }, content: [{ type: 'text', text: '見出し2' }] },
        { type: 'heading', attrs: { level: 3 }, content: [{ type: 'text', text: '見出し3' }] },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('# 見出し1\n\n## 見出し2\n\n### 見出し3');
  });

  it('太字をMarkdownに変換する', () => {
    const doc = {
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [
            { type: 'text', text: '通常テキスト' },
            { type: 'text', text: '太字', marks: [{ type: 'bold' }] },
          ],
        },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('通常テキスト**太字**');
  });

  it('イタリックをMarkdownに変換する', () => {
    const doc = {
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [
            { type: 'text', text: 'テスト', marks: [{ type: 'italic' }] },
          ],
        },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('*テスト*');
  });

  it('箇条書きリストをMarkdownに変換する', () => {
    const doc = {
      type: 'doc',
      content: [
        {
          type: 'bulletList',
          content: [
            { type: 'listItem', content: [{ type: 'paragraph', content: [{ type: 'text', text: '項目1' }] }] },
            { type: 'listItem', content: [{ type: 'paragraph', content: [{ type: 'text', text: '項目2' }] }] },
          ],
        },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('- 項目1\n- 項目2');
  });

  it('番号付きリストをMarkdownに変換する', () => {
    const doc = {
      type: 'doc',
      content: [
        {
          type: 'orderedList',
          content: [
            { type: 'listItem', content: [{ type: 'paragraph', content: [{ type: 'text', text: '手順1' }] }] },
            { type: 'listItem', content: [{ type: 'paragraph', content: [{ type: 'text', text: '手順2' }] }] },
          ],
        },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('1. 手順1\n2. 手順2');
  });

  it('ブロック引用をMarkdownに変換する', () => {
    const doc = {
      type: 'doc',
      content: [
        {
          type: 'blockquote',
          content: [
            { type: 'paragraph', content: [{ type: 'text', text: '引用テキスト' }] },
          ],
        },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('> 引用テキスト');
  });

  it('コードブロックをMarkdownに変換する', () => {
    const doc = {
      type: 'doc',
      content: [
        {
          type: 'codeBlock',
          attrs: { language: 'javascript' },
          content: [{ type: 'text', text: 'const x = 1;' }],
        },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('```javascript\nconst x = 1;\n```');
  });

  it('水平線をMarkdownに変換する', () => {
    const doc = {
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: '上' }] },
        { type: 'horizontalRule' },
        { type: 'paragraph', content: [{ type: 'text', text: '下' }] },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('上\n\n---\n\n下');
  });

  it('画像をMarkdownに変換する', () => {
    const doc = {
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [
            { type: 'image', attrs: { src: 'https://example.com/img.png', alt: '画像', title: null } },
          ],
        },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('![画像](https://example.com/img.png)');
  });

  it('タスクリストをMarkdownに変換する', () => {
    const doc = {
      type: 'doc',
      content: [
        {
          type: 'taskList',
          content: [
            { type: 'taskItem', attrs: { checked: true }, content: [{ type: 'paragraph', content: [{ type: 'text', text: '完了タスク' }] }] },
            { type: 'taskItem', attrs: { checked: false }, content: [{ type: 'paragraph', content: [{ type: 'text', text: '未完了タスク' }] }] },
          ],
        },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('- [x] 完了タスク\n- [ ] 未完了タスク');
  });

  it('リンクをMarkdownに変換する', () => {
    const doc = {
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [
            { type: 'text', text: 'リンク', marks: [{ type: 'link', attrs: { href: 'https://example.com' } }] },
          ],
        },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('[リンク](https://example.com)');
  });

  it('インラインコードをMarkdownに変換する', () => {
    const doc = {
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [
            { type: 'text', text: 'コード', marks: [{ type: 'code' }] },
          ],
        },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('`コード`');
  });

  it('不正なJSONはそのまま返す', () => {
    expect(tiptapToMarkdown('invalid json')).toBe('invalid json');
  });

  it('空内容のパラグラフは空行として扱う', () => {
    const doc = {
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: '上' }] },
        { type: 'paragraph' },
        { type: 'paragraph', content: [{ type: 'text', text: '下' }] },
      ],
    };
    expect(tiptapToMarkdown(JSON.stringify(doc))).toBe('上\n\n\n\n下');
  });
});
