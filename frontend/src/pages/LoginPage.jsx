import { useState, useEffect } from 'react';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import SNSSignInButton from '../components/SNSSignInButton';
import LinkText from '../components/LinkText';
import { getCognitoAuthUrl } from '../utils/auth';
import { useLocation, useNavigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { setAuthData } from '../store/authSlice';

export default function LoginPage() {
  const [form, setForm] = useState({ email: '', password: '' });
  // フラッシュメッセージ
  const [loginMessage, setLoginMessage] = useState(null);

  // SignupPageで登録成功時のメッセージが表示される
  const location = useLocation();
  const navigate = useNavigate();
  const message = location.state?.message;
  const dispatch = useDispatch();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  // useEffect(() => {
  //   if (location.state?.message) {
  //     // stateを空にする（メッセージを１回表示したら消す）
  //     const timer = setTimeout(() => {
  //       navigate(location.pathname, { replace: true });
  //     }, 100); // すぐにリセット

  //     return () => clearTimeout(timer);
  //   }
  // }, [location, navigate]);

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
      });

      const data = await response.json();
      if (response.ok) {
        dispatch(setAuthData(data));
        navigate('/', {
          message: 'ログイン成功しました。',
        });
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
        {message && <p className="text-green-600 text-center">{message}</p>}
        {loginMessage?.type === 'error' && (
          <p className="text-red-600 text-center">{loginMessage.text}</p>
        )}
      </div>
      <h2 className="text-2xl font-bold mb-6 text-center">ログイン</h2>
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
      <div className="flex justify-between mt-4">
        <LinkText to="/forget-password">パスワードをお忘れですか？</LinkText>
        <LinkText to="/signup">アカウントを作成</LinkText>
      </div>
      <hr />
      <div className="my-6 text-center text-sm text-gray-500">
        またはSNSでログイン
      </div>
      <SNSSignInButton
        provider="google"
        onClick={() => {
          window.location.href = getCognitoAuthUrl('Google');
        }}
      />
      <SNSSignInButton
        provider="facebook"
        onClick={() => console.log('Facebook login')}
      />
      <SNSSignInButton provider="x" onClick={() => console.log('X login')} />
    </AuthLayout>
  );
}
