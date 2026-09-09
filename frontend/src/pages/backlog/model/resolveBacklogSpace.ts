import { KbRepository, type KbSpace } from '@/entities/kb';

export interface ResolvedBacklogSpace {
  workspaceSlug: string;
  space: KbSpace;
}

/**
 * resolveBacklogSpaceId は素の /kb/backlog（スペース未指定）で最初に開くスペースの ID を
 * 決める。所属する最初のワークスペース → 最初のスペース（配列の順序=並び順）。
 * どのワークスペースにもスペースが無ければ null。
 */
export async function resolveBacklogSpaceId(): Promise<string | null> {
  const workspaces = await KbRepository.fetchWorkspaces();
  for (const workspace of workspaces) {
    const spaces = await KbRepository.fetchSpaces(workspace.slug);
    if (spaces[0]) return spaces[0].id;
  }
  return null;
}

/**
 * resolveBacklogSpace は spaceId からワークスペースを引く（設計 Ⅳ-H）。
 *
 * `/kb/backlog/:spaceId` は URL にワークスペースを出さない既存の規則を踏襲するが、
 * チケットと違って spaceId から直接ワークスペースを引く backend の口が無い
 * （設計時点で見送り。段2の候補）。ここでは所属ワークスペースを順に見て
 * スペース一覧からその ID を探す。実データでワークスペース数は 2 つ・スペース一覧は
 * 軽い口なので、いまはこれで足りる。
 */
export async function resolveBacklogSpace(spaceId: string): Promise<ResolvedBacklogSpace | null> {
  const workspaces = await KbRepository.fetchWorkspaces();
  for (const workspace of workspaces) {
    const spaces = await KbRepository.fetchSpaces(workspace.slug);
    const space = spaces.find((s) => s.id === spaceId);
    if (space) return { workspaceSlug: workspace.slug, space };
  }
  return null;
}
