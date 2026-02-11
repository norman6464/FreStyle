import { useEffect, useMemo, useState } from 'react';
import { debounce } from 'lodash';
import { useNavigate } from 'react-router-dom';
import MemberList from '../components/MemberList';
import SearchBox from '../components/SearchBox';
import type { MemberUser } from '../types';

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

    fetch(`${API_BASE_URL}/api/chat/users${queryParam}`, {
      headers: { 'Content-Type': 'application/json' },
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
          setMembers(data.users);
        } else if (data?.error) {
          setError(data.error);
        }
      })
      .catch((err: Error) => {
        if (err.name !== 'AbortError') {
          setError(err.message);
        }
      });

    return () => controller.abort();
  }, [debounceQuery, navigate, API_BASE_URL]);

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* ヘッダー */}
      <div className="mb-4">
        <h2 className="text-lg font-bold text-slate-800">チャットメンバー</h2>
        <p className="text-sm text-slate-500">メンバーを検索または選択</p>
      </div>

      {/* 検索ボックス */}
      <div className="mb-6">
        <SearchBox
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder="メンバーを検索..."
        />
      </div>

      {/* エラー表示 */}
      {error && (
        <div className="mb-4 p-3 bg-rose-50 border border-rose-200 rounded-lg">
          <p className="text-sm text-rose-700">{error}</p>
        </div>
      )}

      {/* メンバーリスト */}
      {members.length === 0 && !error && (
        <div className="text-center py-12">
          <p className="text-sm text-slate-500">メンバーがまだいません</p>
        </div>
      )}
      <MemberList users={members} />
    </div>
  );
}
