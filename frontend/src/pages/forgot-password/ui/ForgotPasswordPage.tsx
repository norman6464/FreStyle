import { AuthLayout } from '@/widgets/auth-layout';
import InputField from '@/shared/ui/InputField';
import Button from '@/shared/ui/Button';
import FormMessage from '@/shared/ui/FormMessage';
import { useForgotPassword } from '../model/useForgotPassword';

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
        <Button variant="primary" fullWidth type="submit" loading={loading}>
          {loading ? '送信中...' : '確認コードを送信'}
        </Button>
      </form>
    </AuthLayout>
  );
}
