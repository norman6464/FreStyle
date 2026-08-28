import { Link } from 'react-router-dom';
import { KbPageIcon } from '@/shared/ui/icons/kb';
import type { KbPage, KbSpace } from '@/entities/knowledge-base';

export interface KbSearchResultsProps {
  /** 検索の状態。idle は「まだ何も探していない」（入力直後の待ち時間もここ）。 */
  status: 'loading' | 'done' | 'error';
  pages: KbPage[];
  /** スペース名を引くための一覧（結果はスペースごとに見出しを付けて並べる）。 */
  spaces: KbSpace[];
  workspaceSlug: string;
  activePageId?: string;
  onRetry: () => void;
}

/**
 * KbSearchResults は題名検索の結果。木の代わりにこの面が出る。
 *
 * 検索はサーバーが行う（ツリーと同じ規則で、閲覧できるページだけが返ってくる）。
 * 読み込み済みの木をフロントで絞る方式は採らない — 閉じているスペースや
 * まだ読み込んでいない枝の中が探せず、「検索したのに見つからない」が起きるため。
 *
 * 結果は木の形では出さない。一致した行の場所（祖先）は応答に含まれておらず、
 * ここで木を組み立てるとサーバーの伏せた祖先を推測で埋めることになる。
 * 場所の手掛かりはスペースの見出しまでとし、開けば本文がある。
 */
export default function KbSearchResults({
  status,
  pages,
  spaces,
  workspaceSlug,
  activePageId,
  onRetry,
}: KbSearchResultsProps) {
  if (status === 'loading') {
    return <p className="px-2 py-2 text-xs text-[var(--color-text-muted)]">検索中…</p>;
  }
  if (status === 'error') {
    return (
      <div className="px-2 py-2 text-xs text-red-600">
        <p>検索に失敗しました</p>
        <button type="button" onClick={onRetry} className="mt-0.5 underline hover:no-underline">
          再試行
        </button>
      </div>
    );
  }
  if (pages.length === 0) {
    return (
      <p className="px-2 py-2 text-xs text-[var(--color-text-muted)]">一致するページがありません</p>
    );
  }

  // スペースの並び（一覧の順）を保ったまま、結果をスペースごとに束ねる。
  const bySpace = new Map<string, KbPage[]>();
  for (const page of pages) {
    const list = bySpace.get(page.spaceId) ?? [];
    list.push(page);
    bySpace.set(page.spaceId, list);
  }
  const groups = spaces
    .filter((space) => bySpace.has(space.id))
    .map((space) => ({ space, pages: bySpace.get(space.id) ?? [] }));
  // 見えるスペース一覧に無いスペースのページ（例: 個別に許可されたページ）は
  // 名前が引けないので、末尾に見出しなしで並べる。落とすと「検索では返ったのに
  // 画面に出ない」という消え方をする。
  const known = new Set(spaces.map((space) => space.id));
  const orphan = pages.filter((page) => !known.has(page.spaceId));

  return (
    <div aria-label="検索結果">
      {groups.map(({ space, pages: groupPages }) => (
        <section key={space.id} className="mb-1">
          <h2 className="px-1 py-1 text-xs font-semibold uppercase tracking-wide text-[var(--color-text-muted)]">
            {space.name}
          </h2>
          <ul className="space-y-px">
            {groupPages.map((page) => (
              <li key={page.id}>
                <Link
                  to={`/kb/${workspaceSlug}/pages/${page.id}`}
                  aria-current={page.id === activePageId ? 'page' : undefined}
                  className={`flex min-w-0 items-center gap-1.5 rounded-md px-2 py-1 text-sm ${
                    page.id === activePageId
                      ? 'bg-brand-500/10 font-medium text-brand-600'
                      : 'text-[var(--color-text-primary)] hover:bg-surface-2'
                  }`}
                >
                  <KbPageIcon className="h-4 w-4 shrink-0 text-[var(--color-text-muted)]" />
                  <span className="truncate">{page.title}</span>
                </Link>
              </li>
            ))}
          </ul>
        </section>
      ))}
      {orphan.length > 0 && (
        <ul className="space-y-px">
          {orphan.map((page) => (
            <li key={page.id}>
              <Link
                to={`/kb/${workspaceSlug}/pages/${page.id}`}
                className="flex min-w-0 items-center gap-1.5 rounded-md px-2 py-1 text-sm text-[var(--color-text-primary)] hover:bg-surface-2"
              >
                <KbPageIcon className="h-4 w-4 shrink-0 text-[var(--color-text-muted)]" />
                <span className="truncate">{page.title}</span>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
