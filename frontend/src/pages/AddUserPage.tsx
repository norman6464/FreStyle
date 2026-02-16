import MemberList from '../components/MemberList';
import SearchBox from '../components/SearchBox';
import FormMessage from '../components/FormMessage';
import {
  MagnifyingGlassIcon,
  UserPlusIcon,
} from '@heroicons/react/24/solid';
import { useUserSearch } from '../hooks/useUserSearch';

export default function AddUserPage() {
  const { users, error, searchQuery, setSearchQuery, debounceQuery, loading } = useUserSearch();

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* 検索ボックス */}
      <div className="mb-6">
        <SearchBox
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder="ユーザー名またはメールアドレスで検索..."
        />
      </div>

      {/* エラー表示 */}
      <FormMessage message={error ? { type: 'error', text: error } : null} />

      {/* ローディング中 */}
      {loading && (
        <div className="flex items-center justify-center py-12">
          <p className="text-sm text-[var(--color-text-muted)]">検索中...</p>
        </div>
      )}

      {/* 検索前の状態 */}
      {!loading && users.length === 0 && !debounceQuery && (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <div className="bg-surface-3 rounded-full p-4 mb-4">
            <MagnifyingGlassIcon className="w-8 h-8 text-[var(--color-text-faint)]" />
          </div>
          <h3 className="text-base font-semibold text-[var(--color-text-secondary)] mb-1">
            ユーザーを検索してみましょう
          </h3>
          <p className="text-sm text-[var(--color-text-muted)] max-w-xs">
            名前やメールアドレスを入力して、チャットしたい相手を探してください
          </p>
        </div>
      )}

      {/* 検索結果なし */}
      {!loading && users.length === 0 && debounceQuery && (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <h3 className="text-base font-semibold text-[var(--color-text-secondary)] mb-1">
            ユーザーが見つかりませんでした
          </h3>
          <p className="text-sm text-[var(--color-text-muted)] max-w-xs">
            「{debounceQuery}」に一致するユーザーはいません。
          </p>
        </div>
      )}

      {/* ユーザーリスト */}
      {!loading && users.length > 0 && (
        <div>
          <div className="flex items-center gap-2 mb-3">
            <UserPlusIcon className="w-4 h-4 text-[var(--color-text-muted)]" />
            <span className="text-xs font-semibold text-[var(--color-text-tertiary)]">
              {users.length}人のユーザーが見つかりました
            </span>
          </div>
          <MemberList users={users} />
        </div>
      )}
    </div>
  );
}
