import { XMarkIcon, ChevronDoubleLeftIcon, ChevronDoubleRightIcon } from '@heroicons/react/24/outline';

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

      {/* デスクトップパネル: 右側の縦罫線は背景色の差分で十分なので border-r は付けない */}
      {collapsible && collapsed ? (
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
          <div className="px-4 py-3 border-b border-surface-3">
            <div className="flex items-center justify-between gap-2">
              <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">
                {title}
                {badge && <span className="ml-2 text-xs font-normal text-[var(--color-text-muted)]">{badge}</span>}
              </h2>
              {collapsible && (
                <button
                  onClick={onToggleCollapsed}
                  title="パネルを折りたたむ"
                  aria-label="パネルを折りたたむ"
                  className="p-1 rounded-md text-[var(--color-text-muted)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)] transition-colors flex-shrink-0"
                >
                  <ChevronDoubleLeftIcon className="w-4 h-4" />
                </button>
              )}
            </div>
            {headerContent && <div className="mt-2">{headerContent}</div>}
          </div>
          <div className="flex-1 overflow-y-auto">{children}</div>
        </div>
      )}
    </>
  );
}
