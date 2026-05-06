import { Link } from 'react-router-dom';
import AuthLayout from '../components/AuthLayout';
import PrimaryButton from '../components/PrimaryButton';
import LinkText from '../components/LinkText';
import Loading from '../components/Loading';
import { useAcceptInvitation } from '../hooks/useAcceptInvitation';

/**
 * 招待マジックリンク受諾画面。
 *
 * URL: /invitations/accept?token=<UUID>
 *
 * メールから踏まれた直後の最初の画面。token を backend (`GET /invitations/accept/:token`)
 * で検証して招待先 company / role / displayName を表示し、ユーザーに「ログインへ進む」
 * ボタンを押してもらう。ボタンを踏んだ時点で token は sessionStorage に保存済みなので、
 * Cognito Hosted UI 経由のログイン後 callback で読み出して使う。
 */
export default function AcceptInvitationPage() {
  const { status } = useAcceptInvitation();

  if (status.kind === 'loading') {
    return <Loading fullscreen message="招待リンクを確認しています…" />;
  }

  if (status.kind === 'invalid') {
    return (
      <AuthLayout
        title="招待リンクが無効です"
        footer={
          <p className="text-sm text-[var(--color-text-muted)]">
            すでにログイン済みの方は <LinkText to="/login">ログイン画面</LinkText> へ
          </p>
        }
      >
        <p className="text-[var(--color-text-muted)] leading-relaxed">
          リンクが期限切れか、すでに使用済みの可能性があります。<br />
          招待を送ってくれた管理者に再送をお願いしてください。
        </p>
      </AuthLayout>
    );
  }

  if (status.kind === 'error') {
    return (
      <AuthLayout title="リンクを確認できませんでした">
        <p className="text-[var(--color-text-muted)] leading-relaxed">
          通信に失敗しました。時間を置いてからリンクを開き直してください。
        </p>
      </AuthLayout>
    );
  }

  const { invitation } = status;
  const roleLabel = roleLabels[invitation.role] ?? invitation.role;

  return (
    <AuthLayout
      title="FreStyle へようこそ"
      footer={
        <p className="text-sm text-[var(--color-text-muted)]">
          すでにログイン済みの方は <LinkText to="/login">こちら</LinkText>
        </p>
      }
    >
      <div className="space-y-4">
        <p className="text-[var(--color-text-muted)] leading-relaxed">
          管理者から招待が届いています。下記の内容でアカウントを作成します。
        </p>

        <dl className="bg-[var(--color-surface-alt)] rounded-lg p-4 space-y-2 text-sm">
          {invitation.companyName && (
            <div className="flex justify-between gap-4">
              <dt className="text-[var(--color-text-muted)]">招待元の会社</dt>
              <dd className="font-medium text-right">{invitation.companyName}</dd>
            </div>
          )}
          <div className="flex justify-between gap-4">
            <dt className="text-[var(--color-text-muted)]">付与される権限</dt>
            <dd className="font-medium text-right">{roleLabel}</dd>
          </div>
          {invitation.displayName && (
            <div className="flex justify-between gap-4">
              <dt className="text-[var(--color-text-muted)]">表示名</dt>
              <dd className="font-medium text-right">{invitation.displayName}</dd>
            </div>
          )}
        </dl>

        <Link to="/login" className="block">
          <PrimaryButton type="button">ログインへ進む</PrimaryButton>
        </Link>

        <p className="text-xs text-[var(--color-text-muted)] text-center">
          ボタンを押すと Cognito ログイン画面に遷移します。<br />
          Google アカウントまたはメールアドレスでログインしてください。
        </p>
      </div>
    </AuthLayout>
  );
}

// バックエンドから返ってくる role の生値を、画面表示用の日本語ラベルに変換する。
// 想定外の値は usecase 側で trainee に正規化済みだが、フォールバックとして生値をそのまま出す。
const roleLabels: Record<string, string> = {
  super_admin: '運営管理者',
  company_admin: '会社管理者',
  trainee: '受講者',
};
