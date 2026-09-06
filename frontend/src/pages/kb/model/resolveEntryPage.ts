import { KbRepository, getLastVisitedPageId } from '@/entities/kb';

/**
 * resolveEntryPageId は素の /kb(ページ ID 無し)で最初に開くページの ID を決める。
 *
 * 優先順位:
 *   1. workspaceSlug が指定されている(ヘッダー/サイドバーでワークスペースを切り替えた
 *      直後)なら、その中の最初のページ。切り替えた本人に「切り替えたのに変わらない」
 *      体験をさせないよう、直近の閲覧履歴より優先する。
 *   2. 直近に開いたページ(entities/kb/lib/lastVisitedPage.ts)。
 *   3. 所属する最初のワークスペース → 最初のスペース → 最初のページ(配列の順序=並び順)。
 *
 * どのスペースにも 1 枚もページが無ければ null(呼び出し側は「まだページがありません」を出す)。
 */
export async function resolveEntryPageId(workspaceSlug?: string): Promise<string | null> {
  if (!workspaceSlug) {
    const lastVisited = getLastVisitedPageId();
    if (lastVisited) return lastVisited;
  }

  const workspaces = workspaceSlug
    ? [{ slug: workspaceSlug }]
    : await KbRepository.fetchWorkspaces();

  for (const workspace of workspaces) {
    const spaces = await KbRepository.fetchSpaces(workspace.slug);
    for (const space of spaces) {
      const tree = await KbRepository.fetchPageTree(workspace.slug, space.id);
      const first = tree.pages[0]?.page.id;
      if (first) return first;
    }
  }
  return null;
}
