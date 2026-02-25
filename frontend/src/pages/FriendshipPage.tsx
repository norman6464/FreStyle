import { useState } from 'react';
import { useFriendship } from '../hooks/useFriendship';
import EmptyState from '../components/EmptyState';
import Loading from '../components/Loading';
import FriendCard from '../components/FriendCard';
import { UsersIcon } from '@heroicons/react/24/outline';

export default function FriendshipPage() {
  const { following, followers, loading, unfollowUser } = useFriendship();
  const [tab, setTab] = useState<'following' | 'followers'>('following');

  if (loading) return <Loading message="読み込み中..." className="min-h-[calc(100vh-3.5rem)]" />;

  const currentList = tab === 'following' ? following : followers;

  return (
    <div className="max-w-2xl mx-auto px-4 py-6">
      <div className="flex items-center gap-2 mb-6">
        <UsersIcon className="w-6 h-6 text-[var(--color-text-primary)]" />
        <h1 className="text-lg font-bold text-[var(--color-text-primary)]">フレンド</h1>
      </div>

      <div className="flex gap-1 mb-4 bg-surface-2 rounded-lg p-1">
        <button
          onClick={() => setTab('following')}
          className={`flex-1 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
            tab === 'following'
              ? 'bg-surface-1 text-[var(--color-text-primary)] shadow-sm'
              : 'text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
          }`}
        >
          フォロー中 ({following.length})
        </button>
        <button
          onClick={() => setTab('followers')}
          className={`flex-1 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
            tab === 'followers'
              ? 'bg-surface-1 text-[var(--color-text-primary)] shadow-sm'
              : 'text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
          }`}
        >
          フォロワー ({followers.length})
        </button>
      </div>

      {currentList.length === 0 ? (
        <EmptyState
          icon={UsersIcon}
          title={tab === 'following' ? 'まだ誰もフォローしていません' : 'まだフォロワーがいません'}
        />
      ) : (
        <div className="space-y-2">
          {currentList.map(user => (
            <FriendCard
              key={user.id}
              user={user}
              onUnfollow={unfollowUser}
              showUnfollow={tab === 'following'}
            />
          ))}
        </div>
      )}
    </div>
  );
}
