import { useState, useEffect } from 'react';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import HamburgerMenu from '../components/HamburgerMenu';

export default function ProfilePage() {
  const [form, setForm] = useState({ username: '', email: '', bio: '' });
  const [message, setMessage] = useState(null);
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const token = useSelector((state) => state.auth.accessToken);
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  // ページがレンダリングされたときにプロフィールを取得
  useEffect(() => {
    if (!token) {
      navigate('/login');
      return;
    }

    const fetchProfile = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/profile/me`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        const data = await response.json();
        if (response.ok) {
          setForm({
            username: data.username || '',
            email: data.email || '',
            bio: data.bio || '',
          });
        } else {
          setMessage({
            type: 'error',
            text: data.error || 'プロフィール取得に失敗しました。',
          });
        }
      } catch (error) {
        console.error(error);
        setMessage({ type: 'error', text: '通信エラーが発生しました。' });
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, []);

  const handleChange = (e) => {
    setForm({
      ...form,
      [e.target.name]: e.target.value,
    });
  };

  const handleUpdate = async (e) => {
    e.preventDefault();
    try {
      const token = localStorage.getItem('accessToken');
      const response = await fetch(`${API_BASE_URL}/api/profile/me/update`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(form),
      });
      const data = await response.json();
      if (response.ok) {
        setMessage({
          type: 'success',
          text: data.success || 'プロフィールを更新しました。',
        });
      } else if (response.status === 401) {
        navigate('/login');
        return;
      } else {
        setMessage({
          type: 'error',
          text: data.error || 'プロフィール更新に失敗しました。',
        });
      }
    } catch (error) {
      console.error(error);
      setMessage({ type: 'error', text: '通信エラーが発生しました。' });
    }
  };

  if (loading) {
    return (
      <AuthLayout>
        <p className="text-center">読み込み中...</p>
      </AuthLayout>
    );
  }

  return (
    <>
      <HamburgerMenu title="プロフィール編集" />
      <AuthLayout>
        <div className="mt-16">
          {message && (
            <p
              className={`${
                message.type === 'error' ? 'text-red-600' : 'text-green-600'
              } text-center mb-4`}
            >
              {message.text}
            </p>
          )}
        </div>
        <h2 className="text-2xl font-bold mb-6 text-center">
          プロフィール編集
        </h2>
        <form onSubmit={handleUpdate}>
          <InputField
            label="ニックネーム"
            name="username"
            value={form.username}
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
            label="自己紹介"
            name="bio"
            value={form.bio}
            onChange={handleChange}
          />
          <PrimaryButton type="submit">更新</PrimaryButton>
        </form>
      </AuthLayout>
    </>
  );
}
