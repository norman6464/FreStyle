import { useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import { debounce } from 'lodash';
import { useNavigate } from 'react-router-dom';
import MemberList from '../components/MemberList';
import SearchBox from '../components/SearchBox';

// 今回の場合は検索ボックスを使っているのでloadashライブラリのdebounceでユーザーが入力を終えたらリクエストを送るようにする
export default function MemberPage() {
  const token = useSelector((state) => state.auth.accessToken);
  const navigate = useNavigate();
  const [members, setMembers] = useState([]);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [debounceQuery, setDebounceQuery] = useState('');

  // useMemo（関数自体をメモ化）
  // useMemoでdebounce関数をメモ化して毎回作らないようにする
  const debounceSearch = useMemo(
    // debounceメソッドで500ミリ秒後二検索がかかる
    () => debounce((query) => setDebounceQuery(query), 500),
    []
  );

  // searchQueryが変わったとき、debounceQueryを更新
  useEffect(() => {
    debounceSearch(searchQuery);
    return () => debounceSearch.cancel(); // クリーンアップでキャンセル
  }, [searchQuery, debounceSearch]);

  useEffect(() => {
    if (!token) {
      navigate('/login');
      return;
    }

    const controller = new AbortController();

    const queryParam = debounceQuery
      ? `?query=${encodeURIComponent(debounceQuery)}`
      : '';

    fetch(`http://localhost:8080/api/chat/members${queryParam}`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
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
        if (data?.name) {
          setMembers(data.name);
        } else if (data?.error) {
          setError(data.error);
        }
      })
      .catch((err) => {
        if (err.name !== 'AbortError') {
          setError(err.message);
        }
      });

    return () => controller.abort();
  }, [token, debounceQuery, navigate]);

  return (
    <div className="min-h-screen bg-gray-100 p-4">
      <h1 className="flex justify-center text-2xl font-bold mb-4">友達一覧</h1>

      <div className="mb-4">
        <SearchBox
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder="名前で検索"
        />
      </div>

      {error ? (
        <div className="text-red-500">{error}</div>
      ) : (
        <MemberList members={members} />
      )}
    </div>
  );
}
