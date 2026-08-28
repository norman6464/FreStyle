import { describe, it, expect } from 'vitest';
import { MAIN_NAV_ITEMS, navActive, visibleMainNav } from '../navigation';

describe('ノートへの導線', () => {
  it('主要ナビにノートがある', () => {
    // 導線が無いと、画面が出来ていても URL を手で打つしかたどり着けない。
    const notes = MAIN_NAV_ITEMS.find((item) => item.id === 'notes');

    expect(notes).toBeDefined();
    expect(notes?.to).toBe('/notes');
  });

  it('ナレッジという別項目は無い（ノートに統合した）', () => {
    const paths = MAIN_NAV_ITEMS.map((item) => item.to);

    expect(paths).toContain('/notes');
    expect(paths).not.toContain('/kb');
  });

  it('ページの中（/p/…）にいても選ばれた状態になる', () => {
    // ページの URL は /notes ではなく /p/{pageId}。系統が 2 つあるので両方で光らせる。
    const notes = MAIN_NAV_ITEMS.find((item) => item.id === 'notes');

    expect(navActive(notes!, '/notes')).toBe(true);
    expect(navActive(notes!, '/p/3ca2c0de-0000-0000-0000-000000000000')).toBe(true);
    expect(navActive(notes!, '/courses')).toBe(false);
  });

  it('文字列 1 本の matchPrefix（演習など）も子パスまで選ばれる', () => {
    // notes は配列、code は文字列。navActive は両方の形を受けるので、両方の分岐を固定する。
    const code = MAIN_NAV_ITEMS.find((item) => item.id === 'code');

    expect(navActive(code!, '/code-editor')).toBe(true);
    expect(navActive(code!, '/code-editor/123')).toBe(true);
    expect(navActive(code!, '/code-editor-x')).toBe(false);
    expect(navActive(code!, '/courses')).toBe(false);
  });

  it('名前が前方一致するだけの別パスでは選ばれない', () => {
    // 素の startsWith だと /notes-foo でも「ノート」が光る。
    // いまそういうルートは無いが、足した瞬間に静かに壊れる形なので判定側で塞ぐ。
    const notes = MAIN_NAV_ITEMS.find((item) => item.id === 'notes');

    expect(navActive(notes!, '/notes-foo')).toBe(false);
    expect(navActive(notes!, '/px')).toBe(false);
    expect(navActive(notes!, '/profile')).toBe(false);
  });

  it.each(['company_admin', 'trainee'])('%s に出す', (role) => {
    const ids = visibleMainNav(role).map((item) => item.id);

    expect(ids).toContain('notes');
  });

  it('super_admin にも出す（学習系ではなく書きものの面なので）', () => {
    // 運用の手順や決めごとを書き残すのは、むしろ管理する側の仕事になる。
    const ids = visibleMainNav('super_admin').map(
      (item) => item.id,
    );

    expect(ids).toContain('notes');
    // 学習系は引き続き出さない。
    expect(ids).not.toContain('courses');
    expect(ids).not.toContain('code');
  });
});
