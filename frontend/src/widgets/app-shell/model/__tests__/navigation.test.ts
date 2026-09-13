import { describe, it, expect } from 'vitest';
import { MAIN_NAV_ITEMS, isKbPath, navActive } from '../navigation';

describe('MAIN_NAV_ITEMS', () => {
  it('素のリンクで表せるホームだけを持つ（スペース・最近見たページは専用コンポーネント）', () => {
    const ids = MAIN_NAV_ITEMS.map((item) => item.id);
    expect(ids).toEqual(['home']);
  });

  it('/notes という項目・パスはもう無い（撤去済み）', () => {
    const ids = MAIN_NAV_ITEMS.map((item) => item.id);
    const paths = MAIN_NAV_ITEMS.map((item) => item.to);

    expect(ids).not.toContain('notes');
    expect(paths).not.toContain('/notes');
  });

  it('ホームは / の完全一致でだけ選ばれる', () => {
    const home = MAIN_NAV_ITEMS.find((item) => item.id === 'home')!;
    expect(navActive(home, '/')).toBe(true);
    expect(navActive(home, '/kb')).toBe(false);
  });
});

describe('isKbPath（ヘッダーの「スペース ▾」を光らせる経路）', () => {
  it('/kb 配下（ページ・スペース・バックログ・チケット）はすべて含む', () => {
    expect(isKbPath('/kb')).toBe(true);
    expect(isKbPath('/kb/3ca2c0de-0000-0000-0000-000000000000')).toBe(true);
    expect(isKbPath('/kb/spaces/s-1')).toBe(true);
    expect(isKbPath('/kb/backlog/s-1')).toBe(true);
    expect(isKbPath('/kb/tickets/t-1')).toBe(true);
  });

  it('/kb 以外は含まない', () => {
    expect(isKbPath('/courses')).toBe(false);
    expect(isKbPath('/')).toBe(false);
  });

  it('名前が前方一致するだけの別パスでは選ばれない', () => {
    // 素の startsWith だと /kb-other でも選ばれてしまう。いまそういうルートは
    // 無いが、足した瞬間に静かに壊れる形なので判定側で塞ぐ。
    expect(isKbPath('/kb-other')).toBe(false);
  });
});
