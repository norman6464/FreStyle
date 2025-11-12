import { useState, useEffect } from 'react';
import AuthLayout from '../components/AuthLayout';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import { useSelector, useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import HamburgerMenu from '../components/HamburgerMenu';
import { setAuthData, clearAuthData } from '../store/authSlice';

export default function ProfilePage() {
  const [form, setForm] = useState({ username: '', bio: '' });
  const [message, setMessage] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const dispatch = useDispatch();

  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const accessToken = useSelector((state) => state.auth.accessToken);
  const email = useSelector((state) => state.auth.email);

  // ----------------------------
  // プロフィール取得 (アクセストークン + リフレッシュ対応)
  // ----------------------------
  const fetchProfile = async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/profile/me`, {
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${accessToken}`,
        },
        credentials: 'include', // RefreshToken送信
      });

      // トークン期限切れならリフレッシュを試みる
      if (res.status === 401) {
        console.warn('アクセストークン期限切れ、リフレッシュ試行');
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          {
            method: 'POST',
            credentials: 'include',
          }
        );

        if (!refreshRes.ok) {
          console.error('リフレッシュ失敗、ログインへリダイレクト');
          dispatch(clearAuthData());
          navigate('/login');
          return;
        }

        const refreshData = await refreshRes.json();
        const newAccessToken = refreshData.accessToken;

        if (!newAccessToken) {
          dispatch(clearAuthData());
          navigate('/login');
          return;
        }

        // Redux更新
        dispatch(setAuthData({ accessToken: newAccessToken }));
        console.log('✅ アクセストークン更新済み、再試行');

        const retryRes = await fetch(`${API_BASE_URL}/api/profile/me`, {
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${newAccessToken}`,
          },
          credentials: 'include',
        });

        const retryData = await retryRes.json();
        if (!retryRes.ok)
          throw new Error(
            retryData.error || 'プロフィール再取得に失敗しました。'
          );

        setForm({
          username: retryData.username || '',
          bio: retryData.bio || '',
        });
        setLoading(false);
        return;
      }

      const data = await res.json();
      if (!res.ok)
        throw new Error(data.error || 'プロフィール取得に失敗しました。');

      setForm({
        username: data.username || '',
        bio: data.bio || '',
      });
    } catch (err) {
      console.error('❌ プロフィール取得失敗:', err);
      setMessage({ type: 'error', text: err.message });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, [accessToken, email]);

  // ----------------------------
  // プロフィール更新
  // ----------------------------
  const handleUpdate = async (e) => {
    e.preventDefault();
    try {
      const res = await fetch(`${API_BASE_URL}/api/profile/me/update`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${accessToken}`,
        },
        credentials: 'include',
        body: JSON.stringify(form),
      });

      // 401 → トークン更新を試みる
      if (res.status === 401) {
        console.warn('アクセストークン期限切れ、リフレッシュ試行');
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token?email=${encodeURIComponent(
            email
          )}`,
          {
            method: 'POST',
            credentials: 'include',
          }
        );

        if (!refreshRes.ok) {
          dispatch(clearAuthData());
          navigate('/login');
          return;
        }

        const refreshData = await refreshRes.json();
        const newAccessToken = refreshData.accessToken;
        if (!newAccessToken) {
          dispatch(clearAuthData());
          navigate('/login');
          return;
        }

        // Redux更新
        dispatch(setAuthData({ accessToken: newAccessToken }));

        console.log('✅ アクセストークン更新済み、再試行');
        const retryRes = await fetch(`${API_BASE_URL}/api/profile/me/update`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${newAccessToken}`,
          },
          credentials: 'include',
          body: JSON.stringify(form),
        });

        const retryData = await retryRes.json();
        if (!retryRes.ok)
          throw new Error(
            retryData.error || 'プロフィール更新に失敗しました。'
          );

        setMessage({
          type: 'success',
          text: retryData.success || 'プロフィールを更新しました。',
        });
        return;
      }

      const data = await res.json();
      if (!res.ok)
        throw new Error(data.error || 'プロフィール更新に失敗しました。');

      setMessage({
        type: 'success',
        text: data.success || 'プロフィールを更新しました。',
      });
    } catch (error) {
      console.error(error);
      setMessage({ type: 'error', text: '通信エラーが発生しました。' });
    }
  };

  // ----------------------------
  // ローディング時
  // ----------------------------
  if (loading) {
    return (
      <AuthLayout>
        <p className="text-center">読み込み中...</p>
      </AuthLayout>
    );
  }

  // ----------------------------
  // 表示
  // ----------------------------
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
            onChange={(e) =>
              setForm((prev) => ({ ...prev, username: e.target.value }))
            }
          />
          <InputField
            label="自己紹介"
            name="bio"
            value={form.bio}
            onChange={(e) =>
              setForm((prev) => ({ ...prev, bio: e.target.value }))
            }
          />
          <PrimaryButton type="submit">更新</PrimaryButton>
        </form>
      </AuthLayout>
    </>
  );
}
