import { useState } from 'react';
import { useFriendship } from '../hooks/useFriendship';
import EmptyState from '../components/EmptyState';
import Loading from '../components/Loading';
import { UsersIcon, UserMinusIcon } from '@heroicons/react/24/outline';
import type { FriendshipUser } from '../types';

function FriendCard({ user, onUnfollow, showUnfollow }: { user: FriendshipUser; onUnfollow: (id: number) => void; showUnfollow: boolean }) {
  return (
    <div className="flex items-center gap-3 bg-surface-1 rounded-lg border border-surface-3 p-3">
      <div className="w-10 h-10 rounded-full bg-surface-3 flex items-center justify-center flex-shrink-0 overflow-hidden">
        {user.iconUrl ? (
          <img src={user.iconUrl} alt={user.username} className="w-full h-full object-cover" />
        ) : (
          <span className="text-sm font-bold text-[var(--color-text-muted)]">
            {user.username.charAt(0)}
          </span>
        )}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-[var(--color-text-primary)] truncate">
            {user.username}
          </span>
          {user.mutual && (
            <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-400 font-medium">
              相互
            </span>
          )}
        </div>
        {user.bio && (
          <p className="text-xs text-[var(--color-text-muted)] truncate mt-0.5">{user.bio}</p>
        )}
      </div>
      {showUnfollow && (
        <button
          onClick={() => onUnfollow(user.userId)}
          className="p-1.5 rounded-md text-[var(--color-text-muted)] hover:bg-red-900/30 hover:text-red-400 transition-colors"
          title="フォロー解除"
        >
          <UserMinusIcon className="w-4 h-4" />
        </button>
      )}
    </div>
  );
}

export default function FriendshipPage() {
  const { following, followers, loading, unfollowUser } = useFriendship();
  const [tab, setTab] = useState<'following' | 'followers'>('following');

  if (loading) return <Loading />;

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
          message={tab === 'following' ? 'まだ誰もフォローしていません' : 'まだフォロワーがいません'}
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
