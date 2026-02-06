import { useEffect, useMemo, useState } from 'react';
import { useDispatch } from 'react-redux';
import { debounce } from 'lodash';
import { useNavigate } from 'react-router-dom';
import MemberList from '../components/MemberList';
import HamburgerMenu from '../components/HamburgerMenu';
import {
  MagnifyingGlassIcon,
  UserPlusIcon,
  SparklesIcon,
} from '@heroicons/react/24/solid';

export default function AddUserPage() {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const [users, setUsers] = useState([]);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [debounceQuery, setDebounceQuery] = useState('');

  const debounceSearch = useMemo(
    () => debounce((query) => setDebounceQuery(query), 500),
    []
  );

useEffect(() => {
  const controller = new AbortController();

  const fetchWithAuth = async () => {
    const queryParam = debounceQuery
      ? `?query=${encodeURIComponent(debounceQuery)}`
      : '';

    const fetchUsers = async () => {
      const res = await fetch(`${API_BASE_URL}/api/chat/users${queryParam}`, {
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        signal: controller.signal,
      });
      return res;
    };

    try {
      let res = await fetchUsers();

      // 401ならリフレッシュトークン試行
      if (res.status === 401) {
        console.log('アクセストークン期限切れ。再発行試行');
        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          {
            method: 'POST',
            credentials: 'include',
            signal: controller.signal,
          }
        );

        if (!refreshRes.ok) {
          console.log('リフレッシュ失敗。再ログインへ');
          navigate('/login');
          return;
        }

        // 再試行
        res = await fetchUsers();
      }

      if (!res.ok) {
        throw new Error('ユーザー取得に失敗しました');
      }

      const data = await res.json();
      setUsers(data.users || []);
    } catch (err) {
      if (err.name !== 'AbortError') {
        setError(err.message);
      }
    }
  };

  fetchWithAuth();

  return () => controller.abort();
}, [debounceQuery, navigate, dispatch, API_BASE_URL]);

  return (
    <>
      <HamburgerMenu title="ユーザー検索" />

      <div className="min-h-screen bg-gray-50 pt-16 pb-24">
        {/* ヘッダーセクション */}
        <div className="bg-primary-500 px-4 py-6 mb-6">
          <div className="max-w-2xl mx-auto text-center">
            <h2 className="text-xl font-bold text-white mb-1">
              友達を探す
            </h2>
            <p className="text-white/80 text-sm">
              ユーザー名やメールアドレスで検索して、チャットを始めよう
            </p>
          </div>
        </div>

        {/* 検索セクション */}
        <div className="sticky top-16 z-10 bg-white border-b border-gray-200 px-4 py-4">
          <div className="max-w-2xl mx-auto">
            <div className="relative">
              <MagnifyingGlassIcon className="absolute left-4 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value);
                  debounceSearch(e.target.value);
                }}
                placeholder="ユーザー名またはメールアドレスで検索..."
                className="w-full pl-12 pr-4 py-3 bg-gray-100 rounded-xl border-none focus:outline-none focus:ring-1 focus:ring-primary-500 transition-colors duration-150 text-base"
              />
            </div>
          </div>
        </div>

        {/* メインコンテンツ */}
        <div className="max-w-2xl mx-auto px-4 pt-4">
          {/* エラー表示 */}
          {error && (
            <div className="mb-6 p-4 bg-red-50 border-l-4 border-red-500 rounded-lg">
              <p className="text-red-700 font-medium">
                ⚠️ エラーが発生しました
              </p>
              <p className="text-red-600 text-sm mt-1">{error}</p>
            </div>
          )}

          {/* 検索前の状態 */}
          {users.length === 0 && !debounceQuery && (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <div className="bg-gray-100 rounded-full p-8 mb-6">
                <MagnifyingGlassIcon className="w-12 h-12 text-gray-400" />
              </div>
              <h3 className="text-lg font-semibold text-gray-700 mb-2">
                ユーザーを検索してみましょう
              </h3>
              <p className="text-gray-500 text-sm max-w-xs">
                名前やメールアドレスを入力して、チャットしたい相手を探してください
              </p>
              
              {/* ヒント */}
              <div className="mt-8 bg-yellow-50 rounded-xl p-4 border border-yellow-200 max-w-sm">
                <div className="flex items-start gap-3">
                  <span className="text-2xl">💡</span>
                  <div className="text-left">
                    <p className="text-sm font-semibold text-yellow-800">ヒント</p>
                    <p className="text-xs text-yellow-700 mt-1">
                      メールアドレスの一部でも検索できます。相手のメールアドレスがわかる場合は、より正確に見つけられます。
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* 検索結果なし */}
          {users.length === 0 && debounceQuery && (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <div className="bg-gray-100 rounded-full p-8 mb-6">
                <svg className="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-gray-700 mb-2">
                ユーザーが見つかりませんでした
              </h3>
              <p className="text-gray-500 text-sm max-w-xs">
                「{debounceQuery}」に一致するユーザーはいません。検索条件を変えてみてください。
              </p>
            </div>
          )}

          {/* ユーザーリスト */}
          {users.length > 0 && (
            <div>
              <div className="flex items-center gap-2 mb-4">
                <UserPlusIcon className="w-5 h-5 text-gray-500" />
                <span className="text-sm font-semibold text-gray-600">
                  {users.length}人のユーザーが見つかりました
                </span>
              </div>
              <MemberList users={users} />
            </div>
          )}

          {/* AI分析への導線 */}
          <div className="mt-8">
            <div
              onClick={() => navigate('/chat/ask-ai')}
              className="bg-white rounded-xl p-4 border border-gray-200 cursor-pointer hover:bg-gray-50 transition-colors duration-150"
            >
              <div className="flex items-center gap-3">
                <div className="bg-primary-100 rounded-lg p-2">
                  <SparklesIcon className="w-5 h-5 text-primary-500" />
                </div>
                <div className="flex-1">
                  <p className="font-medium text-gray-800">AIでチャットを分析</p>
                  <p className="text-xs text-gray-500">印象のギャップを発見しよう</p>
                </div>
                <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
