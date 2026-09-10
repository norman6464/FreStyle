import { describe, it, expect } from 'vitest';
import { editorContentToSegments, isEditorContentEmpty, segmentsToEditorContent } from '../mentionComposerContent';
import type { TicketCommentSegment } from '@/entities/ticket';

describe('editorContentToSegments', () => {
  it('text ノードをそのまま区間へ写す', () => {
    const doc = { type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'text', text: 'こんにちは' }] }] };
    expect(editorContentToSegments(doc)).toEqual([{ kind: 'text', text: 'こんにちは' }]);
  });

  it('隣り合う text 区間を 1 つへ結合する', () => {
    const doc = {
      type: 'doc',
      content: [{ type: 'paragraph', content: [{ type: 'text', text: 'あ' }, { type: 'text', text: 'い' }] }],
    };
    expect(editorContentToSegments(doc)).toEqual([{ kind: 'text', text: 'あい' }]);
  });

  it('mention ノードは userId を持つ区間になる（表示名は送らない）', () => {
    const doc = {
      type: 'doc',
      content: [{ type: 'paragraph', content: [{ type: 'mention', attrs: { userId: '42', name: '田中 太郎' } }] }],
    };
    expect(editorContentToSegments(doc)).toEqual([{ kind: 'mention', userId: '42' }]);
  });

  it('hardBreak は文字の \\n として前後の text と結合する（wire に hardBreak ノードを残さない）', () => {
    const doc = {
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [{ type: 'text', text: '1行目' }, { type: 'hardBreak' }, { type: 'text', text: '2行目' }],
        },
      ],
    };
    expect(editorContentToSegments(doc)).toEqual([{ kind: 'text', text: '1行目\n2行目' }]);
  });

  it('text と mention が混ざっても順番どおり区間になる', () => {
    const doc = {
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [
            { type: 'text', text: 'よろしく ' },
            { type: 'mention', attrs: { userId: '7', name: '佐藤' } },
            { type: 'text', text: ' お願いします' },
          ],
        },
      ],
    };
    expect(editorContentToSegments(doc)).toEqual([
      { kind: 'text', text: 'よろしく ' },
      { kind: 'mention', userId: '7' },
      { kind: 'text', text: ' お願いします' },
    ]);
  });

  it('複数段落は \\n で畳む（貼り付けで紛れ込んだ場合の保険）', () => {
    const doc = {
      type: 'doc',
      content: [
        { type: 'paragraph', content: [{ type: 'text', text: '段落1' }] },
        { type: 'paragraph', content: [{ type: 'text', text: '段落2' }] },
      ],
    };
    expect(editorContentToSegments(doc)).toEqual([{ kind: 'text', text: '段落1\n段落2' }]);
  });

  it('中身が無ければ空配列', () => {
    const doc = { type: 'doc', content: [{ type: 'paragraph', content: [] }] };
    expect(editorContentToSegments(doc)).toEqual([]);
  });

  it('userId が空文字の mention は落とす', () => {
    const doc = { type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'mention', attrs: { userId: '' } }] }] };
    expect(editorContentToSegments(doc)).toEqual([]);
  });
});

describe('segmentsToEditorContent / editorContentToSegments の往復', () => {
  it('text だけの区間は往復して同じ内容に戻る', () => {
    const segments: TicketCommentSegment[] = [{ kind: 'text', text: '1行目\n2行目' }];
    const content = segmentsToEditorContent(segments, () => null);
    expect(editorContentToSegments(content)).toEqual(segments);
  });

  it('mention を含む区間も往復する（表示名は解決関数を通す）', () => {
    const segments: TicketCommentSegment[] = [
      { kind: 'text', text: 'よろしく ' },
      { kind: 'mention', userId: '7' },
    ];
    const content = segmentsToEditorContent(segments, (userId) => (userId === '7' ? '佐藤 花子' : null));
    expect(editorContentToSegments(content)).toEqual(segments);
  });

  it('名前を引けない mention は「不明なユーザー」で埋める', () => {
    const content = segmentsToEditorContent([{ kind: 'mention', userId: '99' }], () => null);
    expect(content).toEqual({
      type: 'doc',
      content: [{ type: 'paragraph', content: [{ type: 'mention', attrs: { userId: '99', name: '不明なユーザー' } }] }],
    });
  });
});

describe('isEditorContentEmpty', () => {
  it('空の段落は空扱い', () => {
    expect(isEditorContentEmpty({ type: 'doc', content: [{ type: 'paragraph', content: [] }] })).toBe(true);
  });

  it('空白だけの text は空扱い', () => {
    expect(
      isEditorContentEmpty({ type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'text', text: '   ' }] }] }),
    ).toBe(true);
  });

  it('mention だけでも空ではない', () => {
    expect(
      isEditorContentEmpty({
        type: 'doc',
        content: [{ type: 'paragraph', content: [{ type: 'mention', attrs: { userId: '1', name: 'x' } }] }],
      }),
    ).toBe(false);
  });

  it('文字があれば空ではない', () => {
    expect(
      isEditorContentEmpty({ type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'text', text: 'a' }] }] }),
    ).toBe(false);
  });
});
