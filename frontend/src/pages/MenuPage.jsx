import { useNavigate } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { useState, useEffect } from 'react';
import HamburgerMenu from '../components/HamburgerMenu';
import {
  CalendarIcon,
  UserGroupIcon,
  EnvelopeIcon,
} from '@heroicons/react/24/solid';

export default function MenuPage() {
  const navigate = useNavigate();
  const message = useSelector((state) => state.flash?.message);
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const [stats, setStats] = useState(null);
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
            navigate('/login');
            return;
          }

          // アクセストークンはhttpOnly cookieで管理しているのでここでは取得しない
          const refreshData = await refreshRes.json();

          // リトライ
          const retryRes = await fetch(`${API_BASE_URL}/api/chat/stats`, {
            headers: {
              'Content-Type': 'application/json'
            },
            credentials: 'include',
          });

          if (!retryRes.ok) return;
          const retryData = await retryRes.json();
          setStats(retryData);
          return;
        }

        if (!res.ok) return;
        const data = await res.json();
        setStats(data);
      } catch (err) {
        console.error('Error fetching stats:', err);
      }
    };

    fetchStats();
  }, [navigate, API_BASE_URL]);

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
    <>
      {message && (
        <div className="fixed top-20 left-0 right-0 flex justify-center z-50 animate-fade-in">
          <div className="bg-green-50 border-l-4 border-green-500 text-green-700 px-6 py-4 rounded-lg shadow-lg font-semibold">
            ✓ {message}
          </div>
        </div>
      )}
      <HamburgerMenu title="ホーム" />
      <div className="min-h-screen bg-gradient-to-br from-gray-50 to-primary-50 px-4 pt-20 pb-8">
        <div className="w-full max-w-2xl mx-auto">
          {/* 日時セクション */}
          <div className="mb-8 bg-white rounded-xl shadow-lg p-6 border border-gray-100">
            <div className="flex items-center gap-3 mb-4">
              <CalendarIcon className="w-6 h-6 text-primary-600" />
              <span className="text-lg font-semibold text-gray-700">
                {formatDate(currentTime)}
              </span>
            </div>
            <div className="text-5xl font-bold text-primary-600 font-mono">
              {formatTime(currentTime)}
            </div>
          </div>

          {/* ユーザー情報セクション */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
            {/* ユーザー総数 */}
            <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100 hover:shadow-xl hover:scale-105 transition-all duration-300">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-semibold text-gray-600 uppercase">
                  ユーザー総数
                </h3>
                <UserGroupIcon className="w-6 h-6 text-blue-500" />
              </div>
              <p className="text-3xl font-bold text-blue-600">
                {stats?.totalUsers || '—'}
              </p>
            </div>

            {/* メールアドレス */}
            <div className="bg-white rounded-xl shadow-lg p-6 border border-gray-100 hover:shadow-xl hover:scale-105 transition-all duration-300">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-semibold text-gray-600 uppercase">
                  あなたのメール
                </h3>
                <EnvelopeIcon className="w-6 h-6 text-purple-500" />
              </div>
            </div>
          </div>

          {/* ウェルカムセクション */}
          <div className="text-center mb-10">
            <h1 className="text-4xl font-bold text-gray-800 mb-2">FreStyle</h1>
            <p className="text-gray-600 text-lg">友達とシームレスにチャット</p>
          </div>

          {/* メニュー項目 */}
          <div className="space-y-4">
            {/* プロフィール編集 */}
            <div
              onClick={() => navigate('/profile/me')}
              className="bg-white shadow-lg rounded-xl p-6 cursor-pointer hover:shadow-2xl hover:bg-gradient-to-r hover:from-primary-50 hover:to-secondary-50 transition-all duration-300 border border-gray-100 hover:border-primary-300 transform hover:scale-105 group"
            >
              <div className="flex items-start">
                <div className="bg-gradient-primary rounded-lg p-3 mr-4 group-hover:shadow-lg transition-all">
                  <svg
                    className="w-6 h-6 text-white"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                    />
                  </svg>
                </div>
                <div>
                  <h2 className="text-xl font-bold text-gray-800 group-hover:text-primary-600 transition-colors">
                    プロフィールを編集
                  </h2>
                  <p className="text-gray-600 text-sm mt-1">
                    ユーザー情報の編集をする
                  </p>
                </div>
                <svg
                  className="w-5 h-5 text-gray-300 group-hover:text-primary-400 ml-auto transition-colors transform group-hover:translate-x-1"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </div>
            </div>

            {/* ユーザー検索 */}
            <div
              onClick={() => navigate('/chat/users')}
              className="bg-white shadow-lg rounded-xl p-6 cursor-pointer hover:shadow-2xl hover:bg-gradient-to-r hover:from-primary-50 hover:to-secondary-50 transition-all duration-300 border border-gray-100 hover:border-primary-300 transform hover:scale-105 group"
            >
              <div className="flex items-start">
                <div className="bg-gradient-secondary rounded-lg p-3 mr-4 group-hover:shadow-lg transition-all">
                  <svg
                    className="w-6 h-6 text-white"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                    />
                  </svg>
                </div>
                <div>
                  <h2 className="text-xl font-bold text-gray-800 group-hover:text-secondary-600 transition-colors">
                    ユーザー検索
                  </h2>
                  <p className="text-gray-600 text-sm mt-1">
                    メールアドレスで追加し、チャットを開始する
                  </p>
                </div>
                <svg
                  className="w-5 h-5 text-gray-300 group-hover:text-secondary-400 ml-auto transition-colors transform group-hover:translate-x-1"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </div>
            </div>

            {/* AIに聞いてみる */}
            <div
              onClick={() => navigate('/chat/ask-ai')}
              className="bg-white shadow-lg rounded-xl p-6 cursor-pointer hover:shadow-2xl hover:bg-gradient-to-r hover:from-pink-50 hover:to-orange-50 transition-all duration-300 border border-gray-100 hover:border-pink-300 transform hover:scale-105 group"
            >
              <div className="flex items-start">
                <div className="bg-gradient-to-br from-pink-400 to-orange-400 rounded-lg p-3 mr-4 group-hover:shadow-lg transition-all">
                  <svg
                    className="w-6 h-6 text-white"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-3.536 3.536l-.707.707M9.172 9.172L8.465 8.465m5.656 0l-.707.707"
                    />
                  </svg>
                </div>
                <div>
                  <h2 className="text-xl font-bold text-gray-800 group-hover:text-pink-600 transition-colors\">
                    AIに聞いてみる
                  </h2>
                  <p className="text-gray-600 text-sm mt-1\">
                    AIに質問して素早く答えを得る
                  </p>
                </div>
                <svg
                  className="w-5 h-5 text-gray-300 group-hover:text-pink-400 ml-auto transition-colors transform group-hover:translate-x-1"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
