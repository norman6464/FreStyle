import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import FormMessage from '../components/FormMessage';
import { useForgotPassword } from '../hooks/useForgotPassword';

export default function ForgotPasswordPage() {
  const { email, setEmail, message, loading, handleSubmit, clearMessage } = useForgotPassword();

  return (
    <AuthLayout>
      <FormMessage message={message} onDismiss={clearMessage} />
      <h2 className="text-2xl font-bold mb-6 text-center">
        パスワードリセット
      </h2>
      <form onSubmit={handleSubmit}>
        <InputField
          label="メールアドレス"
          name="email"
          type="email"
          value={email}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEmail(e.target.value)}
          disabled={loading}
        />
        <PrimaryButton type="submit" loading={loading}>
          {loading ? '送信中...' : '確認コードを送信'}
        </PrimaryButton>
      </form>
    </AuthLayout>
  );
}
