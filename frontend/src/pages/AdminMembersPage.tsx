import { useSelector } from 'react-redux';
import { Navigate } from 'react-router-dom';
import type { RootState } from '../store';
import Loading from '../components/Loading';
import PageIntro from '../components/ui/PageIntro';
import { useAdminMembers } from '../hooks/useAdminMembers';
import { Member } from '../repositories/AdminMemberRepository';

// AI 利用可否の 3 状態 ↔ select の値。
function aiValue(m: Member): 'inherit' | 'on' | 'off' {
  if (m.aiChatEnabled === null) return 'inherit';
  return m.aiChatEnabled ? 'on' : 'off';
}

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

/**
 * AdminMembersPage — `/admin/members`。company_admin / super_admin 向けの従業員一覧。
 * 各従業員の AI 利用可否を「会社設定に従う / 有効 / 無効」で個別に設定できる
 * （会社一括設定は従来どおり別途残る。個別設定が会社設定を上書きする）。
 */
export default function AdminMembersPage() {
  const isAdmin = useSelector((state: RootState) => state.auth.isAdmin);
  const authLoading = useSelector((state: RootState) => state.auth.loading);
  const { members, loading, error, updatingId, setAiAccess } = useAdminMembers();

  if (authLoading) return <Loading message="読み込み中..." className="min-h-[50vh]" />;
  if (!isAdmin) return <Navigate to="/" replace />;

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-4xl mx-auto">
      <PageIntro
        title="管理: 従業員一覧"
        description="自社の従業員の一覧です。各従業員の AI 利用可否を個別に設定できます（会社の一括設定を上書きします）。"
      />

      {loading ? (
        <Loading message="読み込み中..." className="min-h-[30vh]" />
      ) : error ? (
        <p className="mt-6 text-rose-600">{error}</p>
      ) : members.length === 0 ? (
        <p className="mt-6 text-[var(--color-text-muted)]">従業員がまだいません。</p>
      ) : (
        <div className="mt-6 overflow-x-auto rounded-lg border border-surface-3">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-surface-2 text-left text-[var(--color-text-muted)]">
                <th className="px-4 py-2 font-medium">氏名</th>
                <th className="px-4 py-2 font-medium">メールアドレス</th>
                <th className="px-4 py-2 font-medium">役割</th>
                <th className="px-4 py-2 font-medium">AI 利用</th>
              </tr>
            </thead>
            <tbody>
              {members.map((m) => (
                <tr key={m.id} className="border-t border-surface-3">
                  <td className="px-4 py-2 text-[var(--color-text-primary)]">{m.displayName || '—'}</td>
                  <td className="px-4 py-2 text-[var(--color-text-muted)]">{m.email}</td>
                  <td className="px-4 py-2 text-[var(--color-text-muted)]">{roleLabel(m.role)}</td>
                  <td className="px-4 py-2">
                    <select
                      aria-label={`${m.displayName || m.email} の AI 利用可否`}
                      value={aiValue(m)}
                      disabled={updatingId === m.id || m.role !== 'trainee'}
                      onChange={(e) => {
                        const v = e.target.value;
                        setAiAccess(m.id, v === 'inherit' ? null : v === 'on');
                      }}
                      className="px-2 py-1 rounded-md bg-surface-2 border border-surface-3 text-[var(--color-text-primary)] focus:outline-none focus:border-brand-500 disabled:opacity-50"
                    >
                      <option value="inherit">会社設定に従う</option>
                      <option value="on">有効（個別ON）</option>
                      <option value="off">無効（個別OFF）</option>
                    </select>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
