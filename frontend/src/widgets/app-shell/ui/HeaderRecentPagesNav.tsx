import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { ChevronDownIcon, DocumentTextIcon } from '@heroicons/react/24/outline';
import { useClickOutside } from '@/shared/lib/hooks/useClickOutside';
import { KbRepository, type KbRecentPage } from '@/entities/kb';

export interface HeaderRecentPagesNavProps {
  /** ナビリンクと揃える見た目のクラス（Header.tsx の navLinkClass）。 */
  className: string;
  /** モバイルメニュー内では縦積みにする（display:block）。 */
  block?: boolean;
}

/**
 * HeaderRecentPagesNav はヘッダーの「最近見たページ ▾」（段2・段3）。
 *
 * ワークスペースをまたぐ自分の閲覧履歴（GET /kb/me/recent-pages）なので、
 * 「今いるスペース」が分からなくても常にドロップダウンとして使える
 * （HeaderSpacesNav と違い、素のリンクに倒す分岐は無い）。開くたびに取得する
 * （キャッシュしない — 直前に別ページを開いた分をすぐ反映したいため）。
 */
export default function HeaderRecentPagesNav({ className, block = false }: HeaderRecentPagesNavProps) {
  const [open, setOpen] = useState(false);
  const [pages, setPages] = useState<KbRecentPage[] | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  useClickOutside(containerRef, open, () => setOpen(false));

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    setPages(null);
    KbRepository.fetchRecentPages()
      .then((list) => {
        if (!cancelled) setPages(list);
      })
      .catch(() => {
        if (!cancelled) setPages([]);
      });
    return () => {
      cancelled = true;
    };
  }, [open]);

  return (
    <div className="relative" ref={containerRef}>
      <button
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        aria-expanded={open}
        className={`${className} ${block ? 'flex w-full items-center justify-between' : 'flex items-center gap-1'}`}
      >
        最近見たページ
        <ChevronDownIcon className="h-3.5 w-3.5" aria-hidden="true" />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-20 mt-1 w-64 rounded-lg border border-surface-3 bg-surface-1 py-1 shadow-lg">
          {pages === null && (
            <p className="px-3 py-1.5 text-xs text-[var(--color-text-muted)]">読み込み中…</p>
          )}
          {pages?.length === 0 && (
            <p className="px-3 py-1.5 text-xs text-[var(--color-text-muted)]">まだ最近見たページはありません</p>
          )}
          {pages?.map((p) => (
            <Link
              key={p.pageId}
              to={`/kb/${p.pageId}`}
              onClick={() => setOpen(false)}
              className="flex items-center gap-2 px-3 py-1.5 text-sm text-[var(--color-text-secondary)] hover:bg-surface-2"
            >
              {p.icon?.type === 'emoji' ? (
                <span className="shrink-0" aria-hidden="true">{p.icon.value}</span>
              ) : (
                <DocumentTextIcon className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
              )}
              <span className="min-w-0 flex-1 truncate">{p.title}</span>
              <span className="shrink-0 truncate text-xs text-[var(--color-text-muted)]">{p.spaceName}</span>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
