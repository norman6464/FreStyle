import { memo } from 'react';
import { UserMinusIcon } from '@heroicons/react/24/outline';
import type { FriendshipUser } from '../types';

interface FriendCardProps {
  user: FriendshipUser;
  onUnfollow: (id: number) => void;
  showUnfollow: boolean;
}

export default memo(function FriendCard({ user, onUnfollow, showUnfollow }: FriendCardProps) {
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
        {user.status && (
          <p className="text-xs text-emerald-400 truncate mt-0.5">{user.status}</p>
        )}
        {user.bio && (
          <p className="text-xs text-[var(--color-text-muted)] truncate mt-0.5">{user.bio}</p>
        )}
      </div>
      {showUnfollow && (
        <button
          onClick={() => onUnfollow(user.userId)}
          className="p-1.5 rounded-md text-[var(--color-text-muted)] hover:bg-red-900/30 hover:text-red-400 transition-colors"
          title="フォロー解除"
          aria-label="フォロー解除"
        >
          <UserMinusIcon className="w-4 h-4" />
        </button>
      )}
    </div>
  );
});
