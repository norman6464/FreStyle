import { useState, useRef, useEffect, useId, ReactNode } from 'react';
import { QuestionMarkCircleIcon } from '@heroicons/react/24/outline';

type HelpTooltipProps = {
  /** ツールチップに表示する説明文。ReactNode を許容することで改行や強調も可能 */
  children: ReactNode;
  /** アクセシブルなラベル（スクリーンリーダー向け）。例: 「5軸評価について」 */
  label?: string;
  /** 吹き出しの表示方向 */
  placement?: 'top' | 'bottom' | 'right' | 'left';
  /** 追加 className */
  className?: string;
};

/**
 * 専門用語や機能の横に置くヘルプアイコン。
 * クリック / フォーカスで説明を表示し、初心者が意味を確認できるようにする。
 */
export default function HelpTooltip({
  children,
  label = '詳細を表示',
  placement = 'top',
  className = '',
}: HelpTooltipProps) {
  const [open, setOpen] = useState(false);
  const wrapperRef = useRef<HTMLSpanElement>(null);
  const tooltipId = useId();

  useEffect(() => {
    if (!open) return;
    const handleClickOutside = (event: MouseEvent) => {
      if (wrapperRef.current && !wrapperRef.current.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpen(false);
    };
    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleEscape);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
    };
  }, [open]);

  const placementClasses: Record<NonNullable<HelpTooltipProps['placement']>, string> = {
    top: 'bottom-full mb-2 left-1/2 -translate-x-1/2',
    bottom: 'top-full mt-2 left-1/2 -translate-x-1/2',
    right: 'left-full ml-2 top-1/2 -translate-y-1/2',
    left: 'right-full mr-2 top-1/2 -translate-y-1/2',
  };

  return (
    <span ref={wrapperRef} className={`relative inline-flex items-center ${className}`}>
      <button
        type="button"
        aria-label={label}
        aria-expanded={open}
        aria-describedby={open ? tooltipId : undefined}
        onClick={() => setOpen((prev) => !prev)}
        onFocus={() => setOpen(true)}
        onBlur={(e) => {
          if (!wrapperRef.current?.contains(e.relatedTarget as Node)) setOpen(false);
        }}
        className="inline-flex items-center justify-center rounded-full text-[var(--color-text-tertiary)] hover:text-primary-300 focus:outline-none focus:ring-2 focus:ring-primary-400 focus:ring-offset-1 focus:ring-offset-[var(--color-surface-1)] transition-colors"
      >
        <QuestionMarkCircleIcon className="w-4 h-4" aria-hidden="true" />
      </button>
      {open && (
        <span
          id={tooltipId}
          role="tooltip"
          className={`absolute z-30 w-60 rounded-lg border border-surface-3 bg-surface-2 px-3 py-2 text-xs text-[var(--color-text-secondary)] shadow-lg animate-fade-in ${placementClasses[placement]}`}
        >
          {children}
        </span>
      )}
    </span>
  );
}
