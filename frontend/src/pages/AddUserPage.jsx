import { useEffect, useMemo, useState } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { debounce } from 'lodash';
import { useNavigate } from 'react-router-dom';
import MemberList from '../components/MemberList';
import SearchBox from '../components/SearchBox';
import HamburgerMenu from '../components/HamburgerMenu';
import { setAuthData, clearAuthData } from '../store/authSlice';

export default function AddUserPage() {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const email = useSelector((state) => state.auth.email); // refresh-token API用

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
          dispatch(clearAuthData());
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
      <HamburgerMenu title="ユーザーチャット" />

      <div className="min-h-screen bg-gradient-to-br from-gray-50 to-primary-50 p-4 pt-20 pb-8">
        <div className="max-w-2xl mx-auto">
          {/* ヘッダーセクション */}
          <div className="mb-8">
            <h2 className="text-4xl font-bold text-gray-800 mb-2">
              ユーザーを探す
            </h2>
            <p className="text-gray-600 text-lg">チャットを始めたい人を検索</p>
          </div>

          {/* 検索ボックス */}
          <div className="mb-8">
            <SearchBox
              value={searchQuery}
              onChange={setSearchQuery}
              placeholder="ユーザー名またはメールアドレスで検索..."
            />
          </div>

          {/* エラー表示 */}
          {error && (
            <div className="mb-6 p-4 bg-red-50 border-l-4 border-red-500 rounded-lg animate-fade-in">
              <p className="text-red-700 font-semibold">
                ⚠️ エラーが発生しました
              </p>
              <p className="text-red-600 text-sm mt-1">{error}</p>
            </div>
          )}

          {/* ユーザーリスト */}
          <div>
            {users.length === 0 && debounceQuery && (
              <div className="text-center py-12">
                <p className="text-gray-400 text-lg">検索結果がありません</p>
              </div>
            )}
            {users.length === 0 && !debounceQuery && (
              <div className="text-center py-12">
                <p className="text-gray-400 text-lg">
                  ユーザーを検索してください
                </p>
              </div>
            )}
            <MemberList users={users} />
          </div>
        </div>
      </div>
    </>
  );
}
