import { ExclamationCircleIcon, InboxIcon } from '@heroicons/react/24/outline';
import type { Ticket, TicketStatus, TicketType } from '@/entities/ticket';
import EmptyState from '@/shared/ui/EmptyState';
import BacklogRow from './BacklogRow';
import BacklogReorderBar from './BacklogReorderBar';
import TicketCreateRow from './TicketCreateRow';

export interface BacklogListProps {
  tickets: Ticket[];
  statuses: TicketStatus[];
  types: TicketType[];
  spaceKey: string;
  loading: boolean;
  error: string | null;
  archived: boolean;
  canEdit: boolean;
  selectedId: string | null;
  busyId: string | null;
  nameOf: (principalId: string | null) => string;
  initialsOf: (principalId: string | null) => string;
  onSelect: (ticketId: string) => void;
  onCreate: (title: string) => Promise<void>;
  onMove: (ticketId: string, input: { anchorTicketId?: string; anchorAfter?: boolean }) => Promise<void>;
  onRetry: () => void;
}

/** バックログの「チケット」タブの中身（一覧 + 並び替えの帯）。設計 Ⅲ・見本 2a。 */
export default function BacklogList({
  tickets,
  statuses,
  types,
  spaceKey,
  loading,
  error,
  archived,
  canEdit,
  selectedId,
  busyId,
  nameOf,
  initialsOf,
  onSelect,
  onCreate,
  onMove,
  onRetry,
}: BacklogListProps) {
  if (error) {
    return (
      <EmptyState
        icon={ExclamationCircleIcon}
        title="チケットを読み込めませんでした"
        description={error}
        action={{ label: '再読み込み', onClick: onRetry }}
      />
    );
  }

  if (!loading && tickets.length === 0) {
    return (
      <EmptyState
        icon={InboxIcon}
        title="まだチケットがありません"
        description="題名だけで作れます。種別と状態は雛形の初期値が入ります。並び替えは 2 件目から出ます。"
      />
    );
  }

  const statusOf = (id: string) => statuses.find((s) => s.id === id);
  const typeOf = (id: string) => types.find((t) => t.id === id);
  const selectedIndex = selectedId ? tickets.findIndex((t) => t.id === selectedId) : -1;
  const selected = selectedIndex >= 0 ? tickets[selectedIndex] : null;

  const handleMoveUp = () => {
    if (!selected || selectedIndex <= 0) return;
    const anchor = tickets[selectedIndex - 1];
    void onMove(selected.id, { anchorTicketId: anchor.id, anchorAfter: false });
  };
  const handleMoveDown = () => {
    if (!selected || selectedIndex < 0 || selectedIndex >= tickets.length - 1) return;
    const anchor = tickets[selectedIndex + 1];
    void onMove(selected.id, { anchorTicketId: anchor.id, anchorAfter: true });
  };
  const handleMoveLast = () => {
    if (!selected) return;
    void onMove(selected.id, {});
  };

  return (
    <div className="flex h-full min-h-0 flex-col">
      {!archived && (
        <div className="flex items-center gap-2 border-b border-surface-3 px-3 py-2">
          <button type="button" disabled className="rounded-full border border-surface-3 px-3 py-1 text-xs text-[var(--color-text-muted)]">
            状態: すべて
          </button>
          <button type="button" disabled className="rounded-full border border-surface-3 px-3 py-1 text-xs text-[var(--color-text-muted)]">
            種別: すべて
          </button>
        </div>
      )}

      <div
        className="grid items-center gap-2.5 border-b border-surface-3 px-3 py-1.5 text-[10.5px] font-semibold uppercase tracking-wide text-[var(--color-text-muted)]"
        style={{ gridTemplateColumns: '24px 1fr 32px 48px 96px' }}
        aria-hidden="true"
      >
        <span>担当</span>
        <span>キー・種別 / 題名</span>
        <span className="text-right">優先度</span>
        <span className="text-right">期限</span>
        <span className="text-right">状態</span>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto">
        {tickets.map((ticket) => (
          <BacklogRow
            key={ticket.id}
            ticket={ticket}
            spaceKey={spaceKey}
            type={typeOf(ticket.typeId)}
            status={statusOf(ticket.statusId)}
            assigneeName={nameOf(ticket.assigneePrincipalId)}
            assigneeInitials={initialsOf(ticket.assigneePrincipalId)}
            selected={ticket.id === selectedId}
            busy={ticket.id === busyId}
            indented={ticket.parentId !== null}
            canEdit={canEdit}
            onOpen={() => onSelect(ticket.id)}
          />
        ))}
        {canEdit && !archived && <TicketCreateRow onCreate={onCreate} />}
      </div>

      {canEdit && !archived && (
        <BacklogReorderBar
          selectedKey={selected ? `${spaceKey.toUpperCase()}-${selected.number}` : null}
          isFirst={selectedIndex <= 0}
          isLast={selectedIndex < 0 || selectedIndex >= tickets.length - 1}
          onMoveUp={handleMoveUp}
          onMoveDown={handleMoveDown}
          onMoveLast={handleMoveLast}
        />
      )}
    </div>
  );
}
