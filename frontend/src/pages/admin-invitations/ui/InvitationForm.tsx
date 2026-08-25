import type { CreateInvitationForm, InvitationMethod } from '@/entities/invitation';
import type { Company } from '@/entities/company';

interface InvitationFormProps {
  form: CreateInvitationForm;
  onChange: (form: CreateInvitationForm) => void;
  companies: Company[];
  isSuperAdmin: boolean;
  isCompanyAdmin: boolean;
  method: InvitationMethod;
  onMethodChange: (method: InvitationMethod) => void;
  submitting: boolean;
  onSubmit: () => void;
}

/** 新規招待フォーム。会社・役職は認可境界（SoD）に応じて固定表示に切り替わる。 */
export default function InvitationForm({
  form,
  onChange,
  companies,
  isSuperAdmin,
  isCompanyAdmin,
  method,
  onMethodChange,
  submitting,
  onSubmit,
}: InvitationFormProps) {
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit();
      }}
      className="space-y-3 p-4 border rounded-lg bg-[var(--color-surface-1)]"
    >
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
            onChange={(e) => onChange({ ...form, companyId: Number(e.target.value) })}
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
          onChange={(e) => onChange({ ...form, email: e.target.value })}
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
      <label className="block text-sm">
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
      </label>

      <label className="block text-sm">
        <span className="block mb-1">表示名（任意）</span>
        <input
          maxLength={100}
          value={form.displayName ?? ''}
          onChange={(e) => onChange({ ...form, displayName: e.target.value })}
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
            onChange={() => onMethodChange('magic_link')}
          />
          <span>招待リンクをメール送信（本人がパスワードを設定）</span>
        </label>
        <label className="flex items-center gap-2 py-0.5">
          <input
            type="radio"
            name="method"
            value="temporary_password"
            checked={method === 'temporary_password'}
            onChange={() => onMethodChange('temporary_password')}
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
  );
}
