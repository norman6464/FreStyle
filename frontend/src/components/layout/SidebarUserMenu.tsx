import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Cog6ToothIcon,
  ArrowLeftOnRectangleIcon,
  ChevronUpDownIcon,
} from '@heroicons/react/24/outline';
import Avatar from '../Avatar';

interface SidebarUserMenuProps {
  expanded: boolean;
  displayName: string;
  avatarUrl?: string | null;
  /** 任意のサブテキスト（例: ロール名）。 collapsed 時は非表示 */
  subText?: string | null;
  onLogout: () => void;
  onNavigate?: () => void;
}

/**
 * SidebarUserMenu — Sidebar 最下部に置く「ユーザーボタン」 + クリックでドロップダウン。
 *
 * 展開時: アバター + 名前 + サブテキスト + ChevronUpDown
 * 折りたたみ時: アバターのみ
 *
 * クリックで上方向に menu が開く（ボタンが下端なので逆さに表示）。
 * メニュー項目: 設定 (`/settings`) / ログアウト
 */
export default function SidebarUserMenu({
  expanded,
  displayName,
  avatarUrl,
  subText,
  onLogout,
  onNavigate,
}: SidebarUserMenuProps) {
  const [open, setOpen] = useState(false);
  const wrapRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  // クリックアウトサイドで閉じる。 button 自体のクリックは含めず逆判定する。
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
        title={!expanded ? displayName : undefined}
        aria-haspopup="menu"
        aria-expanded={open}
        className={`w-full flex items-center gap-2 px-2 py-2 rounded-md hover:bg-[var(--color-nav-hover)] transition-colors ${
          expanded ? 'justify-between' : 'justify-center'
        }`}
      >
        <div className={`flex items-center gap-2 min-w-0 ${expanded ? '' : 'justify-center'}`}>
          <Avatar name={displayName || 'U'} src={avatarUrl ?? undefined} size="sm" />
          {expanded && (
            <div className="text-left min-w-0">
              <p className="text-sm font-medium text-[var(--color-text-primary)] truncate">
                {displayName || 'ユーザー'}
              </p>
              {subText && (
                <p className="text-xs text-[var(--color-text-muted)] truncate">{subText}</p>
              )}
            </div>
          )}
        </div>
        {expanded && <ChevronUpDownIcon className="w-4 h-4 text-[var(--color-text-muted)] flex-shrink-0" />}
      </button>

      {open && (
        <div
          role="menu"
          className="absolute bottom-full left-0 right-0 mb-2 bg-surface-1 border border-surface-3 rounded-lg shadow-lg overflow-hidden z-50 animate-fade-in"
        >
          <button
            type="button"
            role="menuitem"
            onClick={handleSettings}
            className="w-full flex items-center gap-2 px-3 py-2 text-sm text-[var(--color-text-primary)] hover:bg-[var(--color-nav-hover)] transition-colors"
          >
            <Cog6ToothIcon className="w-4 h-4" />
            設定
          </button>
          <div className="border-t border-surface-3" />
          <button
            type="button"
            role="menuitem"
            onClick={handleLogout}
            className="w-full flex items-center gap-2 px-3 py-2 text-sm text-[var(--color-text-muted)] hover:bg-red-900/10 hover:text-red-500 transition-colors"
          >
            <ArrowLeftOnRectangleIcon className="w-4 h-4" />
            ログアウト
          </button>
        </div>
      )}
    </div>
  );
}
