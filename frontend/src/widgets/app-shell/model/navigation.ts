/** NavItem はアプリの主要ナビ 1 項目。ヘッダー・サイドバー・モバイルメニューで共用する。 */
export interface NavItem {
  id: string;
  label: string;
  to: string;
  matchExact?: boolean;
  /** 複数の URL 系統が同じ画面に属するときは配列で並べる（いまはどの項目も 1 系統）。 */
  matchPrefix?: string | string[];
}

/**
 * MAIN_NAV_ITEMS はアプリの主要ナビの正典（single source of truth）。
 * 項目を増やすときはここへ 1 つ足せば、ヘッダー・サイドバー・モバイルメニューすべてに反映される。
 */
export const MAIN_NAV_ITEMS: NavItem[] = [
  { id: 'home', label: 'ホーム', to: '/', matchExact: true },
  { id: 'code', label: '演習', to: '/code-editor', matchPrefix: '/code-editor' },
  // ナレッジは共有される木（workspaces → spaces → pages）。to の /kb はページ未選択の入口で、
  // resolveEntryPageId（pages/kb/model/resolveEntryPage.ts）が続きのページへ即座に移す。
  { id: 'kb', label: 'ナレッジ', to: '/kb', matchPrefix: '/kb' },
];

/**
 * navActive は現在の pathname がその項目を指しているかを判定する。
 *
 * matchPrefix は**パスの区切りまで見る**。素の startsWith だと `/kb` が `/kb-other` にも
 * 一致し、名前が前方一致するだけの無関係な画面でナビが光る。
 * 一致してよいのは、そのものか、`/` で続く下の階層だけ。
 */
export function navActive(item: NavItem, pathname: string): boolean {
  if (item.matchExact) return pathname === item.to;
  if (item.matchPrefix) {
    const prefixes = Array.isArray(item.matchPrefix) ? item.matchPrefix : [item.matchPrefix];
    return prefixes.some((prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`));
  }
  return pathname === item.to;
}
