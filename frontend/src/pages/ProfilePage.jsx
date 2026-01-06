import { useState, useEffect } from 'react';
import InputField from '../components/InputField';
import PrimaryButton from '../components/PrimaryButton';
import { useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import HamburgerMenu from '../components/HamburgerMenu';
import { clearAuth } from '../store/authSlice';
import {
  UserCircleIcon,
  SparklesIcon,
  ChatBubbleLeftRightIcon,
} from '@heroicons/react/24/solid';

export default function ProfilePage() {
  const [form, setForm] = useState({ name: '', bio: '' });
  const [message, setMessage] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const dispatch = useDispatch();

  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  // ----------------------------
  // プロフィール取得 (アクセストークン + リフレッシュ対応)
  // ----------------------------
  const fetchProfile = async () => {
    try {
      console.log('[ProfilePage] Fetching profile');
      const res = await fetch(`${API_BASE_URL}/api/profile/me`, {
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include', // RefreshToken送信
      });

      // トークン期限切れならリフレッシュを試みる
      if (res.status === 401) {
        console.warn('[ProfilePage] Access token expired, attempting refresh');
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          {
            method: 'POST',
            credentials: 'include',
          }
        );

        if (!refreshRes.ok) {
          console.error('[ProfilePage] ERROR: Token refresh failed');
          dispatch(clearAuth());
          return;
        }

        console.log('[ProfilePage] Access token refreshed successfully, retrying fetch');

        const retryRes = await fetch(`${API_BASE_URL}/api/profile/me`, {
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
        });

        const retryData = await retryRes.json();
        if (!retryRes.ok)
          throw new Error(
            retryData.error || 'プロフィール再取得に失敗しました。'
          );

        console.log('[ProfilePage] Profile fetch retry successful');
        setForm({
          name: retryData.name || '',
          bio: retryData.bio || '',
        });
        setLoading(false);
        return;
      }

      const data = await res.json();
      if (!res.ok)
        throw new Error(data.error || 'プロフィール取得に失敗しました。');

      console.log('[ProfilePage] Profile fetched successfully');
      setForm({
        name: data.name || '',
        bio: data.bio || '',
      });
    } catch (err) {
      console.error('[ProfilePage] ERROR: Profile fetch failed -', err.message);
      setMessage({ type: 'error', text: err.message });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, []);

  // ----------------------------
  // プロフィール更新
  // ----------------------------
  const handleUpdate = async (e) => {
    e.preventDefault();
    try {
      console.log('[ProfilePage] Updating profile with name:', form.name);
      const res = await fetch(`${API_BASE_URL}/api/profile/me/update`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify(form),
      });

      // 401 → トークン更新を試みる
      if (res.status === 401) {
        console.warn('[ProfilePage] Access token expired, attempting refresh');
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          {
            method: 'POST',
            credentials: 'include',
          }
        );

        if (!refreshRes.ok) {
          console.error('[ProfilePage] ERROR: Token refresh failed');
          navigate('/login');
          return;
        }

        console.log('[ProfilePage] Access token refreshed, retrying update');
        const refreshData = await refreshRes.json();
        const retryRes = await fetch(`${API_BASE_URL}/api/profile/me/update`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
          body: JSON.stringify(form),
        });

        const retryData = await retryRes.json();
        if (!retryRes.ok)
          throw new Error(
            retryData.error || 'プロフィール更新に失敗しました。'
          );

        console.log('[ProfilePage] Profile update successful');
        setMessage({
          type: 'success',
          text: retryData.success || 'プロフィールを更新しました。',
        });
        return;
      }

      const data = await res.json();
      if (!res.ok)
        throw new Error(data.error || 'プロフィール更新に失敗しました。');

      console.log('[ProfilePage] Profile update successful');
      setMessage({
        type: 'success',
        text: data.success || 'プロフィールを更新しました。',
      });
    } catch (error) {
      console.error('[ProfilePage] ERROR: Profile update failed -', error.message);
      setMessage({ type: 'error', text: '通信エラーが発生しました。' });
    }
  };

  // ローディング時
  if (loading) {
    return (
      <>
        <HamburgerMenu title="プロフィール編集" />
        <div className="min-h-screen bg-gradient-to-br from-gray-50 to-primary-50 flex items-center justify-center pt-20">
          <div className="text-center">
            <div className="animate-pulse">
              <div className="w-16 h-16 bg-primary-200 rounded-full mx-auto mb-4"></div>
              <p className="text-gray-600">読み込み中...</p>
            </div>
          </div>
        </div>
      </>
    );
  }

  // 表示
  return (
    <>
      <HamburgerMenu title="プロフィール" />
      <div className="min-h-screen bg-gradient-to-br from-gray-50 to-primary-50 pt-20 pb-8 px-4">
        <div className="max-w-2xl mx-auto">
          {/* メッセージ */}
          {message && (
            <div
              className={`mb-6 p-4 rounded-lg border-l-4 flex items-start animate-fade-in ${
                message.type === 'error'
                  ? 'bg-red-50 border-red-500'
                  : 'bg-green-50 border-green-500'
              }`}
            >
              <div
                className={`flex-shrink-0 mr-3 ${
                  message.type === 'error' ? 'text-red-600' : 'text-green-600'
                }`}
              >
                {message.type === 'error' ? (
                  <svg
                    className="w-5 h-5"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                      clipRule="evenodd"
                    />
                  </svg>
                ) : (
                  <svg
                    className="w-5 h-5"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                )}
              </div>
              <p
                className={
                  message.type === 'error'
                    ? 'text-red-700 font-semibold'
                    : 'text-green-700 font-semibold'
                }
              >
                {message.text}
              </p>
            </div>
          )}

          {/* プロフィールカード */}
          <div className="bg-white rounded-2xl shadow-lg p-8 border border-gray-100">
            {/* ヘッダー */}
            <div className="text-center mb-8">
              <div className="relative inline-block">
                <div className="w-28 h-28 bg-gradient-to-br from-primary-400 via-secondary-400 to-pink-400 rounded-full mx-auto flex items-center justify-center shadow-xl">
                  <span className="text-white text-5xl font-bold">
                    {form.name?.charAt(0)?.toUpperCase() || 'U'}
                  </span>
                </div>
                {/* 編集アイコン */}
                <div className="absolute bottom-0 right-0 bg-white rounded-full p-2 shadow-lg border-2 border-gray-100">
                  <svg className="w-5 h-5 text-primary-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </div>
              </div>
              <h2 className="text-2xl font-bold text-gray-800 mt-4">
                プロフィールを編集
              </h2>
              <p className="text-gray-500 text-sm mt-1">
                あなたの情報を更新してより良いコミュニケーションを
              </p>
            </div>

            <form onSubmit={handleUpdate} className="space-y-6">
              <InputField
                label="ニックネーム"
                name="name"
                value={form.name}
                onChange={(e) =>
                  setForm((prev) => ({ ...prev, name: e.target.value }))
                }
              />
              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  自己紹介
                </label>
                <textarea
                  name="bio"
                  value={form.bio}
                  onChange={(e) =>
                    setForm((prev) => ({ ...prev, bio: e.target.value }))
                  }
                  placeholder="あなたについて教えてください..."
                  rows="4"
                  className="w-full border-2 border-gray-200 rounded-xl px-4 py-3 focus:border-primary-500 focus:ring-2 focus:ring-primary-100 transition-all duration-200 resize-none"
                />
              </div>
              <PrimaryButton type="submit">プロフィールを更新</PrimaryButton>
            </form>
          </div>

          {/* クイックリンク */}
          <div className="mt-6 grid grid-cols-2 gap-4">
            <div
              onClick={() => navigate('/profile/personality')}
              className="bg-white rounded-xl shadow-md p-4 cursor-pointer hover:shadow-lg transition-all border border-gray-100 hover:border-purple-300 group"
            >
              <div className="flex items-center gap-3">
                <div className="bg-gradient-to-br from-purple-400 to-pink-400 rounded-lg p-2">
                  <SparklesIcon className="w-5 h-5 text-white" />
                </div>
                <div>
                  <p className="font-semibold text-gray-800 text-sm group-hover:text-purple-600 transition-colors">パーソナリティ</p>
                  <p className="text-xs text-gray-500">AI設定を編集</p>
                </div>
              </div>
            </div>
            <div
              onClick={() => navigate('/chat')}
              className="bg-white rounded-xl shadow-md p-4 cursor-pointer hover:shadow-lg transition-all border border-gray-100 hover:border-blue-300 group"
            >
              <div className="flex items-center gap-3">
                <div className="bg-gradient-to-br from-blue-400 to-cyan-400 rounded-lg p-2">
                  <ChatBubbleLeftRightIcon className="w-5 h-5 text-white" />
                </div>
                <div>
                  <p className="font-semibold text-gray-800 text-sm group-hover:text-blue-600 transition-colors">チャット一覧</p>
                  <p className="text-xs text-gray-500">会話を見る</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
