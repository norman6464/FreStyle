import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Cog6ToothIcon,
  ArrowLeftOnRectangleIcon,
  ChevronDownIcon,
} from '@heroicons/react/24/outline';
import Avatar from '@/shared/ui/Avatar';

interface HeaderUserMenuProps {
  displayName: string;
  avatarUrl?: string | null;
  email?: string;
  /** 任意のサブテキスト（例: ロール名）。 */
  subText?: string | null;
  onLogout: () => void;
  onNavigate?: () => void;
}

/**
 * HeaderUserMenu — ヘッダー右端の「ユーザーボタン」 + クリックで下方向ドロップダウン。
 *
 * メニュー項目: 設定 (`/settings`) / ログアウト。
 * サイドバー版 (SidebarUserMenu) と違い、 ヘッダー上端に置くため下方向に開く。
 */
export default function HeaderUserMenu({
  displayName,
  avatarUrl,
  email,
  subText,
  onLogout,
  onNavigate,
}: HeaderUserMenuProps) {
  const [open, setOpen] = useState(false);
  const wrapRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  // クリックアウトサイドで閉じる（ボタン自身のクリックは含めない）。
  useEffect(() => {
    if (!open) return;
    const onDoc = (e: MouseEvent) => {
      if (!wrapRef.current) return;
      if (e.target instanceof Node && wrapRef.current.contains(e.target)) return;
      setOpen(false);
    };
    document.addEventListener('mousedown', onDoc);
    return () => document.removeEventListener('mousedown', onDoc);
  }, [open]);

  const handleSettings = () => {
    setOpen(false);
    onNavigate?.();
    navigate('/settings');
  };

  const handleLogout = () => {
    setOpen(false);
    onLogout();
  };

  return (
    <div ref={wrapRef} className="relative">
      <button
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        aria-haspopup="true"
        aria-expanded={open}
        className="flex items-center gap-2 px-2 py-1.5 rounded-md hover:bg-[var(--color-nav-hover)] transition-colors"
      >
        <Avatar name={displayName || 'U'} src={avatarUrl ?? undefined} size="sm" />
        <span className="hidden sm:block text-sm font-medium text-[var(--color-text-primary)] max-w-[10rem] truncate">
          {displayName || 'ユーザー'}
        </span>
        <ChevronDownIcon className="w-4 h-4 text-[var(--color-text-muted)] flex-shrink-0" />
      </button>

      {open && (
        <div className="absolute top-full right-0 mt-2 w-56 bg-surface-1 border border-surface-3 rounded-lg shadow-lg overflow-hidden z-50 animate-fade-in">

          {(email || subText) && (
            <>
              <div className="px-3 py-2">
                {subText && (
                  <p className="text-xs text-[var(--color-text-muted)]">{subText}</p>
                )}
                {email && (
                  <p className="text-xs text-[var(--color-text-secondary)] truncate" title={email}>
                    {email}
                  </p>
                )}
              </div>
              <div className="border-t border-surface-3" />
            </>
          )}
          <button
            type="button"
            onClick={handleSettings}
            className="w-full flex items-center gap-2 px-3 py-2 text-sm text-[var(--color-text-primary)] hover:bg-[var(--color-nav-hover)] transition-colors whitespace-nowrap"
          >
            <Cog6ToothIcon className="w-4 h-4 flex-shrink-0" />
            設定
          </button>
          <div className="border-t border-surface-3" />
          <button
            type="button"
            onClick={handleLogout}
            className="w-full flex items-center gap-2 px-3 py-2 text-sm text-[var(--color-text-muted)] hover:bg-red-900/10 hover:text-red-500 transition-colors whitespace-nowrap"
          >
            <ArrowLeftOnRectangleIcon className="w-4 h-4 flex-shrink-0" />
            ログアウト
          </button>
        </div>
      )}
    </div>
  );
}
