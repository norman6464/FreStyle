import { describe, it, expect, afterEach } from 'vitest';
import { Editor, type JSONContent } from '@tiptap/react';
import { createEditorExtensions } from '../editorExtensions';

/*
 * ページ参照（pageRef）は「題名がリンク先に追従するインラインの 1 要素」。
 * 題名の正本はサーバーが読み出し時に解決するので、ここで確かめるのは描画の契約:
 * 正しい ID はページへのリンクになり、怪しい ID はリンクにしない（atom として往復する）。
 */

let editor: Editor | null = null;

function makeEditor(content: JSONContent): Editor {
  editor = new Editor({
    element: document.createElement('div'),
    extensions: createEditorExtensions(),
    content,
  });
  return editor;
}

afterEach(() => {
  editor?.destroy();
  editor = null;
});

const uuid = '01a046a3-cf95-73f3-bd0e-1eb9b08eb1d4';

function docWithRef(pageId: unknown, title: string | null = '設計メモ'): JSONContent {
  return {
    type: 'doc',
    content: [
      { type: 'paragraph', content: [{ type: 'pageRef', attrs: { pageId, title } }] },
    ],
  };
}

describe('ページ参照（pageRef）', () => {
  it('正しい ID は /p/{id} へのリンクとして描画され、題名が文字になる', () => {
    const e = makeEditor(docWithRef(uuid));

    const html = e.getHTML();
    expect(html).toContain(`href="/p/${uuid}"`);
    expect(html).toContain('data-page-ref');
    expect(html).toContain('設計メモ');
    // 内部リンクなので _blank / rel の束（外部向けの防御）は付けない。
    expect(html).not.toContain('_blank');
    expect(e.getText()).toContain('設計メモ');
  });

  it('UUID の形でない ID はリンクにしない（押せるのにどこへも行かない要素も作らない）', () => {
    const e = makeEditor(docWithRef('javascript:alert(1)'));

    const html = e.getHTML();
    expect(html).not.toContain('href');
    expect(html).toContain('data-page-ref');
  });

  it('題名が無い参照は「ページ」と表示する（空の要素にしない）', () => {
    const e = makeEditor(docWithRef(uuid, null));

    expect(e.getHTML()).toContain('ページ');
  });

  it('doc JSON として往復する（保存 → 再読込で参照が消えない）', () => {
    const e = makeEditor(docWithRef(uuid));
    const saved = e.getJSON();

    const reloaded = makeEditor(saved);
    const html = reloaded.getHTML();
    expect(html).toContain(`href="/p/${uuid}"`);
    expect(JSON.stringify(reloaded.getJSON())).toContain(`"pageId":"${uuid}"`);
  });
});
