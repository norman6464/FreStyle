import type { AdminInvitation } from '@/entities/invitation';
import Loading from '@/shared/ui/Loading';
import { formatDateTime } from '@/shared/lib/formatters';

interface InvitationListProps {
  invitations: AdminInvitation[];
  loading: boolean;
  onRequestCancel: (invitation: AdminInvitation) => void;
}

/** 未承諾の招待一覧。取り消しボタンは確認モーダルの表示要求だけを親へ返す。 */
export default function InvitationList({ invitations, loading, onRequestCancel }: InvitationListProps) {
  return (
    <section>
      <h2 className="text-base font-bold mb-3">未承諾の招待 ({invitations.length})</h2>
      {loading ? (
        <Loading message="読み込み中..." />
      ) : invitations.length === 0 ? (
        <p className="text-sm text-[var(--color-text-muted)]">未承諾の招待はありません</p>
      ) : (
        <ul className="space-y-2">
          {invitations.map((inv) => (
            <li
              key={inv.id}
              className="p-3 border rounded flex items-start justify-between gap-3 bg-[var(--color-surface-1)]"
            >
              <div className="flex-1 text-sm">
                <p className="font-bold">
                  {inv.email}{' '}
                  <span className="text-xs px-1.5 py-0.5 rounded bg-surface-3">{inv.role}</span>
                </p>
                <p className="text-xs text-[var(--color-text-muted)]">
                  招待日: {formatDateTime(inv.createdAt)} / 有効期限: {formatDateTime(inv.expiresAt)}
                </p>
              </div>
              <button
                onClick={() => onRequestCancel(inv)}
                className="text-xs px-2 py-1 border border-red-300 rounded text-red-700 hover:bg-red-50 transition-colors"
              >
                取り消し
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
