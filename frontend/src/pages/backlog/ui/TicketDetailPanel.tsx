import { lazy, Suspense } from 'react';
import {
  TicketKeyBadge,
  formatTicketKey,
  type Ticket,
  type TicketChangeGroup,
  type TicketPriority,
  type TicketStatus,
  type TicketType,
  type UpdateTicketInput,
} from '@/entities/ticket';
import type { KbGrantablePrincipal } from '@/entities/kb';
import Loading from '@/shared/ui/Loading';
import { SaveStatusIndicator, emptyRichDoc, isRichDoc } from '@/shared/ui/RichTextEditor';
import { useTicketEditor } from '../model/useTicketEditor';

const RichTextEditor = lazy(() => import('@/shared/ui/RichTextEditor').then((m) => ({ default: m.RichTextEditor })));

export interface TicketDetailPanelProps {
  ticket: Ticket;
  spaceKey: string;
  statuses: TicketStatus[];
  types: TicketType[];
  principals: KbGrantablePrincipal[];
  parentTicket: Ticket | undefined;
  history: TicketChangeGroup[];
  historyLoading: boolean;
  canEdit: boolean;
  busy: boolean;
  onUpdate: (ticketId: string, input: UpdateTicketInput) => Promise<Ticket>;
  onChangeStatus: (statusId: string) => Promise<void>;
  onAssign: (principalId: string) => Promise<void>;
  onUnassign: () => Promise<void>;
  onArchive: () => Promise<void>;
  onRestore: () => Promise<void>;
  onClose: () => void;
}

const PRIORITY_LABEL: Record<TicketPriority, string> = { 1: '高', 2: '中', 3: '低' };

const FIELD_LABEL: Record<string, string> = {
  title: '題名',
  doc: '本文',
  status: '状態',
  type: '種別',
  priority: '優先度',
  assignee: '担当',
  parent: '親',
  start_date: '開始日',
  due_date: '期限',
  resolution: '完了理由',
  position: '並び順',
  archived: 'アーカイブ',
  category: '枠',
  milestone: '節目',
  link: 'リンク',
};

function formatDateTime(iso: string): string {
  const d = new Date(iso);
  return `${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
}

/** チケット詳細パネル。設計 Ⅲ・Ⅳ-D（全置換 PUT は開いている行だけ編集できる）。 */
export default function TicketDetailPanel({
  ticket,
  spaceKey,
  statuses,
  types,
  principals,
  parentTicket,
  history,
  historyLoading,
  canEdit,
  busy,
  onUpdate,
  onChangeStatus,
  onAssign,
  onUnassign,
  onArchive,
  onRestore,
  onClose,
}: TicketDetailPanelProps) {
  const type = types.find((t) => t.id === ticket.typeId);
  const archived = ticket.archivedAt !== null;
  const assigneeUsers = principals.filter((p) => p.kind === 'user');

  const editor = useTicketEditor(ticket, canEdit && !archived, (input) => onUpdate(ticket.id, input));
  const docValue = isRichDoc(editor.doc) ? editor.doc : emptyRichDoc();

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex items-center gap-2 border-b border-surface-3 px-3 py-2">
        <span className="text-sm font-semibold text-[var(--color-text-secondary)]">詳細</span>
        <SaveStatusIndicator status={editor.saveStatus} />
        <button
          type="button"
          onClick={onClose}
          aria-label="詳細を閉じる"
          className="ml-auto rounded p-1 text-[var(--color-text-muted)] hover:bg-surface-2"
        >
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" aria-hidden="true">
            <path d="M18 6 6 18" />
            <path d="m6 6 12 12" />
          </svg>
        </button>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto px-3 py-3" tabIndex={0}>
        <div className="mb-1.5 flex items-center gap-2">
          <span className="rounded bg-surface-2 px-1.5 py-0.5 text-[11px] font-semibold text-[var(--color-text-secondary)]">
            {type?.name ?? ''}
          </span>
          <TicketKeyBadge spaceKey={spaceKey} number={ticket.number} />
        </div>

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

        <dl className="mb-4 grid grid-cols-[64px_1fr] gap-x-2 gap-y-2 text-sm">
          <dt className="text-[var(--color-text-muted)]">状態</dt>
          <dd>
            {canEdit ? (
              <select
                value={ticket.statusId}
                disabled={busy}
                onChange={(e) => void onChangeStatus(e.target.value)}
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
                  if (value) void onAssign(value);
                  else void onUnassign();
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
                value={editor.priority}
                onChange={(e) => editor.changePriority(Number(e.target.value) as TicketPriority)}
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
                value={editor.dueDate ?? ''}
                onChange={(e) => editor.changeDueDate(e.target.value || null)}
                aria-label="期限"
                className="rounded border border-surface-3 bg-surface-1 px-1.5 py-0.5 text-sm"
              />
            ) : (
              ticket.dueDate ?? <span className="text-[var(--color-text-muted)]">なし</span>
            )}
          </dd>

          <dt className="text-[var(--color-text-muted)]">親</dt>
          <dd>
            {parentTicket ? (
              formatTicketKey(spaceKey, parentTicket.number)
            ) : (
              <span className="text-[var(--color-text-muted)]">なし</span>
            )}
          </dd>
        </dl>

        <div className="mb-4">
          <div className="mb-1 text-[10.5px] font-semibold uppercase tracking-wide text-[var(--color-text-muted)]">
            本文
          </div>
          <Suspense fallback={<Loading />}>
            <RichTextEditor
              value={docValue}
              editable={canEdit && !archived}
              onChange={editor.changeDoc}
              ariaLabel="チケットの本文"
              placeholder="本文を書く"
            />
          </Suspense>
        </div>

        <div className="mb-4">
          <div className="mb-1 text-[10.5px] font-semibold uppercase tracking-wide text-[var(--color-text-muted)]">
            変更履歴
          </div>
          {historyLoading ? (
            <Loading />
          ) : history.length === 0 ? (
            <p className="text-xs text-[var(--color-text-muted)]">まだ変更はありません</p>
          ) : (
            <ul className="space-y-1.5">
              {history.flatMap((group) =>
                group.items.map((item) => (
                  <li key={item.id} className="flex items-baseline gap-2 text-xs">
                    <span className="flex-shrink-0 text-[var(--color-text-muted)]">{formatDateTime(group.createdAt)}</span>
                    <span className="text-[var(--color-text-secondary)]">
                      {FIELD_LABEL[item.field] ?? item.field}を
                      {item.oldLabel && <b> {item.oldLabel}</b>}
                      {item.oldLabel && ' から '}
                      <b> {item.newLabel ?? item.newValue ?? ''}</b> へ
                    </span>
                  </li>
                )),
              )}
            </ul>
          )}
        </div>

        {canEdit && (
          <div>
            <div className="mb-1 text-[10.5px] font-semibold uppercase tracking-wide text-[var(--color-text-muted)]">
              操作
            </div>
            <button
              type="button"
              onClick={() => void (archived ? onRestore() : onArchive())}
              disabled={busy}
              className="rounded border border-surface-3 px-2.5 py-1 text-xs font-medium text-[var(--color-text-secondary)] hover:bg-surface-2 disabled:opacity-50"
            >
              {archived ? '現役に戻す' : 'アーカイブ'}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
