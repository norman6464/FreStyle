import { describe, it, expect } from 'vitest';
import { toArray } from '../toArray';

/**
 * API が 0 件のとき null を返しても画面が落ちないようにする防御（FRESTYLE-77）。
 * ここが崩れると map / filter / for-of が TypeError になり、
 * データがまだ無い新規ユーザーがそのページを開けなくなる。
 */
describe('toArray', () => {
  it('配列はそのまま返す', () => {
    const rows = [{ id: 1 }, { id: 2 }];
    expect(toArray(rows)).toBe(rows);
  });

  it('空配列はそのまま返す', () => {
    expect(toArray([])).toEqual([]);
  });

  it.each([
    ['null', null],
    ['undefined', undefined],
  ])('%s は空配列にする', (_label, value) => {
    expect(toArray(value)).toEqual([]);
  });

  // 想定外の形（オブジェクト・文字列・数値）でも配列として扱えるようにする。
  // 文字列は spread すると 1 文字ずつの配列になってしまうため、空配列に倒す。
  it.each([
    ['オブジェクト', { items: [] }],
    ['文字列', 'abc'],
    ['数値', 0],
    ['真偽値', false],
  ])('配列でない %s は空配列にする', (_label, value) => {
    expect(toArray(value)).toEqual([]);
  });

  it('結果は必ず反復できる', () => {
    for (const _ of toArray(null)) {
      throw new Error('空配列なので回らないはず');
    }
    expect(toArray<number>(null).map((n) => n * 2)).toEqual([]);
  });
});
