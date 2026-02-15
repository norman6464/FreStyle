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
});
