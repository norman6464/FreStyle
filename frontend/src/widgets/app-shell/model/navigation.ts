/** NavItem はアプリの主要ナビ 1 項目。ヘッダー・サイドバー・モバイルメニューで共用する。 */
export interface NavItem {
  id: string;
  label: string;
  to: string;
  matchExact?: boolean;
  /** 複数の URL 系統が同じ画面に属するとき（例: ノートの /notes と /p）は配列で並べる。 */
  matchPrefix?: string | string[];
}

/**
 * MAIN_NAV_ITEMS はアプリの主要ナビの正典（single source of truth）。
 * 項目を増やすときはここへ 1 つ足せば、ヘッダー・サイドバー・モバイルメニューすべてに反映される。
 */
export const MAIN_NAV_ITEMS: NavItem[] = [
  { id: 'home', label: 'ホーム', to: '/', matchExact: true },
  { id: 'code', label: '演習', to: '/code-editor', matchPrefix: '/code-editor' },
  { id: 'courses', label: 'コース', to: '/courses', matchPrefix: '/courses' },
  // ノートは共有される木（旧ナレッジを統合）。ページの URL は /p/{pageId}。
  { id: 'notes', label: 'ノート', to: '/notes', matchPrefix: ['/notes', '/p'] },
];

/**
 * navActive は現在の pathname がその項目を指しているかを判定する。
 *
 * matchPrefix は**パスの区切りまで見る**。素の startsWith だと `/kb` が `/kb-other` にも、
 * `/notes` が `/notes-foo` にも一致し、名前が前方一致するだけの無関係な画面でナビが光る。
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
