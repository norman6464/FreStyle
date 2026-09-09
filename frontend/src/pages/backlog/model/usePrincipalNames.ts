import { useCallback, useEffect, useRef, useState } from 'react';
import { KbRepository, type KbGrantablePrincipal } from '@/entities/kb';

/**
 * usePrincipalNames はワークスペース内の「権限を張れる相手」を principalId → 表示名の
 * 対応表として読む（設計 Ⅳ-G）。担当者アバターの名前・担当を選ぶプルダウンの選択肢の
 * 両方をこれ 1 つでまかなう。
 *
 * **弱点（設計に明記済みの妥協）**: 相手を引く口 `pagePrincipals` はページ ID を取る。
 * チケットしか無いスペースには渡すページが無いので、ワークスペース内のスペースを
 * 順に見て最初に見つかったページで代表させる。1 枚も見つからなければ対応表は空のまま
 * （呼び出し側は principalId の先頭だけで頭文字を出す）。
 */
export function usePrincipalNames(workspaceSlug: string | undefined) {
  const [principals, setPrincipals] = useState<KbGrantablePrincipal[]>([]);
  const [loading, setLoading] = useState(false);
  const active = useRef<string | null>(null);

  const load = useCallback(async (slug: string) => {
    setLoading(true);
    try {
      const spaces = await KbRepository.fetchSpaces(slug);
      let resolved: KbGrantablePrincipal[] = [];
      for (const space of spaces) {
        if (active.current !== slug) return;
        const tree = await KbRepository.fetchPageTree(slug, space.id);
        const firstPage = tree.pages[0]?.page;
        if (!firstPage) continue;
        resolved = await KbRepository.listGrantablePrincipals(slug, firstPage.id);
        break;
      }
      if (active.current === slug) {
        setPrincipals(resolved);
      }
    } catch {
      if (active.current === slug) {
        setPrincipals([]);
      }
    } finally {
      if (active.current === slug) {
        setLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    active.current = workspaceSlug ?? null;
    if (!workspaceSlug) {
      setPrincipals([]);
      return;
    }
    void load(workspaceSlug);
  }, [workspaceSlug, load]);

  const nameOf = useCallback(
    (principalId: string | null): string => {
      if (!principalId) return '';
      return principals.find((p) => p.id === principalId)?.name ?? '';
    },
    [principals],
  );

  /** アバターの頭文字。名前が引ければ先頭 2 文字、引けなければ principalId の先頭 2 文字。 */
  const initialsOf = useCallback(
    (principalId: string | null): string => {
      if (!principalId) return '';
      const name = nameOf(principalId);
      const basis = name || principalId;
      return basis.slice(0, 2).toUpperCase();
    },
    [nameOf],
  );

  return { principals, loading, nameOf, initialsOf };
}
