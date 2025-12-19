import { useState } from 'react';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import LinkText from '../components/LinkText';
import { useNavigate } from 'react-router-dom';
import {
  CheckCircleIcon,
  ExclamationCircleIcon,
} from '@heroicons/react/24/solid';

export default function SignupPage() {
  const [form, setForm] = useState({ email: '', password: '', name: '' });
  const [message, setMessage] = useState(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const handleChange = (e) => {
    setForm({
      ...form,
      [e.target.name]: e.target.value,
    });
  };

  const handleSignup = async (e) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await fetch(`${API_BASE_URL}/api/auth/cognito/signup`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: form.email,
          password: form.password,
          name: form.name,
        }),
      });

      const data = await response.json();

      if (response.ok) {
        setMessage({ type: 'success', text: data.message });
        setTimeout(() => {
          navigate('/confirm', {
            state: {
              message:
                '✓ サインアップに成功しました！メール確認をお願いします。',
            },
          });
        }, 1500);
      } else {
        setMessage({
          type: 'error',
          text: data.error || '登録に失敗しました。',
        });
      }
    } catch (error) {
      console.log(error);
      setMessage({ type: 'error', text: '通信エラーが発生しました。' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthLayout>
      <div>
        {message?.type === 'error' && (
          <div className="mb-6 p-4 bg-red-50 border-l-4 border-red-500 rounded-lg flex items-start animate-fade-in">
            <ExclamationCircleIcon className="w-5 h-5 text-red-600 mr-3 flex-shrink-0 mt-0.5" />
            <p className="text-red-700 font-semibold text-sm">{message.text}</p>
          </div>
        )}
        {message?.type === 'success' && (
          <div className="mb-6 p-4 bg-green-50 border-l-4 border-green-500 rounded-lg flex items-start animate-fade-in">
            <CheckCircleIcon className="w-5 h-5 text-green-600 mr-3 flex-shrink-0" />
            <p className="text-green-700 font-semibold text-sm">
              {message.text}
            </p>
          </div>
        )}
      </div>
      <h2 className="text-3xl font-bold mb-2 text-center text-gray-800">
        アカウント作成
      </h2>
      <p className="text-center text-gray-600 text-sm mb-8">
        FreStyleに参加して、友達とチャットを始めましょう
      </p>
      <form onSubmit={handleSignup}>
        <InputField
          label="ニックネーム"
          name="name"
          value={form.name}
          onChange={handleChange}
        />
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
        <PrimaryButton type="submit" disabled={loading}>
          {loading ? '作成中...' : 'アカウント作成'}
        </PrimaryButton>
      </form>
      <div className="mt-6 text-center">
        <p className="text-gray-600 text-sm mb-2">
          すでにアカウントをお持ちですか？
        </p>
        <LinkText to="/login">ログインする</LinkText>
      </div>
    </AuthLayout>
  );
}
