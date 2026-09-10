import { useState } from 'react';
import { formatTicketKey, type Ticket, type TicketPriority, type TicketStatus } from '@/entities/ticket';
import type { KbGrantablePrincipal } from '@/entities/kb';
import { useTicketParentCandidates } from '../model/useTicketParentCandidates';
import TicketParentPicker from './TicketParentPicker';

const PRIORITY_LABEL: Record<TicketPriority, string> = { 1: '高', 2: '中', 3: '低' };

export interface TicketAttributePanelProps {
  ticket: Ticket;
  workspaceSlug: string;
  spaceKey: string;
  statuses: TicketStatus[];
  principals: KbGrantablePrincipal[];
  parentTicket: Ticket | undefined;
  canEdit: boolean;
  /** アーカイブ済みは全置換の編集を止める（状態と担当は専用の口なので止めない）。 */
  archived: boolean;
  busy: boolean;
  /** 全置換の下書きが持っている値（保存前の見た目をここに映す）。 */
  priority: TicketPriority;
  dueDate: string | null;
  onChangeStatus: (statusId: string) => void;
  onAssign: (principalId: string) => void;
  onUnassign: () => void;
  onChangePriority: (value: TicketPriority) => void;
  onChangeDueDate: (value: string | null) => void;
  onChangeParent: (parentId: string | null) => void;
}

/**
 * チケットの素性（状態・担当・優先度・期限・親）。
 *
 * 一覧の右のパネルと、チケットを開いた全画面の両方が同じものを出すので、部品として
 * 切り出してある。状態と担当は専用の口があるので即座に送るが、優先度と期限は
 * 全置換に載るため下書きを経由する（値は呼び出し側の下書きから渡ってくる）。
 */
export default function TicketAttributePanel({
  ticket,
  workspaceSlug,
  spaceKey,
  statuses,
  principals,
  parentTicket,
  canEdit,
  archived,
  busy,
  priority,
  dueDate,
  onChangeStatus,
  onAssign,
  onUnassign,
  onChangePriority,
  onChangeDueDate,
  onChangeParent,
}: TicketAttributePanelProps) {
  const assigneeUsers = principals.filter((p) => p.kind === 'user');
  const [parentPickerOpen, setParentPickerOpen] = useState(false);
  // 開くまで問い合わせない（親を触らないチケットのほうが多く、毎回スペース全件を
  // 引くのは無駄なため）。
  const parentCandidatesQuery = useTicketParentCandidates(
    parentPickerOpen ? workspaceSlug : undefined,
    parentPickerOpen ? ticket.spaceId : undefined,
  );

  const handleSelectParent = (parentId: string | null) => {
    onChangeParent(parentId);
    setParentPickerOpen(false);
  };

  return (
    <dl className="mb-4 grid grid-cols-[64px_1fr] gap-x-2 gap-y-2 text-sm">
      <dt className="text-[var(--color-text-muted)]">状態</dt>
      <dd>
        {canEdit ? (
          <select
            value={ticket.statusId}
            disabled={busy}
            onChange={(e) => onChangeStatus(e.target.value)}
            aria-label="状態"
            className="rounded border border-surface-3 bg-surface-1 px-1.5 py-0.5 text-sm"
          >
            {statuses.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        ) : (
          statuses.find((s) => s.id === ticket.statusId)?.name ?? ''
        )}
      </dd>

      <dt className="text-[var(--color-text-muted)]">担当</dt>
      <dd>
        {canEdit ? (
          <select
            value={ticket.assigneePrincipalId ?? ''}
            disabled={busy}
            onChange={(e) => {
              const value = e.target.value;
              if (value) onAssign(value);
              else onUnassign();
            }}
            aria-label="担当"
            className="rounded border border-surface-3 bg-surface-1 px-1.5 py-0.5 text-sm"
          >
            <option value="">未割り当て</option>
            {assigneeUsers.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name || p.id}
              </option>
            ))}
          </select>
        ) : (
          assigneeUsers.find((p) => p.id === ticket.assigneePrincipalId)?.name || '未割り当て'
        )}
      </dd>

      <dt className="text-[var(--color-text-muted)]">優先度</dt>
      <dd>
        {canEdit && !archived ? (
          <select
            value={priority}
            onChange={(e) => onChangePriority(Number(e.target.value) as TicketPriority)}
            aria-label="優先度"
            className="rounded border border-surface-3 bg-surface-1 px-1.5 py-0.5 text-sm"
          >
            <option value={1}>高</option>
            <option value={2}>中</option>
            <option value={3}>低</option>
          </select>
        ) : (
          <span className={ticket.priority === 1 ? 'font-bold text-brand-700' : undefined}>
            {PRIORITY_LABEL[ticket.priority]}
          </span>
        )}
      </dd>

      <dt className="text-[var(--color-text-muted)]">期限</dt>
      <dd>
        {canEdit && !archived ? (
          <input
            type="date"
            value={dueDate ?? ''}
            onChange={(e) => onChangeDueDate(e.target.value || null)}
            aria-label="期限"
            className="rounded border border-surface-3 bg-surface-1 px-1.5 py-0.5 text-sm"
          />
        ) : (
          ticket.dueDate ?? <span className="text-[var(--color-text-muted)]">なし</span>
        )}
      </dd>

      <dt className="text-[var(--color-text-muted)]">親</dt>
      <dd>
        {canEdit && !archived ? (
          <>
            <button
              type="button"
              onClick={() => setParentPickerOpen((v) => !v)}
              disabled={busy}
              aria-expanded={parentPickerOpen}
              aria-label="親を変更"
              className="rounded border border-surface-3 bg-surface-1 px-1.5 py-0.5 text-sm hover:bg-surface-2 disabled:opacity-50"
            >
              {parentTicket ? formatTicketKey(spaceKey, parentTicket.number) : 'なし'}
            </button>
            {parentPickerOpen && (
              <TicketParentPicker
                candidates={parentCandidatesQuery.candidates.filter((c) => c.id !== ticket.id)}
                loading={parentCandidatesQuery.loading}
                error={parentCandidatesQuery.error}
                spaceKey={spaceKey}
                currentParentId={ticket.parentId}
                onSelect={handleSelectParent}
              />
            )}
          </>
        ) : parentTicket ? (
          formatTicketKey(spaceKey, parentTicket.number)
        ) : (
          <span className="text-[var(--color-text-muted)]">なし</span>
        )}
      </dd>
    </dl>
  );
}
