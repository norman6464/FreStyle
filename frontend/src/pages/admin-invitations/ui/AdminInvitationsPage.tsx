import { useEffect, useState, useCallback } from 'react';

import { AdminInvitationRepository, AdminInvitation,
  CreateInvitationForm } from '@/entities/invitation';
import type { InvitationMethod } from '@/entities/invitation';
import { CompanyRepository, Company } from '@/entities/company';
import { AuthRepository, UserInfo } from '@/entities/user';

import Loading from '@/shared/ui/Loading';
import FormMessage from '@/shared/ui/FormMessage';
import PageIntro from '@/shared/ui/PageIntro';
import ConfirmModal from '@/shared/ui/ConfirmModal';
import { logger } from '@/shared/lib/logger';

const EMPTY_FORM: CreateInvitationForm = {
  companyId: 0,
  email: '',
  role: 'trainee',
  displayName: '',
};

// バックエンドが英語のエラーコードをそのまま返した場合のフォールバック日本語化。
// バックエンドでカテゴリ化済みの日本語メッセージ（message フィールド）はそのまま通す。
function translateInviteError(raw: string): string {
  if (raw.includes('super_admin_can_only_invite_company_admin')) {
    return '運営は会社管理者のみ招待できます。受講者の招待は会社管理者から行ってください。';
  }
  if (raw.includes('company_admin_can_only_invite_trainee')) {
    return '会社管理者が招待できるのは受講者のみです。';
  }
  if (raw.includes('UsernameExistsException') || raw.includes('User account already exists')) {
    return 'このメールアドレスはすでに登録済みです。再招待は不要です。';
  }
  if (raw.includes('InvalidParameterException')) {
    return '入力値が不正です。メールアドレス形式を確認してください。';
  }
  if (raw.includes('LimitExceededException') || raw.includes('TooManyRequestsException')) {
    return '招待リクエストが多すぎます。しばらく待ってから再試行してください。';
  }
  if (raw.includes('AccessDeniedException') || raw.includes('not authorized')) {
    return '権限エラー: バックエンドの IAM ロール設定を確認してください。';
  }
  return raw;
}

// 通過条件（管理者かどうか）はルート側の RequireRole が持つ。
export default function AdminInvitationsPage() {
  const [me, setMe] = useState<UserInfo | null>(null);
  const [invitations, setInvitations] = useState<AdminInvitation[]>([]);
  const [companies, setCompanies] = useState<Company[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState<CreateInvitationForm>(EMPTY_FORM);
  const [submitting, setSubmitting] = useState(false);
  const [success, setSuccess] = useState<string | null>(null);
  const [cancelTarget, setCancelTarget] = useState<AdminInvitation | null>(null);
  const [canceling, setCanceling] = useState(false);
  const [method, setMethod] = useState<InvitationMethod>('magic_link');
  // 初期パスワード方式で発行された一時パスワード（1 度だけ表示。閉じると二度と取得できない）。
  const [issuedPassword, setIssuedPassword] = useState<{ email: string; password: string } | null>(null);
  const [copied, setCopied] = useState(false);

  // 認可境界（SoD）に応じて招待 UI を切り替えるため、自分の role / companyId を取得する。
  // backend 側でも同じ境界を強制しているので、フロントは UX 改善目的（不可能な選択肢を見せない）。
  //   - super_admin → role=company_admin で固定 / company は任意選択
  //   - company_admin → role=trainee で固定 / company は自社固定
  const isSuperAdmin = me?.role === 'super_admin';
  const isCompanyAdmin = me?.role === 'company_admin';

  const fetchAll = useCallback(async () => {
    setLoading(true);
    try {
      // 会社一覧 API は super_admin 専用。company_admin で呼ぶと 403 になり、
      // Promise.all ごと reject して招待一覧まで巻き込むため、先に自分の role を
      // 確認してから super_admin のときだけ取得する（company_admin は自社固定で不要）。
      const user = await AuthRepository.getCurrentUser();
      const [data, cos] = await Promise.all([
        AdminInvitationRepository.list(),
        user.role === 'super_admin' ? CompanyRepository.list() : Promise.resolve<Company[]>([]),
      ]);
      setMe(user);
      setInvitations(data);
      setCompanies(cos);

      // 役割に応じてフォームの初期値を上書きする。
      const defaultRole: CreateInvitationForm['role'] =
        user.role === 'super_admin' ? 'company_admin' : 'trainee';
      const defaultCompanyId =
        user.role === 'company_admin' && user.companyId
          ? user.companyId
          : cos[0]?.id ?? 0;
      setForm((f) => ({
        ...f,
        role: defaultRole,
        companyId: f.companyId === 0 ? defaultCompanyId : f.companyId,
      }));
      setError(null);
    } catch (e) {
      setError('データの取得に失敗しました');
      logger.error(e);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

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
      if (method === 'temporary_password') {
        const { invitation, temporaryPassword } =
          await AdminInvitationRepository.createWithTemporaryPassword(form);
        // 一時パスワードは 1 度だけ表示する。保存も再取得もできない。
        setIssuedPassword({ email: invitation.email, password: temporaryPassword });
        setCopied(false);
        setSuccess(null);
      } else {
        const created = await AdminInvitationRepository.create(form);
        setSuccess(
          `${created.email} 宛に招待メールを送信しました。受信者にメール内のリンクを開いてもらい、画面の案内に従ってログインしてもらってください。`
        );
      }
      setForm((f) => ({ ...EMPTY_FORM, companyId: f.companyId }));
      await fetchAll();
    } catch (err: unknown) {
      const raw =
        (err as { response?: { data?: { message?: string; error?: string } } })?.response?.data
          ?.message ||
        (err as { response?: { data?: { message?: string; error?: string } } })?.response?.data
          ?.error ||
        '招待の作成に失敗しました';
      setError(translateInviteError(raw));
      logger.error(err);
    } finally {
      setSubmitting(false);
    }
  };

  const requestCancel = (inv: AdminInvitation) => {
    setError(null);
    setCancelTarget(inv);
  };

  const closeCancelModal = () => {
    if (canceling) return;
    setCancelTarget(null);
  };

  const confirmCancel = async () => {
    if (!cancelTarget) return;
    setCanceling(true);
    try {
      await AdminInvitationRepository.cancel(cancelTarget.id);
      setCancelTarget(null);
      await fetchAll();
    } catch (err) {
      setError('招待のキャンセルに失敗しました');
      logger.error(err);
    } finally {
      setCanceling(false);
    }
  };

  const formatDate = (iso: string) => new Date(iso).toLocaleString('ja-JP');

  return (
    <div className="px-6 pt-6 pb-24 max-w-5xl mx-auto space-y-6">
      <PageIntro
        title="管理: メンバー招待"
        description={
          <>
            メールアドレスを入力すると、招待メールが送信されます。
            受信者はメール内のリンクから FreStyle の受諾画面に進み、
            Google アカウントまたはメールアドレスでログインしてアカウントが作成されます。
          </>
        }
      />

      <FormMessage message={error ? { type: 'error', text: error } : null} />
      {success && (
        <div role="status" className="p-3 rounded border border-emerald-300 bg-emerald-50 text-emerald-800 text-sm">
          {success}
        </div>
      )}

      <form onSubmit={submit} className="space-y-3 p-4 border rounded-lg bg-[var(--color-surface-1)]">
        <h2 className="text-base font-bold">新規招待</h2>

        <label className="block text-sm">
          <span className="block mb-1">会社 *</span>
          {isCompanyAdmin ? (
            // CompanyAdmin は自社にしか招待を出せない仕様。会社一覧 API は呼ばない
            // （super_admin 専用）ため、固定文言で自社宛であることを示す。
            <input
              type="text"
              readOnly
              value="所属会社（自社に固定）"
              className="w-full border rounded px-2 py-1 bg-[var(--color-surface-2)] text-[var(--color-text-muted)]"
            />
          ) : (
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
          )}
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

        {/*
         * 役職は SoD ルールで自動決定する:
         *   - SuperAdmin が招待 → 会社管理者 (company_admin) のみ
         *   - CompanyAdmin が招待 → 受講者 (trainee) のみ
         * select で誤った選択肢を露出させると backend の 403 で弾かれて UX が悪いので、
         * 一律「役職は固定（変更不可）」と表示する。
         */}
        <div className="block text-sm">
          <span className="block mb-1">役職</span>
          <input
            type="text"
            readOnly
            value={
              isSuperAdmin
                ? '会社管理者（招待先の会社の管理者）'
                : '受講者（自社のメンバー）'
            }
            className="w-full border rounded px-2 py-1 bg-[var(--color-surface-2)] text-[var(--color-text-muted)]"
          />
        </div>

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

        <fieldset className="block text-sm">
          <legend className="mb-1">招待方式</legend>
          <label className="flex items-center gap-2 py-0.5">
            <input
              type="radio"
              name="method"
              value="magic_link"
              checked={method === 'magic_link'}
              onChange={() => setMethod('magic_link')}
            />
            <span>招待リンクをメール送信（本人がパスワードを設定）</span>
          </label>
          <label className="flex items-center gap-2 py-0.5">
            <input
              type="radio"
              name="method"
              value="temporary_password"
              checked={method === 'temporary_password'}
              onChange={() => setMethod('temporary_password')}
            />
            <span>初期パスワードを発行（その場で本人に渡す・初回ログインで変更）</span>
          </label>
        </fieldset>

        <button
          type="submit"
          disabled={submitting || form.companyId === 0}
          className="px-4 py-2 rounded bg-emerald-600 text-white disabled:opacity-50"
        >
          {submitting
            ? '送信中...'
            : method === 'temporary_password'
              ? '初期パスワードを発行'
              : '招待メールを送信'}
        </button>
      </form>

      {issuedPassword && (
        <div
          role="status"
          className="p-4 rounded-lg border border-amber-300 bg-amber-50 text-amber-900 space-y-2"
        >
          <p className="font-bold">初期パスワードを発行しました</p>
          <p className="text-sm">
            {issuedPassword.email} の初期パスワードです。<strong>この画面を閉じると二度と表示できません。</strong>
            本人に安全に渡してください。初回ログイン時に本人が新しいパスワードへ変更します。
          </p>
          <div className="flex items-center gap-2">
            <code className="flex-1 px-3 py-2 rounded bg-white border font-mono text-base break-all">
              {issuedPassword.password}
            </code>
            <button
              type="button"
              onClick={async () => {
                try {
                  await navigator.clipboard.writeText(issuedPassword.password);
                  setCopied(true);
                } catch {
                  setCopied(false);
                }
              }}
              className="px-3 py-2 rounded border bg-white text-sm whitespace-nowrap"
            >
              {copied ? 'コピーしました' : 'コピー'}
            </button>
          </div>
          <button
            type="button"
            onClick={() => setIssuedPassword(null)}
            className="text-sm underline"
          >
            閉じる（表示を消す）
          </button>
        </div>
      )}

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
                  onClick={() => requestCancel(inv)}
                  className="text-xs px-2 py-1 border border-red-300 rounded text-red-700 hover:bg-red-50 transition-colors"
                >
                  取り消し
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>

      <ConfirmModal
        isOpen={cancelTarget !== null}
        title="招待を取り消し"
        message={
          cancelTarget
            ? `${cancelTarget.email} 宛の招待を取り消します。受信者は招待リンクから登録できなくなります。`
            : ''
        }
        confirmText={canceling ? '処理中...' : '取り消す'}
        cancelText="戻る"
        onConfirm={confirmCancel}
        onCancel={closeCancelModal}
        isDanger={true}
      />
    </div>
  );
}
