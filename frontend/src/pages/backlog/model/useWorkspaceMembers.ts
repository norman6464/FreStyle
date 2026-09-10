import { useCallback, useEffect, useRef, useState } from 'react';
import { KbRepository, type KbWorkspaceMember } from '@/entities/kb';

/**
 * useWorkspaceMembers はワークスペースに属する人を読む（発言での名指しの候補・表示名解決用）。
 *
 * ラベルと同じく応答をそのまま手元へ持つだけの素朴な一覧（書き込みはしない）。
 */
export function useWorkspaceMembers(workspaceSlug: string | undefined) {
  const [members, setMembers] = useState<KbWorkspaceMember[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const active = useRef<string | null>(null);

  const load = useCallback(async (slug: string) => {
    setLoading(true);
    setError(null);
    try {
      const list = await KbRepository.fetchMembers(slug);
      if (active.current !== slug) return;
      setMembers(list);
    } catch {
      if (active.current !== slug) return;
      setError('候補を読み込めませんでした。');
    } finally {
      if (active.current === slug) setLoading(false);
    }
  }, []);

  useEffect(() => {
    active.current = workspaceSlug ?? null;
    if (!workspaceSlug) {
      setMembers([]);
      return;
    }
    void load(workspaceSlug);
  }, [workspaceSlug, load]);

  return { members, loading, error };
}
