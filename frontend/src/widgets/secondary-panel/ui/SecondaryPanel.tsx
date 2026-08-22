import type { ReactNode } from 'react';
import {
  XMarkIcon,
  Bars3Icon,
  ChevronDoubleLeftIcon,
  ChevronDoubleRightIcon,
} from '@heroicons/react/24/outline';
import { usePanelMode } from '@/shared/lib/hooks/usePanelMode';

interface SecondaryPanelProps {
  title: string;
  badge?: string;
  headerContent?: React.ReactNode;
  children: React.ReactNode;
  mobileOpen?: boolean;
  onMobileClose?: () => void;
  /** デスクトップでパネルを折りたたみ可能にする（章一覧などで本文幅を稼ぐ用途）。 */
  collapsible?: boolean;
  /** collapsible のとき、 折りたたみ中かどうか。 */
  collapsed?: boolean;
  /** 折りたたみ / 展開のトグル。 */
  onToggleCollapsed?: () => void;
  /**
   * 一時表示モード（デスクトップ）。« で隠すと本文が全幅になり、左端ホバーか ☰ で
   * オーバーレイとして浮いて出る。☰ / » クリックで固定へ戻る。状態は storageKey ごとに保存。
   * collapsible とは独立の新しい機構（指定時は collapsible を無視する）。
   */
  peekable?: boolean;
  /** peekable の表示モードを保存する localStorage キー（画面ごとに分ける）。 */
  storageKey?: string;
}

/** PanelTooltip は «/»/☰ に付けるホバー説明（ラベル＋ショートカット）。 */
function PanelTooltip({ label, align = 'left' }: { label: string; align?: 'left' | 'right' }) {
  return (
    <span
      role="tooltip"
      className={`pointer-events-none absolute top-full z-50 mt-1.5 hidden whitespace-pre rounded-md bg-[var(--color-text-primary)] px-2 py-1.5 text-xs font-medium leading-tight text-[var(--color-surface-1)] shadow-lg group-hover/ptip:block ${
        align === 'right' ? 'right-0' : 'left-0'
      }`}
    >
      {`${label}\n⌘\\`}
    </span>
  );
}

/** PanelHeader はデスクトップパネルの見出し行（タイトル＋バッジ＋トグル）。 */
function PanelHeader({
  title,
  badge,
  headerContent,
  toggle,
}: {
  title: string;
  badge?: string;
  headerContent?: ReactNode;
  toggle?: ReactNode;
}) {
  return (
    <div className="px-4 py-3 border-b border-surface-3">
      <div className="flex items-center justify-between gap-2">
        <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">
          {title}
          {badge && <span className="ml-2 text-xs font-normal text-[var(--color-text-muted)]">{badge}</span>}
        </h2>
        {toggle}
      </div>
      {headerContent && <div className="mt-2">{headerContent}</div>}
    </div>
  );
}

/**
 * PeekablePanel は「一時表示 / 固定表示」を切り替えられるデスクトップパネル。
 *
 * - 固定表示: レイアウトに居座り、«（サイドバーを閉じる）で一時表示モードへ
 * - 一時表示: 本文が全幅になり、左端ホバーか ☰（サイドバーを固定表示する）で浮いて出る。
 *   »（またはもう一度 ☰）で固定へ戻す。モードは storageKey ごとに保存され ⌘\ でも切替可
 */
function PeekablePanel({
  title,
  badge,
  headerContent,
  children,
  storageKey,
}: {
  title: string;
  badge?: string;
  headerContent?: ReactNode;
  children: ReactNode;
  storageKey: string;
}) {
  const panel = usePanelMode(storageKey);

  if (panel.mode === 'pinned') {
    return (
      <div className="hidden md:flex w-72 bg-[var(--color-nav)] flex-col h-full flex-shrink-0">
        <PanelHeader
          title={title}
          badge={badge}
          headerContent={headerContent}
          toggle={
            <span className="relative group/ptip inline-flex flex-shrink-0">
              <button
                onClick={panel.collapse}
                aria-label="サイドバーを閉じる"
                className="p-1 rounded-md text-[var(--color-text-muted)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)] transition-colors"
              >
                <ChevronDoubleLeftIcon className="w-4 h-4" />
              </button>
              <PanelTooltip label="サイドバーを閉じる" align="right" />
            </span>
          }
        />
        <div className="flex-1 overflow-y-auto">{children}</div>
      </div>
    );
  }

  // 一時表示モード: レイアウト上は何も占有しない。左端ホバーゾーン＋☰＋オーバーレイを出す。
  return (
    <>
      {/* 左端のホバー検知ゾーン。 */}
      <div
        aria-hidden="true"
        className="hidden md:block fixed left-0 top-16 bottom-0 z-30 w-2"
        onMouseEnter={panel.openPeek}
        onMouseLeave={panel.closePeek}
      />

      {/* 本文左上の ☰（ホバーで一時表示・クリックで固定）。 */}
      {/* fixed の span 自体がツールチップ(absolute)の基準になるため relative は不要。 */}
      <span
        className={`group/ptip hidden md:inline-flex fixed left-3 top-[72px] z-30 transition-opacity ${
          panel.isPeeking ? 'opacity-0 pointer-events-none' : 'opacity-100'
        }`}
      >
        <button
          onClick={panel.pin}
          onMouseEnter={panel.openPeek}
          onMouseLeave={panel.closePeek}
          aria-label="サイドバーを固定表示する"
          className="p-2 rounded-md bg-[var(--color-surface-1)] border border-surface-3 text-[var(--color-text-tertiary)] shadow-sm hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)] transition-colors"
        >
          <Bars3Icon className="w-4 h-4" />
        </button>
        <PanelTooltip label="サイドバーを固定表示する" />
      </span>

      {/* 一時表示のオーバーレイパネル（本文は動かさない）。 */}
      <div
        onMouseEnter={panel.openPeek}
        onMouseLeave={panel.closePeek}
        className={`hidden md:flex fixed left-0 top-[72px] bottom-2 z-40 w-72 flex-col overflow-hidden rounded-r-xl border border-surface-3 bg-[var(--color-nav)] shadow-xl transition-all duration-200 ease-out ${
          panel.isPeeking ? 'translate-x-0 opacity-100' : '-translate-x-full opacity-0 pointer-events-none'
        }`}
      >
        <PanelHeader
          title={title}
          badge={badge}
          headerContent={headerContent}
          toggle={
            <span className="relative group/ptip inline-flex flex-shrink-0">
              <button
                onClick={panel.pin}
                aria-label="サイドバーを固定表示する"
                className="p-1 rounded-md text-[var(--color-text-muted)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)] transition-colors"
              >
                <ChevronDoubleRightIcon className="w-4 h-4" />
              </button>
              <PanelTooltip label="サイドバーを固定表示する" align="right" />
            </span>
          }
        />
        <div className="flex-1 overflow-y-auto">{children}</div>
      </div>
    </>
  );
}

export default function SecondaryPanel({
  title,
  badge,
  headerContent,
  children,
  mobileOpen = false,
  onMobileClose,
  collapsible = false,
  collapsed = false,
  onToggleCollapsed,
  peekable = false,
  storageKey,
}: SecondaryPanelProps) {
  return (
    <>
      {/* モバイルオーバーレイ */}
      {mobileOpen && (
        <div
          className="fixed inset-0 bg-black/40 z-40 md:hidden"
          onClick={onMobileClose}
        />
      )}

      {/* モバイルパネル */}
      <div
        className={`fixed inset-y-0 left-0 z-50 w-72 bg-[var(--color-nav)] border-r border-surface-3 flex flex-col transform transition-transform duration-200 md:hidden ${
          mobileOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <div className="px-4 py-3 border-b border-surface-3 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">
            {title}
            {badge && <span className="ml-2 text-xs font-normal text-[var(--color-text-muted)]">{badge}</span>}
          </h2>
          <button
            onClick={onMobileClose}
            className="p-1 hover:bg-surface-2 rounded transition-colors"
            aria-label="パネルを閉じる"
          >
            <XMarkIcon className="w-4 h-4 text-[var(--color-text-muted)]" />
          </button>
        </div>
        {headerContent && <div className="px-4 py-2 border-b border-surface-3">{headerContent}</div>}
        <div className="flex-1 overflow-y-auto">{children}</div>
      </div>

      {/* デスクトップ: peekable（一時表示/固定）を優先し、従来の collapsible とは独立に扱う。 */}
      {peekable && storageKey ? (
        <PeekablePanel title={title} badge={badge} headerContent={headerContent} storageKey={storageKey}>
          {children}
        </PeekablePanel>
      ) : collapsible && collapsed ? (
        // 折りたたみ中: 細い帯に「開く」ボタンだけ出す。本文が全幅に広がる。
        <div className="hidden md:flex w-10 bg-[var(--color-nav)] flex-col items-center pt-3 h-full flex-shrink-0">
          <button
            onClick={onToggleCollapsed}
            title="パネルを開く"
            aria-label="パネルを開く"
            className="p-1.5 rounded-md text-[var(--color-text-muted)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)] transition-colors"
          >
            <ChevronDoubleRightIcon className="w-4 h-4" />
          </button>
        </div>
      ) : (
        <div className="hidden md:flex w-72 bg-[var(--color-nav)] flex-col h-full flex-shrink-0">
          <PanelHeader
            title={title}
            badge={badge}
            headerContent={headerContent}
            toggle={
              collapsible ? (
                <button
                  onClick={onToggleCollapsed}
                  title="パネルを折りたたむ"
                  aria-label="パネルを折りたたむ"
                  className="p-1 rounded-md text-[var(--color-text-muted)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)] transition-colors flex-shrink-0"
                >
                  <ChevronDoubleLeftIcon className="w-4 h-4" />
                </button>
              ) : undefined
            }
          />
          <div className="flex-1 overflow-y-auto">{children}</div>
        </div>
      )}
    </>
  );
}
