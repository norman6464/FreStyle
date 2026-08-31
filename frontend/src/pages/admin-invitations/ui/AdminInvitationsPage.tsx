import FormMessage from '@/shared/ui/FormMessage';
import PageIntro from '@/shared/ui/PageIntro';
import ConfirmModal from '@/shared/ui/ConfirmModal';

import { useAdminInvitations } from '../model/useAdminInvitations';
import InvitationForm from './InvitationForm';
import InvitationList from './InvitationList';

// 通過条件（管理者かどうか）はルート側の RequireRole が持つ。
// 通信と状態は model/useAdminInvitations にあり、このファイルは組み立てだけを行う。
export default function AdminInvitationsPage() {
  const {
    invitations,
    loading,
    error,
    success,
    form,
    setForm,
    submitting,
    submit,
    method,
    setMethod,
    cancelTarget,
    canceling,
    requestCancel,
    closeCancelModal,
    confirmCancel,
    issuedPassword,
    copied,
    copyIssuedPassword,
    dismissIssuedPassword,
  } = useAdminInvitations();

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

      <InvitationForm
        form={form}
        onChange={setForm}
        method={method}
        onMethodChange={setMethod}
        submitting={submitting}
        onSubmit={submit}
      />

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
              onClick={copyIssuedPassword}
              className="px-3 py-2 rounded border bg-white text-sm whitespace-nowrap"
            >
              {copied ? 'コピーしました' : 'コピー'}
            </button>
          </div>
          <button
            type="button"
            onClick={dismissIssuedPassword}
            className="text-sm underline"
          >
            閉じる（表示を消す）
          </button>
        </div>
      )}

      <InvitationList invitations={invitations} loading={loading} onRequestCancel={requestCancel} />

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
