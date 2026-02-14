import MemberList from '../components/MemberList';
import SearchBox from '../components/SearchBox';
import { useUserSearch } from '../hooks/useUserSearch';

export default function MemberPage() {
  const { users, error, searchQuery, setSearchQuery } = useUserSearch();

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* ヘッダー */}
      <div className="mb-4">
        <h2 className="text-lg font-bold text-[var(--color-text-primary)]">チャットメンバー</h2>
        <p className="text-sm text-[var(--color-text-muted)]">メンバーを検索または選択</p>
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
        <div className="mb-4 p-3 bg-rose-900/30 border border-rose-800 rounded-lg">
          <p className="text-sm text-rose-400">{error}</p>
        </div>
      )}

      {/* メンバーリスト */}
      {users.length === 0 && !error && (
        <div className="text-center py-12">
          <p className="text-sm text-[var(--color-text-muted)]">メンバーがまだいません</p>
        </div>
      )}
      <MemberList users={users} />
    </div>
  );
}
