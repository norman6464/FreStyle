import { useSelector } from 'react-redux';
import { Navigate } from 'react-router-dom';
import type { RootState } from '../store';
import Loading from '../components/Loading';
import PageIntro from '../components/ui/PageIntro';
import { useAuditLog } from '../hooks/useAuditLog';
import type { AuditEvent } from '../repositories/AuditRepository';

// actor の role を日本語表記に。
function roleLabel(role: string): string {
  switch (role) {
    case 'super_admin':
      return '運営管理者';
    case 'company_admin':
      return '会社管理者';
    case 'trainee':
      return '受講者';
    default:
      return role;
  }
}

// 「METHOD ルートパターン」を日本語の操作名に。prefix（/api/v2）に依存しないよう部分一致で判定。
function actionLabel(action: string): string {
  const a = action.toUpperCase();
  if (a.includes('/COMPANIES/') && a.includes('ACTIVE')) return '会社の有効/無効を変更';
  if (a.includes('/MEMBERS/') && a.includes('ACTIVE')) return '従業員の有効/無効を変更';
  if (a.startsWith('DELETE') && a.includes('/MEMBERS/')) return '従業員を削除';
  if (a.startsWith('POST') && a.includes('/INVITATIONS')) return '招待を作成';
  if (a.startsWith('DELETE') && a.includes('/INVITATIONS/')) return '招待を取消';
  if (a.includes('/COMPANY-APPLICATIONS/') && a.includes('STATUS')) return '利用申請を承認/却下';
  return action;
}

function AuditRow({ e }: { e: AuditEvent }) {
  return (
    <li className="p-3 border border-surface-3 rounded-lg bg-[var(--color-surface-1)] space-y-1">
      <div className="flex items-center justify-between gap-3">
        <span className="text-sm font-medium text-[var(--color-text-primary)]">{actionLabel(e.action)}</span>
        <span className="text-xs text-[var(--color-text-muted)] flex-shrink-0">
          {new Date(e.createdAt).toLocaleString('ja-JP')}
        </span>
      </div>
      <p className="text-xs text-[var(--color-text-muted)]">
        実行者: {e.actorEmail || '(不明)'}（{roleLabel(e.actorRole)}）
        {e.targetId > 0 && <> ・ 対象 ID: {e.targetId}</>}
      </p>
    </li>
  );
}

/**
 * AdminAuditLogPage — `/admin/audit`。super_admin 専用の監査ログ。
 * 管理者の重要操作（会社の有効/無効・従業員の停止/削除・招待）を新しい順で確認できる。
 */
export default function AdminAuditLogPage() {
  const authLoading = useSelector((state: RootState) => state.auth.loading);
  const role = useSelector((state: RootState) => state.auth.role);
  const isSuperAdmin = role === 'super_admin';
  const { events, loading, error } = useAuditLog(isSuperAdmin);

  if (authLoading) return <Loading message="認証情報を確認中..." className="min-h-[50vh]" />;
  // 監査ログは全テナント横断の運営機能なので super_admin 専用。
  if (!isSuperAdmin) return <Navigate to="/" replace />;

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-3xl mx-auto space-y-6">
      <PageIntro
        title="監査ログ"
        description="管理者の重要操作（会社の有効/無効・従業員の停止/削除・招待の作成/取消）の記録です。新しい順に最大 200 件を表示します。"
      />

      {error && (
        <div role="alert" className="p-3 rounded border border-rose-300 bg-rose-50 text-rose-800 text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <Loading message="読み込み中..." className="min-h-[30vh]" />
      ) : events.length === 0 ? (
        <p className="text-sm text-[var(--color-text-muted)]">監査ログはまだありません。</p>
      ) : (
        <ul className="space-y-2">
          {events.map((e) => (
            <AuditRow key={e.id} e={e} />
          ))}
        </ul>
      )}
    </div>
  );
}
