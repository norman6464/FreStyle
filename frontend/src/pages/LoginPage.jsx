import { useState, useEffect } from 'react';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import SNSSignInButton from '../components/SNSSignInButton';
import LinkText from '../components/LinkText';
import { getCognitoAuthUrl } from '../utils/auth';
import { useLocation, useNavigate } from 'react-router-dom';

export default function LoginPage() {
  const [form, setForm] = useState({ email: '', password: '' });
  // フラッシュメッセージ
  const [loginMessage, setLoginMessage] = useState(null);

  // SignupPageで登録成功時のメッセージが表示される
  const location = useLocation();
  const navigate = useNavigate();
  const message = location.state?.message;
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const handleLogin = async (e) => {
    e.preventDefault();

    try {
      const response = await fetch(`${API_BASE_URL}/api/auth/cognito/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: form.email,
          password: form.password,
        }),
        credentials: 'include',
      });

      const data = await response.json();

      console.log('ステータスコード', response.status);

      if (response.ok) {
        dispatch(setAuthData());
        navigate('/');
      } else {
        setLoginMessage({
          type: 'error',
          text: data.error || 'ログインに失敗しました。',
        });
      }
    } catch (error) {
      setLoginMessage({ type: 'error', text: '通信エラーが発生しました。' });
    }
  };

  const handleChange = (e) => {
    setForm({
      ...form,
      [e.target.name]: e.target.value,
    });
  };

  return (
    <AuthLayout>
      {/* flash Message */}
      <div>
        {message && (
          <p className="text-green-600 text-center mb-4 p-3 bg-green-50 rounded-lg font-semibold animate-fade-in">
            ✓ {message}
          </p>
        )}
        {loginMessage?.type === 'error' && (
          <p className="text-red-600 text-center mb-4 p-3 bg-red-50 rounded-lg font-semibold animate-fade-in">
            ✕ {loginMessage.text}
          </p>
        )}
      </div>
      <h2 className="text-3xl font-bold mb-2 text-center text-gray-800">
        ログイン
      </h2>
      <p className="text-center text-gray-500 text-sm mb-6">
        アカウントにアクセスしてください
      </p>
      <form onSubmit={handleLogin}>
        <InputField
          label="メールアドレス"
          name="email"
          type="email"
          value={form.email}
          onChange={handleChange}
        />
        <InputField
          label="パスワード"
          name="password"
          type="password"
          value={form.password}
          onChange={handleChange}
        />
        <PrimaryButton type="submit">ログイン</PrimaryButton>
      </form>
      <div className="flex justify-between items-center mt-6 text-sm">
        <LinkText to="/forgot-password">パスワードをお忘れですか？</LinkText>
        <LinkText to="/signup">アカウント作成</LinkText>
      </div>
      <div className="relative my-6">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-gray-200"></div>
        </div>
        <div className="relative flex justify-center text-sm">
          <span className="px-2 bg-white text-gray-500">
            またはSNSでログイン
          </span>
        </div>
      </div>
      <SNSSignInButton
        provider="google"
        onClick={() => {
          window.location.href = getCognitoAuthUrl('Google');
        }}
      />
    </AuthLayout>
  );
}
