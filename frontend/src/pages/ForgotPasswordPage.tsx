import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import { useForgotPassword } from '../hooks/useForgotPassword';

export default function ForgotPasswordPage() {
  const { email, setEmail, message, handleSubmit } = useForgotPassword();

  return (
    <AuthLayout>
      {message?.type === 'error' && (
        <p className="text-rose-600 text-center">{message.text}</p>
      )}
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
        />
        <PrimaryButton type="submit">確認コードを送信</PrimaryButton>
      </form>
    </AuthLayout>
  );
}
