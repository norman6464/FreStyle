import { useState } from 'react';
import { useDispatch } from 'react-redux';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import SNSSignInButton from '../components/SNSSignInButton';
import LinkText from '../components/LinkText';
import { getCognitoAuthUrl } from '../utils/auth';
import { useLocation, useNavigate } from 'react-router-dom';
import { setAuthData } from '../store/authSlice';

interface LoginMessage {
  type: 'success' | 'error';
  text: string;
}

export default function LoginPage() {
  const [form, setForm] = useState({ email: '', password: '' });
  const [loginMessage, setLoginMessage] = useState<LoginMessage | null>(null);

  const location = useLocation();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const message = (location.state as { message?: string })?.message;
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    console.log('\n========== ログインリクエスト開始 ==========');
    console.log('📌 リクエスト情報:');
    console.log('   - URL:', `${API_BASE_URL}/api/auth/cognito/login`);
    console.log('   - email:', form.email);
    console.log('   - credentials: include');

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

      console.log('📨 レスポンス受信:');
      console.log('   - status:', response.status);
      console.log('   - ok:', response.ok);
      console.log('   - statusText:', response.statusText);

      console.log('📋 レスポンスヘッダー:');
      response.headers.forEach((value, key) => {
        console.log(`   - ${key}: ${value}`);
      });

      const setCookie = response.headers.get('Set-Cookie');
      console.log('🍪 Set-Cookie ヘッダー:', setCookie || '(アクセス不可 - httpOnlyのため正常)');

      console.log('🍪 現在のdocument.cookie:', document.cookie || '(空)');

      const data = await response.json();
      console.log('📄 レスポンスボディ:', data);

      console.log('ステータスコード', response.status);

      if (response.ok) {
        console.log('✅ ログイン成功 - ホームへ遷移');
        console.log('🍪 遷移前のdocument.cookie:', document.cookie || '(空)');
        dispatch(setAuthData());
        navigate('/');
      } else {
        console.log('❌ ログイン失敗:', data.error);
        setLoginMessage({
          type: 'error',
          text: data.error || 'ログインに失敗しました。',
        });
      }
      console.log('========== ログインリクエスト終了 ==========\n');
    } catch (error) {
      console.log('❌ 通信エラー:', (error as Error).message);
      console.log('========== ログインリクエスト終了(エラー) ==========\n');
      setLoginMessage({ type: 'error', text: '通信エラーが発生しました。' });
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
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
          <p className="text-emerald-600 text-center mb-4 p-3 bg-emerald-50 rounded-lg font-medium">
            {message}
          </p>
        )}
        {loginMessage?.type === 'error' && (
          <p className="text-rose-600 text-center mb-4 p-3 bg-rose-50 rounded-lg font-medium">
            ✕ {loginMessage.text}
          </p>
        )}
      </div>
      <h2 className="text-3xl font-bold mb-2 text-center text-slate-800">
        ログイン
      </h2>
      <p className="text-center text-slate-500 text-sm mb-6">
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
          <div className="w-full border-t border-slate-200"></div>
        </div>
        <div className="relative flex justify-center text-sm">
          <span className="px-2 bg-white text-slate-500">
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
