import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import LinkText from '../components/LinkText';
import FormMessage from '../components/FormMessage';
import { useConfirmSignup } from '../hooks/useConfirmSignup';

export default function ConfirmPage() {
  const { form, message, handleChange, handleConfirm } = useConfirmSignup();

  return (
    <AuthLayout>
      <FormMessage message={message} />
      <h2 className="text-2xl font-bold mb-6 text-center">確認コードの入力</h2>
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
        <PrimaryButton type="submit">確認する</PrimaryButton>
      </form>
      <div className="mt-4 text-center">
        <LinkText to="/signup">アカウント作成に戻る</LinkText>
      </div>
    </AuthLayout>
  );
}
