import { useEffect, useRef, useState } from 'react';
import { TicketRepository, type TicketChangeGroup } from '@/entities/ticket';

/**
 * useTicketDetail は詳細パネルが選択中のチケットの変更履歴だけを読む。
 *
 * チケット本体（title / doc / 状態 等）は一覧の応答に既に全項目が入っているため、
 * `useTicketList` の配列から `find` すれば足りる（別に取得しない）。この hook が
 * 持つのは、一覧の応答に含まれない履歴（`GET .../history`）だけ。
 */
export function useTicketDetail(workspaceSlug: string | undefined, ticketId: string | null) {
  const [history, setHistory] = useState<TicketChangeGroup[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const active = useRef<string | null>(null);

  useEffect(() => {
    const key = workspaceSlug && ticketId ? `${workspaceSlug} ${ticketId}` : null;
    active.current = key;
    if (!workspaceSlug || !ticketId) {
      setHistory([]);
      setError(null);
      return;
    }
    setLoading(true);
    setError(null);
    TicketRepository.fetchTicketHistory(workspaceSlug, ticketId)
      .then((groups) => {
        if (active.current !== key) return;
        setHistory(groups);
      })
      .catch(() => {
        if (active.current !== key) return;
        setHistory([]);
        setError('変更履歴を読み込めませんでした。');
      })
      .finally(() => {
        if (active.current === key) setLoading(false);
      });
  }, [workspaceSlug, ticketId]);

  return { history, loading, error };
}
