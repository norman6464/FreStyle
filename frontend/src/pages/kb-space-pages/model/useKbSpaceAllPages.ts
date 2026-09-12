import { useCallback, useEffect, useState } from 'react';
import { KbRepository, type KbPage, type KbPageTreeNode } from '@/entities/kb';

export interface KbFlatPage {
  page: KbPage;
  depth: number;
}

/** flatten は木を深さ優先で平坦な一覧に開く（親の直後に子が並ぶ）。 */
function flatten(nodes: KbPageTreeNode[], depth: number): KbFlatPage[] {
  const out: KbFlatPage[] = [];
  for (const node of nodes) {
    out.push({ page: node.page, depth });
    out.push(...flatten(node.children, depth + 1));
  }
  return out;
}

export function useKbSpaceAllPages(workspaceSlug: string, spaceId: string) {
  const [pages, setPages] = useState<KbFlatPage[]>([]);
  const [hasHiddenChildren, setHasHiddenChildren] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(() => {
    setLoading(true);
    setError(null);
    KbRepository.fetchPageTree(workspaceSlug, spaceId)
      .then((tree) => {
        setPages(flatten(tree.pages, 0));
        setHasHiddenChildren(tree.hasHiddenChildren);
      })
      .catch(() => {
        setError('ページを読み込めませんでした。');
      })
      .finally(() => {
        setLoading(false);
      });
  }, [workspaceSlug, spaceId]);

  useEffect(() => {
    load();
  }, [load]);

  return { pages, hasHiddenChildren, loading, error, retry: load };
}
