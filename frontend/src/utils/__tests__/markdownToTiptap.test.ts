import { describe, it, expect } from 'vitest';
import { markdownToTiptap } from '../markdownToTiptap';

describe('markdownToTiptap', () => {
  it('空文字列で空のdocを返す', () => {
    const result = markdownToTiptap('');
    expect(result).toEqual({ type: 'doc', content: [] });
  });

  it('通常テキストをparagraphに変換する', () => {
    const result = markdownToTiptap('普通のテキスト');
    expect(result).toEqual({
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: '普通のテキスト' }] },
      ],
    });
  });

  it('見出し1を変換する', () => {
    const result = markdownToTiptap('# 見出し1');
    expect(result.content[0]).toEqual({
      type: 'heading',
      attrs: { level: 1 },
      content: [{ type: 'text', text: '見出し1' }],
    });
  });

  it('見出し2を変換する', () => {
    const result = markdownToTiptap('## 見出し2');
    expect(result.content[0]).toEqual({
      type: 'heading',
      attrs: { level: 2 },
      content: [{ type: 'text', text: '見出し2' }],
    });
  });

  it('見出し3を変換する', () => {
    const result = markdownToTiptap('### 見出し3');
    expect(result.content[0]).toEqual({
      type: 'heading',
      attrs: { level: 3 },
      content: [{ type: 'text', text: '見出し3' }],
    });
  });

  it('箇条書き（-）を変換する', () => {
    const result = markdownToTiptap('- りんご\n- みかん');
    expect(result.content[0].type).toBe('bulletList');
    expect(result.content[0].content).toHaveLength(2);
    expect(result.content[0].content![0].type).toBe('listItem');
  });

  it('箇条書き（・）を変換する', () => {
    const result = markdownToTiptap('・りんご\n・みかん');
    expect(result.content[0].type).toBe('bulletList');
    expect(result.content[0].content).toHaveLength(2);
  });

  it('番号付きリストを変換する', () => {
    const result = markdownToTiptap('1. 手順1\n2. 手順2');
    expect(result.content[0].type).toBe('orderedList');
    expect(result.content[0].content).toHaveLength(2);
  });

  it('太字を変換する', () => {
    const result = markdownToTiptap('これは**太字**です');
    const paragraph = result.content[0];
    expect(paragraph.content).toBeDefined();
    const boldNode = paragraph.content!.find(
      (n: { marks?: { type: string }[] }) => n.marks?.some(m => m.type === 'bold')
    );
    expect(boldNode).toBeDefined();
    expect(boldNode!.text).toBe('太字');
  });

  it('斜体を変換する', () => {
    const result = markdownToTiptap('これは*斜体*です');
    const paragraph = result.content[0];
    const italicNode = paragraph.content!.find(
      (n: { marks?: { type: string }[] }) => n.marks?.some(m => m.type === 'italic')
    );
    expect(italicNode).toBeDefined();
    expect(italicNode!.text).toBe('斜体');
  });

  it('複数ブロックの混合を変換する', () => {
    const md = '# タイトル\n\n普通のテキスト\n\n- 項目1\n- 項目2';
    const result = markdownToTiptap(md);
    expect(result.content.length).toBeGreaterThanOrEqual(3);
    expect(result.content[0].type).toBe('heading');
    expect(result.content.some((b: { type: string }) => b.type === 'paragraph')).toBe(true);
    expect(result.content.some((b: { type: string }) => b.type === 'bulletList')).toBe(true);
  });

  it('CRLF改行を正しく処理する', () => {
    const result = markdownToTiptap('- 項目1\r\n- 項目2\r\n- 項目3');
    expect(result.content[0].type).toBe('bulletList');
    expect(result.content[0].content).toHaveLength(3);
  });

  it('空白のみの文字列で空のdocを返す', () => {
    const result = markdownToTiptap('   \n  \n   ');
    expect(result).toEqual({ type: 'doc', content: [] });
  });

  it('リスト項目内の太字を変換する', () => {
    const result = markdownToTiptap('- **重要な**項目');
    const listItem = result.content[0].content![0];
    const paragraph = listItem.content![0];
    const boldNode = paragraph.content!.find(
      (n: { marks?: { type: string }[] }) => n.marks?.some(m => m.type === 'bold')
    );
    expect(boldNode).toBeDefined();
    expect(boldNode!.text).toBe('重要な');
  });

  it('CR改行を正しく処理する', () => {
    const result = markdownToTiptap('行1\r行2');
    expect(result.content).toHaveLength(2);
    expect(result.content[0].type).toBe('paragraph');
    expect(result.content[1].type).toBe('paragraph');
  });
});
