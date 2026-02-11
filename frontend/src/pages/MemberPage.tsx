import { useEffect, useMemo, useState } from 'react';
import { debounce } from 'lodash';
import { useNavigate } from 'react-router-dom';
import MemberList from '../components/MemberList';
import SearchBox from '../components/SearchBox';
import HamburgerMenu from '../components/HamburgerMenu';
import type { MemberUser } from '../types';

// 今回の場合は検索ボックスを使っているのでloadashライブラリのdebounceでユーザーが入力を終えたらリクエストを送るようにする
export default function MemberPage() {
  const navigate = useNavigate();
  const [members, setMembers] = useState<MemberUser[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [debounceQuery, setDebounceQuery] = useState('');
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const debounceSearch = useMemo(
    () => debounce((query: string) => setDebounceQuery(query), 500),
    []
  );

  useEffect(() => {
    debounceSearch(searchQuery);
    return () => debounceSearch.cancel();
  }, [searchQuery, debounceSearch]);

  useEffect(() => {
    const controller = new AbortController();

    const queryParam = debounceQuery
      ? `?query=${encodeURIComponent(debounceQuery)}`
      : '';

    console.log('リクエスト開始');

    fetch(`${API_BASE_URL}/api/chat/users${queryParam}`, {
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      signal: controller.signal,
    })
      .then((res) => {
        if (res.status === 401) {
          navigate('/login');
          return;
        } else if (res.status === 500) {
          navigate('/');
          return;
        }
        return res.json();
      })
      .then((data) => {
        if (data?.users) {
          console.log('ユーザー一覧： ', data);
          setMembers(data.users);
        } else if (data?.error) {
          setError(data.error);
        }
      })
      .catch((err: Error) => {
        console.log('error fetch /api/chat/members');
        if (err.name !== 'AbortError') {
          setError(err.message);
        }
      });

    return () => controller.abort();
  }, [debounceQuery, navigate, API_BASE_URL]);

  return (
    <>
      <HamburgerMenu title="チャットメンバー" />
      <div className="min-h-screen bg-gray-50 p-4 pt-20 pb-8">
        <div className="max-w-2xl mx-auto">
          {/* ヘッダー */}
          <div className="mb-8">
            <h2 className="text-4xl font-bold text-gray-800 mb-2">
              あなたのチャット
            </h2>
            <p className="text-gray-600 text-lg">メンバーを検索または選択</p>
          </div>

          {/* 検索ボックス */}
          <div className="mb-8">
            <SearchBox
              value={searchQuery}
              onChange={setSearchQuery}
              placeholder="メンバーを検索..."
            />
          </div>

          {/* エラー表示 */}
          {error && (
            <div className="mb-6 p-4 bg-red-50 border-l-4 border-red-500 rounded-lg">
              <p className="text-red-700 font-semibold">{error}</p>
            </div>
          )}

          {/* メンバーリスト */}
          {members.length === 0 && !error && (
            <div className="text-center py-12">
              <div className="w-16 h-16 bg-gray-200 rounded-full mx-auto mb-4 flex items-center justify-center">
                <svg
                  className="w-8 h-8 text-gray-500"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M17 20h5v-2a3 3 0 00-5.856-1.487M15 10a3 3 0 11-6 0 3 3 0 016 0zM15 20H9m6 0h.01"
                  />
                </svg>
              </div>
              <p className="text-gray-500 text-lg">メンバーがまだいません</p>
            </div>
          )}
          <MemberList users={members} />
        </div>
      </div>
    </>
  );
}
