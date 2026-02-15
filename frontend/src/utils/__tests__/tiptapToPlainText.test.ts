import { describe, it, expect } from 'vitest';
import { tiptapToPlainText } from '../tiptapToPlainText';

describe('tiptapToPlainText', () => {
  it('空文字列は空文字を返す', () => {
    expect(tiptapToPlainText('')).toBe('');
  });

  it('レガシーマークダウンはそのまま返す', () => {
    expect(tiptapToPlainText('普通のテキスト')).toBe('普通のテキスト');
  });

  it('段落テキストを抽出する', () => {
    const json = JSON.stringify({
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: 'こんにちは' }] },
      ],
    });
    expect(tiptapToPlainText(json)).toBe('こんにちは');
  });

  it('複数段落をスペースで結合する', () => {
    const json = JSON.stringify({
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: '行1' }] },
        { type: 'paragraph', content: [{ type: 'text', text: '行2' }] },
      ],
    });
    expect(tiptapToPlainText(json)).toBe('行1 行2');
  });

  it('見出しテキストを抽出する', () => {
    const json = JSON.stringify({
      type: 'doc',
      content: [
        { type: 'heading', attrs: { level: 1 }, content: [{ type: 'text', text: 'タイトル' }] },
      ],
    });
    expect(tiptapToPlainText(json)).toBe('タイトル');
  });

  it('リスト内テキストを抽出する', () => {
    const json = JSON.stringify({
      type: 'doc',
      content: [
        {
          type: 'bulletList',
          content: [
            { type: 'listItem', content: [{ type: 'paragraph', content: [{ type: 'text', text: 'りんご' }] }] },
            { type: 'listItem', content: [{ type: 'paragraph', content: [{ type: 'text', text: 'みかん' }] }] },
          ],
        },
      ],
    });
    expect(tiptapToPlainText(json)).toBe('りんご みかん');
  });

  it('空のdocは空文字を返す', () => {
    const json = JSON.stringify({ type: 'doc', content: [] });
    expect(tiptapToPlainText(json)).toBe('');
  });

  it('contentのないノードを処理できる', () => {
    const json = JSON.stringify({
      type: 'doc',
      content: [{ type: 'paragraph' }],
    });
    expect(tiptapToPlainText(json)).toBe('');
  });

  it('インラインノード間にスペースを挿入しない', () => {
    const json = JSON.stringify({
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [
            { type: 'text', text: 'これは' },
            { type: 'text', text: '太字', marks: [{ type: 'bold' }] },
            { type: 'text', text: 'です' },
          ],
        },
      ],
    });
    expect(tiptapToPlainText(json)).toBe('これは太字です');
  });

  it('タスクリストのテキストを抽出する', () => {
    const json = JSON.stringify({
      type: 'doc',
      content: [
        {
          type: 'taskList',
          content: [
            { type: 'taskItem', attrs: { checked: false }, content: [{ type: 'paragraph', content: [{ type: 'text', text: 'TODO1' }] }] },
            { type: 'taskItem', attrs: { checked: true }, content: [{ type: 'paragraph', content: [{ type: 'text', text: 'TODO2' }] }] },
          ],
        },
      ],
    });
    expect(tiptapToPlainText(json)).toBe('TODO1 TODO2');
  });

  it('コードブロックのテキストを抽出する', () => {
    const json = JSON.stringify({
      type: 'doc',
      content: [
        { type: 'codeBlock', attrs: { language: 'javascript' }, content: [{ type: 'text', text: 'const x = 1;' }] },
      ],
    });
    expect(tiptapToPlainText(json)).toBe('const x = 1;');
  });

  it('テーブルのテキストを抽出する', () => {
    const json = JSON.stringify({
      type: 'doc',
      content: [
        {
          type: 'table',
          content: [
            {
              type: 'tableRow',
              content: [
                { type: 'tableHeader', content: [{ type: 'paragraph', content: [{ type: 'text', text: '名前' }] }] },
                { type: 'tableHeader', content: [{ type: 'paragraph', content: [{ type: 'text', text: 'スコア' }] }] },
              ],
            },
            {
              type: 'tableRow',
              content: [
                { type: 'tableCell', content: [{ type: 'paragraph', content: [{ type: 'text', text: 'Alice' }] }] },
                { type: 'tableCell', content: [{ type: 'paragraph', content: [{ type: 'text', text: '90' }] }] },
              ],
            },
          ],
        },
      ],
    });
    expect(tiptapToPlainText(json)).toBe('名前 スコア Alice 90');
  });

  it('コールアウトのテキストを抽出する', () => {
    const json = JSON.stringify({
      type: 'doc',
      content: [
        {
          type: 'callout',
          attrs: { type: 'info', emoji: '💡' },
          content: [
            { type: 'paragraph', content: [{ type: 'text', text: '重要な情報です' }] },
          ],
        },
      ],
    });
    expect(tiptapToPlainText(json)).toBe('重要な情報です');
  });

  it('トグルリストのテキストを抽出する', () => {
    const json = JSON.stringify({
      type: 'doc',
      content: [
        {
          type: 'toggleList',
          attrs: { open: true },
          content: [
            { type: 'toggleSummary', content: [{ type: 'text', text: 'まとめ' }] },
            { type: 'toggleContent', content: [{ type: 'paragraph', content: [{ type: 'text', text: '詳細' }] }] },
          ],
        },
      ],
    });
    expect(tiptapToPlainText(json)).toBe('まとめ 詳細');
  });
});
