import type { TicketStatus } from '@/entities/ticket';
import { useTicketChildren } from '../model/useTicketChildren';
import TicketChildrenList from './TicketChildrenList';

export interface TicketChildrenSectionProps {
  workspaceSlug: string;
  ticketId: string;
  spaceKey: string;
  statuses: TicketStatus[];
}

/**
 * TicketChildrenSection は直下の子の一式（取得のみ）をまとめる。
 *
 * 全画面の副列とバックログの副パネルの両方から使う（TicketCommentSection /
 * TicketAttachmentSection と同じ考え方 — 互いに同時マウントされない別ルートなので、
 * それぞれが自分の useTicketChildren を持ってよい）。
 */
export default function TicketChildrenSection({ workspaceSlug, ticketId, spaceKey, statuses }: TicketChildrenSectionProps) {
  const { children, loading, error } = useTicketChildren(workspaceSlug, ticketId);
  return <TicketChildrenList tickets={children} loading={loading} error={error} spaceKey={spaceKey} statuses={statuses} />;
}
