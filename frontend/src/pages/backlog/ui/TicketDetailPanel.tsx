import { lazy, Suspense } from 'react';
import {
  TicketKeyBadge,
  type Ticket,
  type TicketStatus,
  type TicketType,
  type UpdateTicketInput,
} from '@/entities/ticket';
import type { KbGrantablePrincipal } from '@/entities/kb';
import Loading from '@/shared/ui/Loading';
import { SaveStatusIndicator, emptyRichDoc, isRichDoc } from '@/shared/ui/RichTextEditor';
import { useTicketEditor } from '../model/useTicketEditor';
import TicketAttributePanel from './TicketAttributePanel';
import TicketLabelChip from './TicketLabelChip';
import TicketSection from './TicketSection';

const RichTextEditor = lazy(() => import('@/shared/ui/RichTextEditor').then((m) => ({ default: m.RichTextEditor })));

export interface TicketDetailPanelProps {
  ticket: Ticket;
  spaceKey: string;
  statuses: TicketStatus[];
  types: TicketType[];
  principals: KbGrantablePrincipal[];
  parentTicket: Ticket | undefined;
  canEdit: boolean;
  busy: boolean;
  onUpdate: (ticketId: string, input: UpdateTicketInput) => Promise<Ticket>;
  onChangeStatus: (statusId: string) => Promise<void>;
  onAssign: (principalId: string) => Promise<void>;
  onUnassign: () => Promise<void>;
  onArchive: () => Promise<void>;
  onRestore: () => Promise<void>;
}

/**
 * チケット詳細パネル。一覧を捌きながら 1 件を確かめ、軽く直すための面。
 *
 * 見出しと閉じるボタンは器（SecondaryPanel）が描く。ここで同じ見出しをもう 1 行出すと
 * 二重になるので持たない。保存状態は「本文」の節の見出しに添える。
 */
export default function TicketDetailPanel({
  ticket,
  spaceKey,
  statuses,
  types,
  principals,
  parentTicket,
  canEdit,
  busy,
  onUpdate,
  onChangeStatus,
  onAssign,
  onUnassign,
  onArchive,
  onRestore,
}: TicketDetailPanelProps) {
  const type = types.find((t) => t.id === ticket.typeId);
  const archived = ticket.archivedAt !== null;

  const editor = useTicketEditor(ticket, canEdit && !archived, (input) => onUpdate(ticket.id, input));
  const docValue = isRichDoc(editor.doc) ? editor.doc : emptyRichDoc();

  return (
    <div className="min-h-0 flex-1 overflow-y-auto px-3 py-3" tabIndex={0}>
      <div className="mb-1.5 flex items-center gap-2">
        <span className="rounded bg-surface-2 px-1.5 py-0.5 text-[11px] font-semibold text-[var(--color-text-secondary)]">
          {type?.name ?? ''}
        </span>
        <TicketKeyBadge spaceKey={spaceKey} number={ticket.number} />
      </div>

      {ticket.labels.length > 0 && (
        <div className="mb-2 flex flex-wrap gap-1.5">
          {ticket.labels.map((label) => (
            <TicketLabelChip key={label.id} label={label} />
          ))}
        </div>
      )}

      {canEdit && !archived ? (
        <input
          type="text"
          value={editor.title}
          onChange={(e) => editor.changeTitle(e.target.value)}
          onBlur={editor.commitTitle}
          aria-label="題名"
          className="mb-3 w-full bg-transparent text-lg font-bold text-[var(--color-text-primary)] focus:outline-none"
        />
      ) : (
        <h6 className="mb-3 text-lg font-bold text-[var(--color-text-primary)]">{ticket.title}</h6>
      )}

      <TicketAttributePanel
        ticket={ticket}
        spaceKey={spaceKey}
        statuses={statuses}
        principals={principals}
        parentTicket={parentTicket}
        canEdit={canEdit}
        archived={archived}
        busy={busy}
        priority={editor.priority}
        dueDate={editor.dueDate}
        onChangeStatus={(statusId) => void onChangeStatus(statusId)}
        onAssign={(principalId) => void onAssign(principalId)}
        onUnassign={() => void onUnassign()}
        onChangePriority={editor.changePriority}
        onChangeDueDate={editor.changeDueDate}
      />

      <TicketSection title="本文" action={<SaveStatusIndicator status={editor.saveStatus} />}>
        <Suspense fallback={<Loading />}>
          <RichTextEditor
            value={docValue}
            editable={canEdit && !archived}
            onChange={editor.changeDoc}
            ariaLabel="チケットの本文"
            placeholder="本文を書く"
          />
        </Suspense>
      </TicketSection>

      {canEdit && (
        <TicketSection title="操作">
          <button
            type="button"
            onClick={() => void (archived ? onRestore() : onArchive())}
            disabled={busy}
            className="rounded border border-surface-3 px-2.5 py-1 text-xs font-medium text-[var(--color-text-secondary)] hover:bg-surface-2 disabled:opacity-50"
          >
            {archived ? '現役に戻す' : 'アーカイブ'}
          </button>
        </TicketSection>
      )}
    </div>
  );
}
