import { useEffect, useState } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { clearAuth } from '../store/authSlice';
import {
  ChatBubbleLeftIcon,
  UserGroupIcon,
  SparklesIcon,
  CalendarIcon,
} from '@heroicons/react/24/solid';

export default function HomePage() {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const email = useSelector((state) => state.auth.email);
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [currentTime, setCurrentTime] = useState(new Date());

  // 時刻の自動更新
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date());
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  // 統計情報の取得
  useEffect(() => {
    const fetchStats = async () => {
      try {
        setLoading(true);

        const res = await fetch(`${API_BASE_URL}/api/chat/stats`, {
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
        });

        // トークン期限切れの場合
        if (res.status === 401) {
          const refreshRes = await fetch(
            `${API_BASE_URL}/api/auth/cognito/refresh-token`,
            {
              method: 'POST',
              credentials: 'include',
            }
          );

          if (!refreshRes.ok) {
            dispatch(clearAuth());
            navigate('/login');
            return;
          }

          const refreshData = await refreshData.json();
          dispatch(setAuthData({ accessToken: refreshData.accessToken }));

          // リトライ
          const retryRes = await fetch(`${API_BASE_URL}/api/chat/stats`, {
            headers: {
              'Content-Type': 'application/json',
            },
            credentials: 'include',
          });

          if (!retryRes.ok) throw new Error('統計情報の取得に失敗');
          const retryData = await retryRes.json();
          setStats(retryData);
          return;
        }

        if (!res.ok) throw new Error('統計情報の取得に失敗');
        const data = await res.json();
        setStats(data);
      } catch (err) {
        console.error('Error fetching stats:', err);
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    if (accessToken) {
      fetchStats();
    }
  }, [dispatch, navigate, API_BASE_URL]);

  const formatDate = (date) => {
    return date.toLocaleDateString('ja-JP', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      weekday: 'long',
    });
  };

  const formatTime = (date) => {
    return date.toLocaleTimeString('ja-JP', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-primary-50 p-4 pt-6 pb-8">
      <div className="max-w-2xl mx-auto">
        {/* ウェルカムセクション */}
        <div className="mb-8 bg-white rounded-xl shadow-lg p-8 border border-gray-100">
          <h1 className="text-4xl font-bold text-gray-800 mb-2">ようこそ！</h1>
          {stats?.username && (
            <p className="text-xl text-primary-600 font-semibold mb-4">
              {stats.username} さん
            </p>
          )}

          {/* 日付と時刻 */}
          <div className="bg-gradient-to-br from-primary-50 to-secondary-50 rounded-lg p-6 mb-6">
            <div className="flex items-center gap-3 mb-3">
              <CalendarIcon className="w-6 h-6 text-primary-600" />
              <span className="text-lg font-semibold text-gray-700">
                {formatDate(currentTime)}
              </span>
            </div>
            <div className="text-4xl font-bold text-primary-600">
              {formatTime(currentTime)}
            </div>
          </div>

          <p className="text-gray-600 text-lg leading-relaxed">
            今日も一日頑張りましょう！新しいメッセージや、友達との会話をチェックしてみてください。
          </p>
        </div>

        {/* 統計情報 */}
        {loading ? (
          <div className="text-center py-8">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
            <p className="text-gray-600 mt-4">統計情報を読み込み中...</p>
          </div>
        ) : error ? (
          <div className="bg-red-50 border-l-4 border-red-500 rounded-lg p-6 mb-8">
            <p className="text-red-700 font-semibold">⚠️ {error}</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
            {/* ユーザー数 */}
            <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100 hover:shadow-xl hover:scale-105 transition-all duration-300">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-800">
                  登録ユーザー数
                </h3>
                <UserGroupIcon className="w-8 h-8 text-blue-500" />
              </div>
              <p className="text-4xl font-bold text-blue-600">
                {stats?.totalUsers || 0}
              </p>
              <p className="text-sm text-gray-500 mt-2">
                全体のアクティブユーザー
              </p>
            </div>

            {/* メールアドレス */}
            <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100 hover:shadow-xl hover:scale-105 transition-all duration-300">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-800">
                  あなたのメール
                </h3>
                <ChatBubbleLeftIcon className="w-8 h-8 text-purple-500" />
              </div>
              <p className="text-lg font-mono text-purple-600 break-all">
                {stats?.email || email || 'メールアドレス不明'}
              </p>
              <p className="text-sm text-gray-500 mt-2">
                アカウント登録に使用されたメールアドレス
              </p>
            </div>
          </div>
        )}

        {/* クイックアクションボタン */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <button
            onClick={() => navigate('/chat/members')}
            className="bg-gradient-primary text-white rounded-lg p-4 font-semibold hover:shadow-lg hover:scale-105 transition-all duration-300 flex items-center justify-center gap-2"
          >
            <ChatBubbleLeftIcon className="w-5 h-5" />
            メッセージ
          </button>
          <button
            onClick={() => navigate('/chat/users')}
            className="bg-gradient-secondary text-white rounded-lg p-4 font-semibold hover:shadow-lg hover:scale-105 transition-all duration-300 flex items-center justify-center gap-2"
          >
            <UserGroupIcon className="w-5 h-5" />
            ユーザー検索
          </button>
          <button
            onClick={() => navigate('/chat/ask-ai')}
            className="bg-gradient-to-r from-pink-500 to-orange-500 text-white rounded-lg p-4 font-semibold hover:shadow-lg hover:scale-105 transition-all duration-300 flex items-center justify-center gap-2"
          >
            <SparklesIcon className="w-5 h-5" />
            AI チャット
          </button>
        </div>
      </div>
    </div>
  );
}
