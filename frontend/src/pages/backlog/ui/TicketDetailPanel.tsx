import { lazy, Suspense } from 'react';
import {
  TicketKeyBadge,
  type Label,
  type Ticket,
  type TicketStatus,
  type TicketType,
  type UpdateTicketInput,
} from '@/entities/ticket';
import type { KbGrantablePrincipal } from '@/entities/kb';
import Loading from '@/shared/ui/Loading';
import { SaveStatusIndicator, emptyRichDoc, isRichDoc } from '@/shared/ui/RichTextEditor';
import { useTicketEditor } from '../model/useTicketEditor';
import TicketAttachmentSection from './TicketAttachmentSection';
import TicketAttributePanel from './TicketAttributePanel';
import TicketChildrenSection from './TicketChildrenSection';
import TicketCommentSection from './TicketCommentSection';
import TicketLabelBar from './TicketLabelBar';
import TicketSection from './TicketSection';

const RichTextEditor = lazy(() => import('@/shared/ui/RichTextEditor').then((m) => ({ default: m.RichTextEditor })));

export interface TicketDetailPanelProps {
  ticket: Ticket;
  spaceKey: string;
  workspaceSlug: string;
  statuses: TicketStatus[];
  types: TicketType[];
  principals: KbGrantablePrincipal[];
  parentTicket: Ticket | undefined;
  canEdit: boolean;
  busy: boolean;
  allLabels: Label[];
  onUpdate: (ticketId: string, input: UpdateTicketInput) => Promise<Ticket>;
  onChangeStatus: (statusId: string) => Promise<void>;
  onAssign: (principalId: string) => Promise<void>;
  onUnassign: () => Promise<void>;
  onArchive: () => Promise<void>;
  onRestore: () => Promise<void>;
  onToggleLabel: (label: Label) => void;
  onCreateLabel: (name: string, color: string) => Promise<Label>;
  onChangeParent: (parentId: string | null) => Promise<void>;
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
  workspaceSlug,
  statuses,
  types,
  principals,
  parentTicket,
  canEdit,
  busy,
  allLabels,
  onUpdate,
  onChangeStatus,
  onAssign,
  onUnassign,
  onArchive,
  onRestore,
  onToggleLabel,
  onCreateLabel,
  onChangeParent,
}: TicketDetailPanelProps) {
  const type = types.find((t) => t.id === ticket.typeId);
  const archived = ticket.archivedAt !== null;

  const editor = useTicketEditor(ticket, canEdit && !archived, (input) => onUpdate(ticket.id, input));
  const docValue = isRichDoc(editor.doc) ? editor.doc : emptyRichDoc();

  return (
    // スクロールは器（SecondaryPanel の中身ラッパー）が持つ。ここに overflow-y-auto を
    // 付けると「スクロール範囲ゼロの空の容器」になり、overscroll-contain と相まって
    // ホイール操作を飲み込んで器までスクロールが届かなくなる（実測で確認）。
    // flex-1 / min-h-0 も親が flex コンテナではないため効かない。素の中身として置く。
    <div className="px-3 py-3" tabIndex={0}>
      <div className="mb-1.5 flex items-center gap-2">
        <span className="rounded bg-surface-2 px-1.5 py-0.5 text-[11px] font-semibold text-[var(--color-text-secondary)]">
          {type?.name ?? ''}
        </span>
        <TicketKeyBadge spaceKey={spaceKey} number={ticket.number} />
      </div>

      <TicketLabelBar
        attached={ticket.labels}
        allLabels={allLabels}
        canEdit={canEdit && !archived}
        onToggle={onToggleLabel}
        onCreate={onCreateLabel}
      />

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

      <TicketSection title="本文" action={<SaveStatusIndicator status={editor.saveStatus} />}>
        <Suspense fallback={<Loading />}>
          <RichTextEditor
            value={docValue}
            editable={canEdit && !archived}
            onChange={editor.changeDoc}
            ariaLabel="チケットの本文"
            placeholder="本文を書く"
            className="rte-compact"
          />
        </Suspense>
      </TicketSection>

      <TicketAttributePanel
        ticket={ticket}
        workspaceSlug={workspaceSlug}
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
        onChangeParent={(parentId) => void onChangeParent(parentId)}
      />

      <TicketSection title="子">
        <TicketChildrenSection workspaceSlug={workspaceSlug} ticketId={ticket.id} spaceKey={spaceKey} statuses={statuses} />
      </TicketSection>

      <TicketSection title="添付">
        <TicketAttachmentSection workspaceSlug={workspaceSlug} ticketId={ticket.id} canEdit={canEdit && !archived} />
      </TicketSection>

      <TicketSection title="コメント">
        <TicketCommentSection workspaceSlug={workspaceSlug} ticketId={ticket.id} compact />
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
