import type { CreateInvitationForm, InvitationMethod } from '@/entities/invitation';

interface InvitationFormProps {
  form: CreateInvitationForm;
  onChange: (form: CreateInvitationForm) => void;
  method: InvitationMethod;
  onMethodChange: (method: InvitationMethod) => void;
  submitting: boolean;
  onSubmit: () => void;
}

/** 新規招待フォーム。招待先と役職はどちらも固定なので、選ばせずに表示だけする。 */
export default function InvitationForm({
  form,
  onChange,
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
        <span className="block mb-1">招待先</span>
        {/* 招待先は常に自分の所属ワークスペース。選ばせる余地が無いので入力にしない。 */}
        <input
          type="text"
          readOnly
          value="自分のワークスペース"
          className="w-full border rounded px-2 py-1 bg-[var(--color-surface-2)] text-[var(--color-text-muted)]"
        />
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
       * 招待できるのは受講者だけ。select で選ばせても backend の 403 で弾かれるので、
       * 固定であることをそのまま見せる。
       */}
      <label className="block text-sm">
        <span className="block mb-1">役職</span>
        <input
          type="text"
          readOnly
          value="受講者（自分のワークスペースのメンバー）"
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
        disabled={submitting}
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
