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

  const accessToken = useSelector((state) => state.auth.accessToken);
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
    debounceSearch(searchQuery);
    return () => debounceSearch.cancel();
  }, [searchQuery, debounceSearch]);

  useEffect(() => {
    if (!accessToken) {
      console.warn('アクセストークンがありません。ログイン画面へ');
      navigate('/login');
      return;
    }

    const controller = new AbortController();
    const queryParam = debounceQuery
      ? `?query=${encodeURIComponent(debounceQuery)}`
      : '';

    const fetchWithAuth = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/chat/users${queryParam}`, {
          credentials: 'include', // Cookie（リフレッシュトークン）送信
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${accessToken}`,
          },
          signal: controller.signal,
        });

        if (res.status === 401) {
          console.log('アクセストークン期限切れ。再発行試行');
          const refreshRes = await fetch(
            `${API_BASE_URL}/api/auth/cognito/refresh-token`,
            {
              method: 'POST',
              credentials: 'include', // Cookieを送る
            }
          );

          if (!refreshRes.ok) {
            console.log('リフレッシュ失敗。再ログインへ');
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

          // Redux + localStorage 更新
          dispatch(setAuthData({ accessToken: newAccessToken }));

          console.log('アクセストークン更新済み。再試行します。');

          // 再試行
          const retryRes = await fetch(
            `${API_BASE_URL}/api/chat/users${queryParam}`,
            {
              credentials: 'include',
              headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${newAccessToken}`,
              },
              signal: controller.signal,
            }
          );

          if (!retryRes.ok) {
            throw new Error('再試行でも失敗しました。');
          }

          const retryData = await retryRes.json();
          setUsers(retryData.users || []);
          return;
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
  }, [debounceQuery, navigate, accessToken, email, dispatch, API_BASE_URL]);

  return (
    <>
      <HamburgerMenu title="ユーザーチャット" />
      
      {/* 修正点: sm:max-w-xl と sm:mx-auto で中央寄せと幅制限を適用 */}
      <div className="min-h-screen bg-gray-100 p-4 pt-16 sm:max-w-xl sm:mx-auto">
        <h2 className="text-2xl font-bold mb-4 text-gray-800">ユーザー追加</h2>

        <div className="mb-4">
          <SearchBox
            value={searchQuery}
            onChange={setSearchQuery}
            placeholder="名前で検索"
          />
        </div>
        
        {error ? (
          <div className="text-red-500 bg-red-100 p-4 rounded-lg border border-red-300">
            エラー: {error}
          </div>
        ) : (
          <MemberList users={users} />
        )}
      </div>
    </>
  );
}