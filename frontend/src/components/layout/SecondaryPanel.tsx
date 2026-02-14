import { XMarkIcon } from '@heroicons/react/24/outline';

interface SecondaryPanelProps {
  title: string;
  headerContent?: React.ReactNode;
  children: React.ReactNode;
  mobileOpen?: boolean;
  onMobileClose?: () => void;
}

export default function SecondaryPanel({ title, headerContent, children, mobileOpen = false, onMobileClose }: SecondaryPanelProps) {
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
        className={`fixed inset-y-0 left-0 z-50 w-72 bg-surface-1 border-r border-surface-3 flex flex-col transform transition-transform duration-200 md:hidden ${
          mobileOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <div className="px-4 py-3 border-b border-surface-3 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">{title}</h2>
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

      {/* デスクトップパネル */}
      <div className="hidden md:flex w-72 bg-surface-1 border-r border-surface-3 flex-col h-full flex-shrink-0">
        <div className="px-4 py-3 border-b border-surface-3">
          <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">{title}</h2>
          {headerContent && <div className="mt-2">{headerContent}</div>}
        </div>
        <div className="flex-1 overflow-y-auto">{children}</div>
      </div>
    </>
  );
}
