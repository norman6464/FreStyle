import { describe, it, expect } from 'vitest';
import { MAIN_NAV_ITEMS, navActive, visibleMainNav } from '../navigation';

describe('ナレッジへの導線', () => {
  it('主要ナビにナレッジがある', () => {
    // 導線が無いと、画面が出来ていても URL を手で打つしかたどり着けない。
    const kb = MAIN_NAV_ITEMS.find((item) => item.id === 'kb');

    expect(kb).toBeDefined();
    expect(kb?.to).toBe('/kb');
  });

  it('ノートとは別の項目にする', () => {
    // ノートは自分のための平らな一覧、ナレッジは共有される木。系統が違う。
    const paths = MAIN_NAV_ITEMS.map((item) => item.to);

    expect(paths).toContain('/notes');
    expect(paths).toContain('/kb');
  });

  it('ページの中にいても選ばれた状態になる', () => {
    const kb = MAIN_NAV_ITEMS.find((item) => item.id === 'kb');

    expect(navActive(kb!, '/kb')).toBe(true);
    expect(navActive(kb!, '/kb/acme/pages/abc')).toBe(true);
    expect(navActive(kb!, '/notes')).toBe(false);
  });

  it.each(['company_admin', 'trainee'])('%s に出す', (role) => {
    const ids = visibleMainNav(role, { aiChatEnabledForTrainees: true }).map((item) => item.id);

    expect(ids).toContain('kb');
  });

  it('super_admin にも出す（学習系ではなく書きものの面なので）', () => {
    // 運用の手順や決めごとを書き残すのは、むしろ管理する側の仕事になる。
    const ids = visibleMainNav('super_admin', { aiChatEnabledForTrainees: true }).map(
      (item) => item.id,
    );

    expect(ids).toContain('kb');
    // 学習系は引き続き出さない。
    expect(ids).not.toContain('courses');
    expect(ids).not.toContain('code');
  });
});
