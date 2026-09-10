import { describe, it, expect } from 'vitest';
import { readCommentBody, buildCommentBody } from '../commentBody';

describe('readCommentBody', () => {
  it('text と mention を区間の列に畳む', () => {
    const body = [
      { type: 'text', text: 'お疲れさまです ' },
      { type: 'mention', attrs: { userId: '42' } },
      { type: 'text', text: ' さん' },
    ];
    expect(readCommentBody(body)).toEqual([
      { kind: 'text', text: 'お疲れさまです ' },
      { kind: 'mention', userId: '42' },
      { kind: 'text', text: ' さん' },
    ]);
  });

  it('mention の userId が数値でも文字列化する', () => {
    expect(readCommentBody([{ type: 'mention', attrs: { userId: 42 } }])).toEqual([
      { kind: 'mention', userId: '42' },
    ]);
  });

  it('userId が無い・空文字の mention は落とす', () => {
    expect(readCommentBody([{ type: 'mention', attrs: {} }])).toEqual([]);
    expect(readCommentBody([{ type: 'mention', attrs: { userId: '' } }])).toEqual([]);
  });

  it('未知の type は文字を持っていれば拾い、持っていなければ捨てる', () => {
    expect(readCommentBody([{ type: 'bold', text: '強調' }])).toEqual([{ kind: 'text', text: '強調' }]);
    expect(readCommentBody([{ type: 'hardBreak' }])).toEqual([]);
  });

  it('配列でない・要素が record でない本文は空として扱う', () => {
    expect(readCommentBody(null)).toEqual([]);
    expect(readCommentBody('text')).toEqual([]);
    expect(readCommentBody([null, 'x', 42])).toEqual([]);
  });

  it('空文字の text ノードは落とす', () => {
    expect(readCommentBody([{ type: 'text', text: '' }])).toEqual([]);
  });
});

describe('buildCommentBody', () => {
  it('区間の列を送信できるノード配列に組み立てる', () => {
    expect(
      buildCommentBody([
        { kind: 'text', text: 'こんにちは ' },
        { kind: 'mention', userId: '42' },
      ]),
    ).toEqual([
      { type: 'text', text: 'こんにちは ' },
      { type: 'mention', attrs: { userId: '42' } },
    ]);
  });

  it('空白だけの text は落とす（名指しの区切り空白がこれに当たる）', () => {
    expect(
      buildCommentBody([
        { kind: 'mention', userId: '1' },
        { kind: 'text', text: ' ' },
        { kind: 'mention', userId: '2' },
      ]),
    ).toEqual([
      { type: 'mention', attrs: { userId: '1' } },
      { type: 'mention', attrs: { userId: '2' } },
    ]);
  });

  it('userId が空文字の mention は落とす', () => {
    expect(buildCommentBody([{ kind: 'mention', userId: '' }])).toEqual([]);
  });

  it('全部落ちたら空配列を返す（呼び出し側が送信可否を見る）', () => {
    expect(buildCommentBody([{ kind: 'text', text: '   ' }])).toEqual([]);
  });
});
