import { useCallback, useEffect, useRef, useState } from 'react';
import { TicketRepository, type Ticket } from '@/entities/ticket';

const LOAD_FAILED = '子チケットを読み込めませんでした。時間をおいて開き直すと最新の状態が出ます。';

/**
 * useTicketChildren はチケット 1 件の直下の子（孫は含まない）を読む。
 *
 * 書き込みはここでは持たない — 子を作る・親を付け替える操作は、そのチケット自身を
 * 開いたときの「親」欄（useTicketPage 等）から行う。ここは表示専用の一覧。
 */
export function useTicketChildren(workspaceSlug: string | undefined, ticketId: string | undefined) {
  const [children, setChildren] = useState<Ticket[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const active = useRef<string | null>(null);

  const load = useCallback(async (slug: string, id: string) => {
    const key = `${slug} ${id}`;
    setLoading(true);
    setError(null);
    try {
      const list = await TicketRepository.fetchTicketChildren(slug, id);
      if (active.current !== key) return;
      setChildren(list);
    } catch {
      if (active.current !== key) return;
      setError(LOAD_FAILED);
    } finally {
      if (active.current === key) setLoading(false);
    }
  }, []);

  useEffect(() => {
    const key = workspaceSlug && ticketId ? `${workspaceSlug} ${ticketId}` : null;
    active.current = key;
    if (!workspaceSlug || !ticketId) {
      setChildren([]);
      return;
    }
    void load(workspaceSlug, ticketId);
  }, [workspaceSlug, ticketId, load]);

  const refresh = useCallback(() => {
    if (workspaceSlug && ticketId) void load(workspaceSlug, ticketId);
  }, [workspaceSlug, ticketId, load]);

  return { children, loading, error, refresh };
}
