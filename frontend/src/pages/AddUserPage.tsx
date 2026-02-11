import { useEffect, useMemo, useState } from 'react';
import { useDispatch } from 'react-redux';
import { debounce } from 'lodash';
import { useNavigate } from 'react-router-dom';
import MemberList from '../components/MemberList';
import SearchBox from '../components/SearchBox';
import {
  MagnifyingGlassIcon,
  UserPlusIcon,
} from '@heroicons/react/24/solid';
import type { MemberUser } from '../types';

export default function AddUserPage() {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const [users, setUsers] = useState<MemberUser[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [debounceQuery, setDebounceQuery] = useState('');

  const debounceSearch = useMemo(
    () => debounce((query: string) => setDebounceQuery(query), 500),
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

        if (res.status === 401) {
          const refreshRes = await fetch(
            `${API_BASE_URL}/api/auth/cognito/refresh-token`,
            {
              method: 'POST',
              credentials: 'include',
              signal: controller.signal,
            }
          );
          if (!refreshRes.ok) {
            navigate('/login');
            return;
          }
          res = await fetchUsers();
        }

        if (!res.ok) throw new Error('ユーザー取得に失敗しました');

        const data = await res.json();
        setUsers(data.users || []);
      } catch (err) {
        if ((err as Error).name !== 'AbortError') {
          setError((err as Error).message);
        }
      }
    };

    fetchWithAuth();
    return () => controller.abort();
  }, [debounceQuery, navigate, dispatch, API_BASE_URL]);

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* 検索ボックス */}
      <div className="mb-6">
        <SearchBox
          value={searchQuery}
          onChange={(value) => {
            setSearchQuery(value);
            debounceSearch(value);
          }}
          placeholder="ユーザー名またはメールアドレスで検索..."
        />
      </div>

      {/* エラー表示 */}
      {error && (
        <div className="mb-4 p-3 bg-rose-50 border border-rose-200 rounded-lg">
          <p className="text-sm text-rose-700">{error}</p>
        </div>
      )}

      {/* 検索前の状態 */}
      {users.length === 0 && !debounceQuery && (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <div className="bg-slate-100 rounded-full p-4 mb-4">
            <MagnifyingGlassIcon className="w-8 h-8 text-slate-400" />
          </div>
          <h3 className="text-base font-semibold text-slate-700 mb-1">
            ユーザーを検索してみましょう
          </h3>
          <p className="text-sm text-slate-500 max-w-xs">
            名前やメールアドレスを入力して、チャットしたい相手を探してください
          </p>
        </div>
      )}

      {/* 検索結果なし */}
      {users.length === 0 && debounceQuery && (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <h3 className="text-base font-semibold text-slate-700 mb-1">
            ユーザーが見つかりませんでした
          </h3>
          <p className="text-sm text-slate-500 max-w-xs">
            「{debounceQuery}」に一致するユーザーはいません。
          </p>
        </div>
      )}

      {/* ユーザーリスト */}
      {users.length > 0 && (
        <div>
          <div className="flex items-center gap-2 mb-3">
            <UserPlusIcon className="w-4 h-4 text-slate-500" />
            <span className="text-xs font-semibold text-slate-600">
              {users.length}人のユーザーが見つかりました
            </span>
          </div>
          <MemberList users={users} />
        </div>
      )}
    </div>
  );
}
