import KbRepository from '../api/kbRepository';
import type { KbMySpace } from './types';

export interface ResolvedKbSpace {
  workspaceSlug: string;
  space: KbMySpace;
}

/**
 * resolveEntryKbSpaceId は素の /kb/spaces（スペース未指定）で最初に開くスペースの ID を
 * 決める。所属する最初のワークスペース → 自分がアクセスできる最初のスペース
 * （配列の順序=並び順）。どのワークスペースにもアクセスできるスペースが無ければ null。
 *
 * pages/backlog/model/resolveBacklogSpace.ts と同じ形だが、fetchSpaces（可視スペース
 * 全件）ではなく fetchMySpaces（自分の役割つき）を使う。バックログ側は今回のスコープ外
 * として触らない（動いている別機能への影響を避けるため、あえて共有しない）。
 */
export async function resolveEntryKbSpaceId(): Promise<string | null> {
  const workspaces = await KbRepository.fetchWorkspaces();
  for (const workspace of workspaces) {
    const spaces = await KbRepository.fetchMySpaces(workspace.slug);
    if (spaces[0]) return spaces[0].id;
  }
  return null;
}

/**
 * resolveKbSpace は spaceId からワークスペースを引く。spaceId から直接ワークスペースを
 * 引く backend の口が無いため、所属ワークスペースを順に見てスペース一覧からその ID を
 * 探す（resolveBacklogSpace と同じ理由・同じ制約）。
 */
export async function resolveKbSpace(spaceId: string): Promise<ResolvedKbSpace | null> {
  const workspaces = await KbRepository.fetchWorkspaces();
  for (const workspace of workspaces) {
    const spaces = await KbRepository.fetchMySpaces(workspace.slug);
    const space = spaces.find((s) => s.id === spaceId);
    if (space) return { workspaceSlug: workspace.slug, space };
  }
  return null;
}
