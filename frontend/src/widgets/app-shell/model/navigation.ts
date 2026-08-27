export type AppRole = 'super_admin' | 'company_admin' | 'trainee';

/** NavItem はアプリの主要ナビ 1 項目。ヘッダー・サイドバー・モバイルメニューで共用する。 */
export interface NavItem {
  id: string;
  label: string;
  to: string;
  matchExact?: boolean;
  matchPrefix?: string;
}

/** AdminSubItem は管理メニューの 1 項目。allowedRoles 未指定なら管理者ロール全員に出す。 */
export interface AdminSubItem {
  label: string;
  to: string;
  matchPrefix: string;
  allowedRoles?: ReadonlyArray<'super_admin' | 'company_admin'>;
}

/**
 * MAIN_NAV_ITEMS はアプリの主要ナビの正典（single source of truth）。
 * 項目を増やすときはここへ 1 つ足せば、ヘッダー・サイドバー・モバイルメニューすべてに反映される。
 */
export const MAIN_NAV_ITEMS: NavItem[] = [
  { id: 'home', label: 'ホーム', to: '/dashboard', matchExact: true },
  { id: 'ai', label: 'AI', to: '/chat/ask-ai', matchPrefix: '/chat/ask-ai' },
  { id: 'code', label: '演習', to: '/code-editor', matchPrefix: '/code-editor' },
  { id: 'courses', label: 'コース', to: '/courses', matchPrefix: '/courses' },
  { id: 'notes', label: 'ノート', to: '/notes', matchPrefix: '/notes' },
  // ナレッジ（/kb）はノートとは別の系統。ノートは自分のための平らな一覧、こちらは
  // 共有される木で、権限が付く。当面は 2 つ並ぶ（設計の「置く」— 移行は期限を決めてから）。
  { id: 'kb', label: 'ナレッジ', to: '/kb', matchPrefix: '/kb' },
  { id: 'reports', label: 'レポート', to: '/reports', matchExact: true },
];

// super_admin は企業管理に専念するロールなので**学習系**メニューは出さない。
// ナレッジは学習系ではなく書きもの・共有の面なので出す（運用の手順や決めごとを
// 書き残すのは、むしろ管理する側の仕事になる）。
const SUPER_ADMIN_MAIN_NAV_IDS = new Set(['home', 'kb']);

export const ADMIN_SUB_ITEMS: AdminSubItem[] = [
  { label: '概況', to: '/admin/dashboard', matchPrefix: '/admin/dashboard', allowedRoles: ['super_admin'] },
  { label: '会社一覧', to: '/admin/companies', matchPrefix: '/admin/companies', allowedRoles: ['super_admin'] },
  { label: '利用申請', to: '/admin/applications', matchPrefix: '/admin/applications', allowedRoles: ['super_admin'] },
  { label: '従業員一覧', to: '/admin/members', matchPrefix: '/admin/members' },
  { label: '招待管理', to: '/admin/invitations', matchPrefix: '/admin/invitations' },
  { label: '監査ログ', to: '/admin/audit', matchPrefix: '/admin/audit', allowedRoles: ['super_admin'] },
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
    return pathname === item.matchPrefix || pathname.startsWith(`${item.matchPrefix}/`);
  }
  return pathname === item.to;
}

/**
 * visibleMainNav はロールと機能フラグに応じて出す主要ナビを返す。
 * - super_admin はホームのみ
 * - AI チャットが受講者に無効なら trainee には出さない
 */
export function visibleMainNav(
  role: string | null,
  options: { aiChatEnabledForTrainees: boolean },
): NavItem[] {
  const base = role === 'super_admin'
    ? MAIN_NAV_ITEMS.filter((item) => SUPER_ADMIN_MAIN_NAV_IDS.has(item.id))
    : MAIN_NAV_ITEMS;
  return base.filter(
    (item) => !(item.id === 'ai' && role === 'trainee' && !options.aiChatEnabledForTrainees),
  );
}

/** visibleAdminSubs はロールに応じて出す管理メニューを返す。 */
export function visibleAdminSubs(role: string | null): AdminSubItem[] {
  return ADMIN_SUB_ITEMS.filter(
    (sub) =>
      !sub.allowedRoles ||
      (role !== null && sub.allowedRoles.includes(role as 'super_admin' | 'company_admin')),
  );
}

/** roleLabel はロールの日本語表示名を返す。 */
export function roleLabel(role: string | null): string {
  switch (role) {
    case 'super_admin':
      return '運営管理者';
    case 'company_admin':
      return '会社管理者';
    case 'trainee':
      return '受講者';
    default:
      return '';
  }
}
