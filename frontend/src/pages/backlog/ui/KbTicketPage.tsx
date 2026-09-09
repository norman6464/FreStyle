import { useEffect, useState } from 'react';
import { Navigate, useParams } from 'react-router-dom';
import { TicketRepository } from '@/entities/ticket';
import { getApiError } from '@/shared/lib/classifyApiError';
import Loading from '@/shared/ui/Loading';

/**
 * KbTicketPage は `/kb/tickets/:ticketId`（ワークスペースを URL に持たない解決の口）の
 * 受け皿。設計 Ⅳ-H・kb の `/kb/:pageId` と同じ役割 — 通知・本文中の ticketRef の href・
 * ブックマークからの再訪はワークスペースを知らないまま来るので、ID だけで開ける必要がある。
 *
 * バックログ画面自体はスペース単位（`/kb/backlog/:spaceId`）なので、ここでは
 * `GET /kb/tickets/:id` でチケットの所在（spaceId）を解決し、そのスペースの
 * バックログへ差し替える。チケット自体の選択（詳細パネルを開く）まではしない
 * （段 1 の簡略化 — アーカイブ済みだと現役タブに現れず選べないため）。
 */
export default function KbTicketPage() {
  const { ticketId } = useParams<{ ticketId: string }>();
  const [state, setState] = useState<{ spaceId: string | null; error: string | null }>({
    spaceId: null,
    error: null,
  });

  useEffect(() => {
    if (!ticketId) return;
    let active = true;
    TicketRepository.resolveTicket(ticketId)
      .then((resolved) => {
        if (active) setState({ spaceId: resolved.ticket.spaceId, error: null });
      })
      .catch((cause) => {
        if (!active) return;
        setState({
          spaceId: null,
          error: getApiError(cause).status === 404 ? 'チケットが見つかりませんでした。' : 'チケットを開けませんでした。',
        });
      });
    return () => {
      active = false;
    };
  }, [ticketId]);

  if (state.error) {
    return (
      <div className="flex h-full items-center justify-center px-6 text-center text-sm text-[var(--color-text-muted)]">
        {state.error}
      </div>
    );
  }

  if (state.spaceId) {
    return <Navigate to={`/kb/backlog/${state.spaceId}`} replace />;
  }

  return <Loading />;
}
