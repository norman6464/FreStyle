import { Link } from 'react-router-dom';
import { AuthLayout } from '@/widgets/auth-layout';
import Button from '@/shared/ui/Button';
import LinkText from '@/shared/ui/LinkText';
import Loading from '@/shared/ui/Loading';
import { useAcceptInvitation } from '@/hooks/useAcceptInvitation';

/**
 * 招待マジックリンク受諾画面。
 *
 * URL: /invitations/accept?token=<UUID>
 *
 * メールから踏まれた直後の最初の画面。token を backend (`GET /invitations/accept/:token`)
 * で検証して招待先 company / displayName を表示し、ユーザーに「ログインへ進む」
 * ボタンを押してもらう。ボタンを踏んだ時点で token は sessionStorage に保存済みなので、
 * ログイン後 callback で読み出して使う。
 *
 * UI で role の生値（company_admin / trainee 等）は表示しない。
 * 一般ユーザーには社内ロール体系の名称が伝わらないため、ログイン後の機能の出し分けで
 * 暗黙的に体感してもらう設計。
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

        <dl className="bg-surface-2 rounded-lg p-4 space-y-2 text-sm">
          {invitation.companyName && (
            <div className="flex justify-between gap-4">
              <dt className="text-[var(--color-text-muted)]">招待元の会社</dt>
              <dd className="font-medium text-right">{invitation.companyName}</dd>
            </div>
          )}
          {invitation.displayName && (
            <div className="flex justify-between gap-4">
              <dt className="text-[var(--color-text-muted)]">表示名</dt>
              <dd className="font-medium text-right">{invitation.displayName}</dd>
            </div>
          )}
        </dl>

        <Link to="/login" className="block">
          <Button variant="primary" fullWidth type="button">ログインへ進む</Button>
        </Link>

        <p className="text-xs text-[var(--color-text-muted)] text-center">
          ボタンを押すとログイン画面に進みます。<br />
          Google アカウント、またはメールアドレス + パスワードでログインしてください。
        </p>
      </div>
    </AuthLayout>
  );
}
