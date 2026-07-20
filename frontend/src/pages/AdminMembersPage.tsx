import { useState, useMemo } from 'react';
import { useSelector } from 'react-redux';
import { Navigate } from 'react-router-dom';
import type { RootState } from '../store';
import Loading from '../components/Loading';
import PageIntro from '@/shared/ui/PageIntro';
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
  const { members, loading, error, updatingId, setAiAccess, setActive, remove } = useAdminMembers();
  const [query, setQuery] = useState('');

  // 氏名・メールアドレス・役割で部分一致（大文字小文字を無視）フィルタ。
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return members;
    return members.filter(
      (m) =>
        (m.displayName || '').toLowerCase().includes(q) ||
        m.email.toLowerCase().includes(q) ||
        roleLabel(m.role).toLowerCase().includes(q),
    );
  }, [members, query]);

  if (authLoading) return <Loading message="読み込み中..." className="min-h-[50vh]" />;
  if (!isAdmin) return <Navigate to="/" replace />;

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-4xl mx-auto">
      <PageIntro
        title="管理: 従業員一覧"
        description="自社の従業員の一覧です。AI 利用可否の個別設定に加え、アカウントの有効/無効（停止）と削除ができます。無効化された従業員はログイン/利用不可になります。"
      />

      {error && (
        <p role="alert" className="mt-4 text-rose-600">
          {error}
        </p>
      )}

      {loading ? (
        <Loading message="読み込み中..." className="min-h-[30vh]" />
      ) : members.length === 0 ? (
        <p className="mt-6 text-[var(--color-text-muted)]">従業員がまだいません。</p>
      ) : (
        <>
          <div className="mt-6">
            <input
              type="search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="氏名・メールアドレス・役割で検索"
              aria-label="従業員を検索"
              className="w-full sm:max-w-xs px-3 py-2 rounded-md bg-surface-2 border border-surface-3 text-sm text-[var(--color-text-primary)] focus:outline-none focus:border-brand-500"
            />
          </div>
          {filtered.length === 0 ? (
            <p className="mt-4 text-[var(--color-text-muted)]">「{query}」に一致する従業員がいません。</p>
          ) : (
            <div className="mt-4 overflow-x-auto rounded-lg border border-surface-3">
              <table className="w-full text-sm">
            <thead>
              <tr className="bg-surface-2 text-left text-[var(--color-text-muted)]">
                <th className="px-4 py-2 font-medium whitespace-nowrap">氏名</th>
                <th className="px-4 py-2 font-medium whitespace-nowrap">メールアドレス</th>
                <th className="px-4 py-2 font-medium whitespace-nowrap">役割</th>
                <th className="px-4 py-2 font-medium whitespace-nowrap">AI 利用</th>
                <th className="px-4 py-2 font-medium whitespace-nowrap">状態</th>
                <th className="px-4 py-2 font-medium whitespace-nowrap">操作</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((m) => (
                <tr key={m.id} className="border-t border-surface-3">
                  <td className="px-4 py-2 text-[var(--color-text-primary)]">{m.displayName || '—'}</td>
                  <td className="px-4 py-2 text-[var(--color-text-muted)]">{m.email}</td>
                  <td className="px-4 py-2 text-[var(--color-text-muted)] whitespace-nowrap">{roleLabel(m.role)}</td>
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
                  <td className="px-4 py-2 whitespace-nowrap">
                    {m.isActive ? (
                      <span className="text-emerald-600">有効</span>
                    ) : (
                      <span className="text-rose-600 font-medium">無効</span>
                    )}
                  </td>
                  <td className="px-4 py-2 whitespace-nowrap">
                    <div className="flex gap-2">
                      <button
                        type="button"
                        onClick={() => setActive(m.id, !m.isActive)}
                        disabled={updatingId === m.id}
                        className={`text-xs px-2.5 py-1 rounded border whitespace-nowrap transition-colors disabled:opacity-50 ${
                          m.isActive
                            ? 'border-rose-300 text-rose-700 hover:bg-rose-50'
                            : 'border-emerald-300 text-emerald-700 hover:bg-emerald-50'
                        }`}
                      >
                        {m.isActive ? '無効化' : '有効化'}
                      </button>
                      <button
                        type="button"
                        onClick={() => {
                          if (window.confirm(`${m.displayName || m.email} を削除します。よろしいですか？`)) {
                            remove(m.id);
                          }
                        }}
                        disabled={updatingId === m.id}
                        className="text-xs px-2.5 py-1 rounded border border-surface-3 text-[var(--color-text-muted)] whitespace-nowrap hover:bg-surface-2 transition-colors disabled:opacity-50"
                      >
                        削除
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  );
}
