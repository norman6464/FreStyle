import { AuthLayout } from '@/widgets/auth-layout';
import PublicHeader from '@/shared/ui/PublicHeader';
import InputField from '@/shared/ui/InputField';
import TextareaField from '@/shared/ui/TextareaField';
import Button from '@/shared/ui/Button';
import LinkText from '@/shared/ui/LinkText';
import { useCompanyApplication } from '../model/useCompanyApplication';
import { XCircleIcon, CheckCircleIcon } from '@heroicons/react/24/outline';

/** CompanyApplicationPage は未登録の企業担当者がログイン前に出す利用申請フォーム。 */
export default function CompanyApplicationPage() {
  const { form, loading, submitted, error, handleChange, handleSubmit } =
    useCompanyApplication();

  return (
    <AuthLayout
      title="企業の利用申請"
      header={<PublicHeader />}
      footer={
        <p className="text-sm text-[var(--color-text-muted)]">
          すでにアカウントをお持ちの方{' '}
          <LinkText to="/login">ログイン</LinkText>
        </p>
      }
    >
      {submitted ? (
        <p
          role="status"
          className="flex items-center justify-center gap-1 text-emerald-700 text-center p-4 bg-emerald-50 border border-emerald-200 rounded-lg font-medium"
        >
          <CheckCircleIcon className="w-5 h-5" aria-hidden="true" />
          申請を受け付けました。担当者より追ってご連絡いたします。
        </p>
      ) : (
        <>
          <p className="text-sm text-[var(--color-text-muted)] mb-3">
            FreStyle の導入をご検討の企業さまはこちらからお申し込みください。運営が内容を確認し、ご登録手続きのご案内をいたします。
          </p>

          {error && (
            <p
              role="alert"
              className="flex items-center justify-center gap-1 text-rose-700 text-center mb-4 p-3 bg-rose-50 border border-rose-200 rounded-lg font-medium"
            >
              <XCircleIcon className="w-4 h-4" aria-hidden="true" />
              {error}
            </p>
          )}

          {/* このページは項目数が多く縦長になりやすいため、フィールド間余白(共有 InputField の
              mb-6)をこの画面だけ mb-4 に詰める。共有コンポーネントは他の認証ページでも使うため
              変更せず、ここでの子セレクタ上書きで完結させる。 */}
          <form
            onSubmit={handleSubmit}
            aria-label="企業利用申請フォーム"
            className="[&>div]:mb-4"
          >
            <InputField
              label="会社名"
              name="companyName"
              value={form.companyName}
              onChange={handleChange}
              disabled={loading}
              maxLength={200}
            />
            <InputField
              label="お名前（ご担当者）"
              name="applicantName"
              value={form.applicantName}
              onChange={handleChange}
              disabled={loading}
              maxLength={120}
            />
            <InputField
              label="メールアドレス"
              name="email"
              type="email"
              value={form.email}
              onChange={handleChange}
              disabled={loading}
              maxLength={255}
            />
            <TextareaField
              label="メッセージ（任意）"
              name="message"
              value={form.message}
              onChange={handleChange}
              rows={3}
              maxLength={2000}
              placeholder="ご利用人数や導入時期など、ご要望があればご記入ください。"
            />
            <Button variant="primary" fullWidth type="submit" loading={loading}>
              {loading ? '送信中...' : '申請する'}
            </Button>
          </form>
        </>
      )}
    </AuthLayout>
  );
}
