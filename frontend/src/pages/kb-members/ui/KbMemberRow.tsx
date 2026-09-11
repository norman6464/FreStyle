import { ArrowPathIcon, NoSymbolIcon, UserMinusIcon } from '@heroicons/react/24/outline';
import type { KbAdminWorkspaceMember, KbGrantRole } from '@/entities/kb';
import Avatar from '@/shared/ui/Avatar';

const ROLE_OPTIONS: { value: KbGrantRole; label: string }[] = [
  { value: 'admin', label: 'admin' },
  { value: 'editor', label: 'editor' },
  { value: 'commenter', label: 'commenter' },
  { value: 'viewer', label: 'viewer' },
];

export interface KbMemberRowProps {
  member: KbAdminWorkspaceMember;
  /** このメンバーが操作している本人か。自分自身には停止・削除の入口を出さない。 */
  isSelf: boolean;
  /** このメンバー宛ての操作が飛んでいる間 true（役割変更・停止・復帰・削除のどれか）。 */
  busy: boolean;
  onChangeRole: (role: KbGrantRole | null) => void;
  onSuspend: () => void;
  onRestore: () => void;
  onRemove: () => void;
}

/** メンバー管理画面（段 7）の 1 行。役割はその場の select で変える（保存ボタンを挟まない）。 */
export default function KbMemberRow({
  member,
  isSelf,
  busy,
  onChangeRole,
  onSuspend,
  onRestore,
  onRemove,
}: KbMemberRowProps) {
  const suspended = member.accountStatus === 'suspended';

  return (
    // 停止中は行全体を opacity で薄めない — 文字色との掛け合わせでコントラスト比が基準を
    // 割り込む（実測: 4.5:1 必要なところ 2.5 前後まで落ちる）。「状態」列のバッジだけで示す。
    <tr className="border-b border-surface-2 last:border-b-0">
      <td className="px-4 py-3">
        <div className="flex min-w-0 items-center gap-2.5">
          <Avatar name={member.name || '?'} src={member.avatarUrl || undefined} size="sm" />
          <div className="min-w-0">
            <div className="flex items-center gap-1.5 truncate text-sm font-semibold text-[var(--color-text-primary)]">
              <span className="truncate">{member.name || '（名前未設定）'}</span>
              {isSelf && <span className="text-xs font-normal text-[var(--color-text-muted)]">自分</span>}
            </div>
            {member.statusMessage && (
              <div className="truncate text-xs text-[var(--color-text-muted)]">{member.statusMessage}</div>
            )}
          </div>
        </div>
      </td>
      <td className="px-4 py-3">
        {suspended ? (
          <span className="text-sm text-[var(--color-text-muted)]">{member.role ?? '役割なし'}</span>
        ) : (
          <select
            aria-label={`${member.name || '相手'} の役割`}
            value={member.role ?? ''}
            disabled={busy}
            onChange={(e) => onChangeRole(e.target.value === '' ? null : (e.target.value as KbGrantRole))}
            className="rounded-md border border-surface-3 bg-surface-1 px-2 py-1 text-sm font-medium text-[var(--color-text-secondary)] disabled:opacity-50"
          >
            <option value="">役割なし</option>
            {ROLE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        )}
      </td>
      <td className="px-4 py-3">
        <span
          className={`inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-semibold ${
            suspended
              ? 'bg-amber-50 text-amber-700'
              : 'bg-surface-2 text-[var(--color-text-tertiary)]'
          }`}
        >
          <span className={`h-1.5 w-1.5 rounded-full ${suspended ? 'bg-amber-500' : 'bg-green-600'}`} />
          {suspended ? '停止中' : '有効'}
        </span>
      </td>
      <td className="px-4 py-3">
        {isSelf ? (
          <span className="text-xs text-[var(--color-text-muted)]">自分自身は操作できません</span>
        ) : (
          <div className="flex justify-end gap-1">
            {suspended ? (
              <button
                type="button"
                onClick={onRestore}
                disabled={busy}
                aria-label={`${member.name || '相手'} を復帰させる`}
                title="復帰させる"
                className="rounded-md p-1.5 text-[var(--color-text-muted)] hover:bg-brand-50 hover:text-brand-600 disabled:opacity-50"
              >
                <ArrowPathIcon className="h-4 w-4" aria-hidden="true" />
              </button>
            ) : (
              <button
                type="button"
                onClick={onSuspend}
                disabled={busy}
                aria-label={`${member.name || '相手'} を停止する`}
                title="アカウントを停止する"
                className="rounded-md p-1.5 text-[var(--color-text-muted)] hover:bg-amber-50 hover:text-amber-700 disabled:opacity-50"
              >
                <NoSymbolIcon className="h-4 w-4" aria-hidden="true" />
              </button>
            )}
            <button
              type="button"
              onClick={onRemove}
              disabled={busy}
              aria-label={`${member.name || '相手'} をワークスペースから外す`}
              title="ワークスペースから外す"
              className="rounded-md p-1.5 text-[var(--color-text-muted)] hover:bg-red-50 hover:text-red-600 disabled:opacity-50"
            >
              <UserMinusIcon className="h-4 w-4" aria-hidden="true" />
            </button>
          </div>
        )}
      </td>
    </tr>
  );
}
