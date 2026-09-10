import { TicketKeyBadge, TicketStatusPill, type Ticket, type TicketStatus, type TicketType } from '@/entities/ticket';
import TicketLabelChip from './TicketLabelChip';

/** 一覧の行では場所を取りすぎないよう、ラベルは最大でこの件数だけチップにし、残りは件数へ畳む。 */
const MAX_VISIBLE_LABELS = 2;

export interface BacklogRowProps {
  ticket: Ticket;
  /** チケットが属するスペースの key（表示キーの組み立てに使う。例 "FRESTYLE"）。 */
  spaceKey: string;
  type: TicketType | undefined;
  status: TicketStatus | undefined;
  assigneeName: string;
  assigneeInitials: string;
  selected: boolean;
  busy: boolean;
  /** 子チケット（parentId が現在見えている親を指す）なら字下げを出す。 */
  indented: boolean;
  canEdit: boolean;
  onOpen: () => void;
}

const PRIORITY_LABEL: Record<number, string> = { 1: '高', 2: '中', 3: '低' };

function isOverdue(dueDate: string | null): boolean {
  if (!dueDate) return false;
  const today = new Date().toISOString().slice(0, 10);
  return dueDate < today;
}

function formatDue(dueDate: string | null): string {
  if (!dueDate) return '—';
  // 'YYYY-MM-DD' → 'MM/DD'（見本と同じ短縮表記）。
  return dueDate.slice(5).replace('-', '/');
}

/** バックログ一覧の行 1 件（見本 2a・行型）。アバター先頭・キー+種別/題名・右に優先度/期限/状態。 */
export default function BacklogRow({
  ticket,
  spaceKey,
  type,
  status,
  assigneeName,
  assigneeInitials,
  selected,
  busy,
  indented,
  canEdit,
  onOpen,
}: BacklogRowProps) {
  const done = status?.category === 'done';
  const over = isOverdue(ticket.dueDate);

  return (
    <button
      type="button"
      onClick={onOpen}
      aria-current={selected}
      aria-busy={busy || undefined}
      className={`flex w-full items-center gap-2.5 border-b border-surface-3 px-3 py-2.5 text-left text-sm transition-colors last:border-b-0 ${
        selected ? 'bg-surface-3' : 'hover:bg-surface-2'
      }`}
    >
      {ticket.assigneePrincipalId ? (
        <span
          className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-taupe-500 text-[10px] font-bold text-white"
          title={assigneeName || undefined}
        >
          {assigneeInitials}
        </span>
      ) : (
        <span
          className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full border border-dashed border-surface-3 text-[var(--color-text-muted)]"
          aria-label="未割り当て"
        >
          –
        </span>
      )}

      <span className="min-w-0 flex-1">
        <span className="flex items-center gap-1.5 text-[10.5px] font-bold tracking-wide text-[var(--color-text-muted)]">
          <TicketKeyBadge spaceKey={spaceKey} number={ticket.number} className="tabular-nums" />
          <span>{type?.name ?? ''}</span>
        </span>
        <span className="flex items-center gap-1.5">
          <span
            className={`min-w-0 truncate ${done ? 'text-[var(--color-text-muted)] line-through' : 'text-[var(--color-text-primary)]'}`}
          >
            {indented && (
              <span className="mr-1 text-[var(--color-text-muted)]" aria-hidden="true">
                └
              </span>
            )}
            {ticket.title}
          </span>
          {ticket.labels.length > 0 && (
            <span className="flex flex-shrink-0 items-center gap-1">
              {ticket.labels.slice(0, MAX_VISIBLE_LABELS).map((label) => (
                <TicketLabelChip key={label.id} label={label} />
              ))}
              {ticket.labels.length > MAX_VISIBLE_LABELS && (
                <span className="text-[11px] text-[var(--color-text-muted)]">
                  +{ticket.labels.length - MAX_VISIBLE_LABELS}
                </span>
              )}
            </span>
          )}
        </span>
      </span>

      <span
        className={`w-8 flex-shrink-0 text-right text-xs ${
          ticket.priority === 1 ? 'font-bold text-brand-700' : 'text-[var(--color-text-muted)]'
        }`}
      >
        {PRIORITY_LABEL[ticket.priority]}
      </span>

      <span
        className={`w-12 flex-shrink-0 text-right text-xs tabular-nums ${
          over ? 'font-semibold text-red-600' : 'text-[var(--color-text-muted)]'
        }`}
      >
        {formatDue(ticket.dueDate)}
      </span>

      <span className="w-24 flex-shrink-0 text-right">
        {status && (
          <TicketStatusPill
            name={status.name}
            color={status.color}
            category={status.category}
            showChevron={canEdit && !busy}
          />
        )}
      </span>
    </button>
  );
}
