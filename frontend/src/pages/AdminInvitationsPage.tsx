import { useEffect, useState, useCallback } from 'react';
import { useSelector } from 'react-redux';
import { Navigate } from 'react-router-dom';
import AdminInvitationRepository, {
  AdminInvitation,
  CreateInvitationForm,
} from '../repositories/AdminInvitationRepository';
import CompanyRepository, { Company } from '../repositories/CompanyRepository';
import type { RootState } from '../store';
import Loading from '../components/Loading';
import PageIntro from '../components/ui/PageIntro';
import { logger } from '../lib/logger';

const EMPTY_FORM: CreateInvitationForm = {
  companyId: 0,
  email: '',
  role: 'trainee',
  displayName: '',
};

export default function AdminInvitationsPage() {
  const isAdmin = useSelector((state: RootState) => state.auth.isAdmin);
  const authLoading = useSelector((state: RootState) => state.auth.loading);

  const [invitations, setInvitations] = useState<AdminInvitation[]>([]);
  const [companies, setCompanies] = useState<Company[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState<CreateInvitationForm>(EMPTY_FORM);
  const [submitting, setSubmitting] = useState(false);
  const [success, setSuccess] = useState<string | null>(null);

  const fetchAll = useCallback(async () => {
    setLoading(true);
    try {
      const [data, cos] = await Promise.all([
        AdminInvitationRepository.list(),
        CompanyRepository.list(),
      ]);
      setInvitations(data);
      setCompanies(cos);
      if (cos.length > 0 && form.companyId === 0) {
        setForm((f) => ({ ...f, companyId: cos[0].id }));
      }
      setError(null);
    } catch (e) {
      setError('データの取得に失敗しました');
      logger.error(e);
    } finally {
      setLoading(false);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (isAdmin) fetchAll();
  }, [isAdmin, fetchAll]);

  if (authLoading) return <Loading message="認証情報を確認中..." />;
  if (!isAdmin) return <Navigate to="/" replace />;

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.companyId) {
      setError('会社を選択してください');
      return;
    }
    setSubmitting(true);
    setError(null);
    setSuccess(null);
    try {
      const created = await AdminInvitationRepository.create(form);
      setSuccess(
        `${created.email} 宛に招待メールを送信しました。受信者は Cognito Hosted UI で初回パスワードを変更してログインしてください。`
      );
      setForm((f) => ({ ...EMPTY_FORM, companyId: f.companyId }));
      await fetchAll();
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string; error?: string } } })?.response?.data
          ?.message ||
        (err as { response?: { data?: { message?: string; error?: string } } })?.response?.data
          ?.error ||
        '招待の作成に失敗しました';
      setError(msg);
      logger.error(err);
    } finally {
      setSubmitting(false);
    }
  };

  const cancel = async (id: number) => {
    if (!confirm('この招待をキャンセルしますか？')) return;
    try {
      await AdminInvitationRepository.cancel(id);
      await fetchAll();
    } catch (err) {
      setError('招待のキャンセルに失敗しました');
      logger.error(err);
    }
  };

  const formatDate = (iso: string) => new Date(iso).toLocaleString('ja-JP');

  return (
    <div className="px-6 pt-6 pb-24 max-w-3xl mx-auto space-y-6">
      <PageIntro
        title="管理: メンバー招待"
        description={
          <>
            メールアドレスを入力すると、Cognito 経由で一時パスワード付きの招待メールが送信されます。
            受信者は初回ログイン時にパスワード変更を要求されます。
          </>
        }
      />

      {error && (
        <div role="alert" className="p-3 rounded border border-red-300 bg-red-50 text-red-800 text-sm">
          {error}
        </div>
      )}
      {success && (
        <div role="status" className="p-3 rounded border border-emerald-300 bg-emerald-50 text-emerald-800 text-sm">
          {success}
        </div>
      )}

      <form onSubmit={submit} className="space-y-3 p-4 border rounded-lg bg-[var(--color-surface-1)]">
        <h2 className="text-base font-bold">新規招待</h2>

        <label className="block text-sm">
          <span className="block mb-1">会社 *</span>
          <select
            required
            value={form.companyId}
            onChange={(e) => setForm({ ...form, companyId: Number(e.target.value) })}
            className="w-full border rounded px-2 py-1"
          >
            <option value={0} disabled>会社を選択してください</option>
            {companies.map((c) => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
        </label>

        <label className="block text-sm">
          <span className="block mb-1">メールアドレス *</span>
          <input
            required
            type="email"
            maxLength={254}
            value={form.email}
            onChange={(e) => setForm({ ...form, email: e.target.value })}
            placeholder="newmember@example.com"
            className="w-full border rounded px-2 py-1"
          />
        </label>

        <label className="block text-sm">
          <span className="block mb-1">ロール *</span>
          <select
            required
            value={form.role}
            onChange={(e) =>
              setForm({ ...form, role: e.target.value as CreateInvitationForm['role'] })
            }
            className="w-full border rounded px-2 py-1"
          >
            <option value="trainee">Trainee（新卒研修対象）</option>
            <option value="company_admin">CompanyAdmin（メンター）</option>
          </select>
        </label>

        <label className="block text-sm">
          <span className="block mb-1">表示名（任意）</span>
          <input
            maxLength={100}
            value={form.displayName ?? ''}
            onChange={(e) => setForm({ ...form, displayName: e.target.value })}
            placeholder="例: 山田太郎"
            className="w-full border rounded px-2 py-1"
          />
          <span className="block mt-1 text-xs text-[var(--color-text-muted)]">
            未入力の場合はメールアドレスのローカル部から自動生成されます。
          </span>
        </label>

        <button
          type="submit"
          disabled={submitting || form.companyId === 0}
          className="px-4 py-2 rounded bg-emerald-600 text-white disabled:opacity-50"
        >
          {submitting ? '送信中...' : '招待メールを送信'}
        </button>
      </form>

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
                    招待日: {formatDate(inv.createdAt)} / 有効期限: {formatDate(inv.expiresAt)}
                  </p>
                </div>
                <button
                  onClick={() => cancel(inv.id)}
                  className="text-xs px-2 py-1 border border-red-300 rounded text-red-700"
                >
                  取り消し
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
