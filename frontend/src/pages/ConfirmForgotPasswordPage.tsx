import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import FormMessage from '../components/FormMessage';
import { useConfirmForgotPassword } from '../hooks/useConfirmForgotPassword';

export default function ConfirmForgotPasswordPage() {
  const { form, message, loading, handleChange, handleConfirm } = useConfirmForgotPassword();

  return (
    <AuthLayout>
      <FormMessage message={message} />
      <h2 className="text-2xl font-bold mb-6 text-center">
        パスワードリセット確認
      </h2>
      <form onSubmit={handleConfirm}>
        <InputField
          label="メールアドレス"
          name="email"
          type="email"
          value={form.email}
          onChange={handleChange}
          disabled={loading}
        />
        <InputField
          label="確認コード"
          name="code"
          value={form.code}
          onChange={handleChange}
          disabled={loading}
        />
        <InputField
          label="新しいパスワード"
          name="newPassword"
          type="password"
          value={form.newPassword}
          onChange={handleChange}
          disabled={loading}
        />
        <PrimaryButton type="submit" loading={loading}>
          {loading ? 'リセット中...' : 'パスワードをリセット'}
        </PrimaryButton>
      </form>
    </AuthLayout>
  );
}
