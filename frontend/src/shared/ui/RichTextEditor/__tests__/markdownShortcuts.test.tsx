import { describe, it, expect, afterEach } from 'vitest';
import { Editor, type JSONContent } from '@tiptap/react';
import { createEditorExtensions } from '../editorExtensions';
import { emptyRichDoc } from '../emptyRichDoc';

let editor: Editor | null = null;

function makeEditor(content: JSONContent = emptyRichDoc()): Editor {
  editor = new Editor({
    element: document.createElement('div'),
    extensions: createEditorExtensions(),
    content,
  });
  return editor;
}

/**
 * typeConfirmed は IME の変換確定と同じ経路（handleTextInput を経ない doc 変化）で
 * テキストを入力する。input rule は発火せず、appendTransaction（MarkdownShortcuts）だけが働く。
 */
function typeConfirmed(e: Editor, text: string) {
  e.chain().focus('end').insertContent(text).run();
}

const firstNode = (e: Editor) => e.getJSON().content?.[0];

afterEach(() => {
  editor?.destroy();
  editor = null;
});

describe('MarkdownShortcuts（IME 確定経路でも効く変換）', () => {
  it.each([
    ['＃ ', 1],
    ['＃＃ ', 2],
    ['＃＃＃ ', 3],
    ['# ', 1],
    ['＃# ', 2],
  ])('確定入力 %s で見出し%i になる', (text, level) => {
    const e = makeEditor();
    typeConfirmed(e, text);
    expect(firstNode(e)?.type).toBe('heading');
    expect(firstNode(e)?.attrs?.level).toBe(level);
  });

  it('全角スペース確定（＃＃　）でも見出し2になる', () => {
    const e = makeEditor();
    typeConfirmed(e, '＃＃　');
    expect(firstNode(e)?.type).toBe('heading');
    expect(firstNode(e)?.attrs?.level).toBe(2);
  });

  it('``` でコードブロックになる（言語なし）', () => {
    const e = makeEditor();
    typeConfirmed(e, '``` ');
    expect(firstNode(e)?.type).toBe('codeBlock');
    expect(firstNode(e)?.attrs?.language ?? null).toBeNull();
  });

  it('```sql で言語付きコードブロックになる', () => {
    const e = makeEditor();
    typeConfirmed(e, '```sql ');
    expect(firstNode(e)?.type).toBe('codeBlock');
    expect(firstNode(e)?.attrs?.language).toBe('sql');
  });

  it('全角 ｀｀｀ でもコードブロックになる', () => {
    const e = makeEditor();
    typeConfirmed(e, '｀｀｀ ');
    expect(firstNode(e)?.type).toBe('codeBlock');
  });

  it('本文が続く行では変換しない（# の後に文章）', () => {
    const e = makeEditor();
    typeConfirmed(e, '# 見出しではない本文');
    expect(firstNode(e)?.type).toBe('paragraph');
  });

  it('スペースで確定しなければ変換しない', () => {
    const e = makeEditor();
    typeConfirmed(e, '＃＃');
    expect(firstNode(e)?.type).toBe('paragraph');
  });

  it('既存の見出し・コードブロック内では暴発しない', () => {
    const e = makeEditor({
      type: 'doc',
      content: [{ type: 'codeBlock', attrs: { language: null }, content: [{ type: 'text', text: '```' }] }],
    });
    typeConfirmed(e, ' ');
    // codeBlock の中はそのまま（段落だけが変換対象）。
    expect(firstNode(e)?.type).toBe('codeBlock');
    expect(firstNode(e)?.content?.[0]?.text).toBe('``` ');
  });
});
