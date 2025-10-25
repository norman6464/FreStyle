import { useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import { debounce } from 'lodash';
import { useNavigate } from 'react-router-dom';
import MemberList from '../components/MemberList';
import SearchBox from '../components/SearchBox';
import HamburgerMenu from '../components/HamburgerMenu';

export default function AddUserPage() {
  const token = useSelector((state) => state.auth.accessToken);
  const navigate = useNavigate();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  // メールアドレス一覧を取得する
  const [users, setUsers] = useState([]);
  const [error, setError] = useState(null);
  // 検索ボックスに含めるもの
  const [searchQuery, setSearchQuery] = useState('');
  const [debounceQuery, setDebounceQuery] = useState('');

  // debounceで検索入力後にリクエストを送る
  const debounceSearch = useMemo(
    () => debounce((query) => setDebounceQuery(query), 500),
    []
  );

  useEffect(() => {
    debounceSearch(searchQuery);
    return () => debounceSearch.cancel();
  }, [searchQuery, debounceSearch]);

  useEffect(() => {
    if (!token) {
      navigate('/login');
      return;
    }

    // abort（中止）
    const controller = new AbortController();
    const queryParam = debounceQuery
      ? `?query=${encodeURIComponent(debounceQuery)}`
      : '';

    console.log('検索開始');

    fetch(`${API_BASE_URL}/api/chat/users${queryParam}`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      // これは途中でリクエストを送るかどうかを決めるためのシグナル
      signal: controller.signal,
    })
      .then((res) => {
        if (res.status === 401) {
          navigate('/login');
          return;
        }
        return res.json();
      })
      .then((data) => {
        // 今回はメールアドレスで検索をかける
        if (data?.users && data.users.length > 0) {
          console.log(data.users);
          setUsers(data.users);
        } else if (data?.users.length === 0) {
          console.log('ログ');
        }
      })
      .catch((err) => {
        if (err.name !== 'AbortError') {
          setError(err.message);
        }
      });

    // リクエスト中止
    return () => controller.abort();
  }, [token, debounceQuery, navigate]);

  return (
    <>
      <HamburgerMenu />
      <div className="min-h-screen bg-gray-100 p-4 mt-16">
        <h2 className="text-xl font-semibold mb-4">ユーザー追加</h2>

        <div className="mb-4">
          <SearchBox
            value={searchQuery}
            onChange={setSearchQuery}
            placeholder="名前で検索"
          />
        </div>
        {/* {error ? (
          <div className="text-red-500">{error}</div>
        ) : (
          <MemberList members={users} />
        )} */}
        <MemberList users={users} token={token} />
      </div>
    </>
  );
}
