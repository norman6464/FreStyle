import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import { useConfirmForgotPassword } from '../hooks/useConfirmForgotPassword';

export default function ConfirmForgotPasswordPage() {
  const { form, message, handleChange, handleConfirm } = useConfirmForgotPassword();

  return (
    <AuthLayout>
      {message?.type === 'error' && (
        <p className="text-rose-400 text-center">{message.text}</p>
      )}
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
        />
        <InputField
          label="確認コード"
          name="code"
          value={form.code}
          onChange={handleChange}
        />
        <InputField
          label="新しいパスワード"
          name="newPassword"
          type="password"
          value={form.newPassword}
          onChange={handleChange}
        />
        <PrimaryButton type="submit">パスワードをリセット</PrimaryButton>
      </form>
    </AuthLayout>
  );
}
