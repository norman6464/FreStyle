import type { ReactNode } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { ChevronDoubleLeftIcon, ChevronDoubleRightIcon, Cog6ToothIcon } from '@heroicons/react/24/outline';
import { useAppSelector } from '@/shared/lib/store';
import { navActive, visibleAdminSubs, visibleMainNav } from '../model/navigation';
import type { SidebarMode } from '../model/useSidebarMode';

/** SidebarIconButton はツールチップ付きの小さな操作ボタン（固定 / 閉じる）。 */
function SidebarIconButton({
  label,
  shortcut,
  onClick,
  align = 'left',
  children,
}: {
  label: string;
  shortcut?: string;
  onClick: () => void;
  align?: 'left' | 'right';
  children: ReactNode;
}) {
  return (
    <span className="relative group/tip inline-flex">
      <button
        type="button"
        onClick={onClick}
        aria-label={label}
        className="inline-flex h-7 w-7 items-center justify-center rounded-md text-[var(--color-text-tertiary)] transition-colors hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)]"
      >
        {children}
      </button>
      {/* ホバー / フォーカス時だけ出る説明。位置は端で見切れないよう左右を選べる。 */}
      <span
        role="tooltip"
        className={`pointer-events-none absolute top-full z-50 mt-1.5 hidden whitespace-pre rounded-md bg-[var(--color-text-primary)] px-2 py-1.5 text-xs font-medium leading-tight text-[var(--color-surface-1)] shadow-lg group-hover/tip:block group-focus-within/tip:block ${
          align === 'right' ? 'right-0' : 'left-0'
        }`}
      >
        {shortcut ? `${label}\n${shortcut}` : label}
      </span>
    </span>
  );
}

export interface NavSidebarProps {
  mode: SidebarMode;
  onPin: () => void;
  onCollapse: () => void;
  /** 一時表示中にポインタが入った / 離れたことを親へ伝える。 */
  onPointerEnter?: () => void;
  onPointerLeave?: () => void;
}

/**
 * NavSidebar はアプリの主要ナビを縦に並べる左サイドバー。
 *
 * 一時表示（collapsed）ではオーバーレイとして浮かび、»（固定表示する）で固定へ。
 * 固定表示（pinned）ではレイアウトに居座り、«（閉じる）で一時表示へ戻る。
 * ナビ項目は widgets/app-shell/model/navigation の正典から描画する。
 */
export default function NavSidebar({
  mode,
  onPin,
  onCollapse,
  onPointerEnter,
  onPointerLeave,
}: NavSidebarProps) {
  const { pathname } = useLocation();
  const role = useAppSelector((s) => s.auth.role);
  const isAdmin = useAppSelector((s) => s.auth.isAdmin);
  const aiChatEnabledForTrainees = useAppSelector((s) => s.auth.aiChatEnabledForTrainees);

  const mainNav = visibleMainNav(role, { aiChatEnabledForTrainees });
  const adminSubs = isAdmin ? visibleAdminSubs(role) : [];

  const itemClass = (active: boolean) =>
    `flex items-center gap-2 rounded-md px-2.5 py-1.5 text-sm transition-colors ${
      active
        ? 'bg-[var(--color-nav-active)] font-medium text-[var(--color-text-primary)]'
        : 'text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)]'
    }`;

  return (
    <nav
      aria-label="サイドナビゲーション"
      onMouseEnter={onPointerEnter}
      onMouseLeave={onPointerLeave}
      className="flex h-full w-60 flex-col bg-[var(--color-nav)]"
    >
      <div className="flex h-12 flex-shrink-0 items-center gap-2 px-3">
        <Link to="/dashboard" className="flex min-w-0 flex-1 items-center gap-2" aria-label="FreStyle ホーム">
          <img src="/favicon.svg" alt="" aria-hidden="true" className="h-6 w-6 flex-shrink-0" />
          <span className="truncate text-sm font-semibold text-[var(--color-text-primary)]">FreStyle</span>
        </Link>
        {mode === 'pinned' ? (
          <SidebarIconButton label="サイドバーを閉じる" shortcut="⌘\" onClick={onCollapse} align="right">
            <ChevronDoubleLeftIcon className="h-4 w-4" />
          </SidebarIconButton>
        ) : (
          <SidebarIconButton label="サイドバーを固定表示する" shortcut="⌘\" onClick={onPin} align="right">
            <ChevronDoubleRightIcon className="h-4 w-4" />
          </SidebarIconButton>
        )}
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto px-2 pb-3">
        <ul className="space-y-0.5">
          {mainNav.map((item) => (
            <li key={item.id}>
              <Link
                to={item.to}
                className={itemClass(navActive(item, pathname))}
                aria-current={navActive(item, pathname) ? 'page' : undefined}
              >
                {item.label}
              </Link>
            </li>
          ))}
        </ul>

        {adminSubs.length > 0 && (
          <>
            <p className="px-2.5 pb-1 pt-4 text-xs text-[var(--color-text-muted)]">管理</p>
            <ul className="space-y-0.5">
              {adminSubs.map((sub) => (
                <li key={sub.to}>
                  <Link
                    to={sub.to}
                    className={itemClass(pathname.startsWith(sub.matchPrefix))}
                    aria-current={pathname.startsWith(sub.matchPrefix) ? 'page' : undefined}
                  >
                    {sub.label}
                  </Link>
                </li>
              ))}
            </ul>
          </>
        )}
      </div>

      <div className="flex-shrink-0 border-t border-surface-3 p-2">
        <Link to="/settings" className={itemClass(pathname === '/settings')}>
          <Cog6ToothIcon className="h-4 w-4" />
          設定
        </Link>
      </div>
    </nav>
  );
}
