import { useEffect, useRef, useState } from 'react';
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/24/outline';
import type { KbWorkspace } from '@/entities/knowledge-base';

export interface KbWorkspaceSwitcherProps {
  workspaces: KbWorkspace[];
  activeSlug: string | null;
  onSelect: (slug: string) => void;
}

/**
 * KbWorkspaceSwitcher は最上段のワークスペース切替。
 *
 * 切替（同時に 1 つ）にしてあるのは、ワークスペースが**会社の境界**だから。
 * 同時に 2 社ぶんを見る場面が無く、並べると「いまどちらを触っているか」が曖昧になる。
 * 逆にスペースは同時に見たいので、あちらは見出しとして並べてある。
 */
export default function KbWorkspaceSwitcher({
  workspaces,
  activeSlug,
  onSelect,
}: KbWorkspaceSwitcherProps) {
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const active = workspaces.find((w) => w.slug === activeSlug) ?? null;

  // 外を押したら閉じる。
  useEffect(() => {
    if (!open) return;
    const onDocumentMouseDown = (event: MouseEvent) => {
      if (
        containerRef.current &&
        event.target instanceof Node &&
        containerRef.current.contains(event.target)
      ) {
        return;
      }
      setOpen(false);
    };
    document.addEventListener('mousedown', onDocumentMouseDown);
    return () => document.removeEventListener('mousedown', onDocumentMouseDown);
  }, [open]);

  // 所属が 1 つしか無いなら切り替える先が無いので、押せる見た目にしない。
  if (workspaces.length <= 1) {
    return (
      <p className="truncate px-2 py-2 text-sm font-semibold text-[var(--color-text-primary)]">
        {active?.name ?? 'ワークスペースがありません'}
      </p>
    );
  }

  return (
    <div ref={containerRef} className="relative">
      <button
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        aria-haspopup="menu"
        aria-expanded={open}
        className="flex w-full items-center gap-1 rounded-md px-2 py-2 text-left hover:bg-surface-2"
      >
        <span className="min-w-0 flex-1 truncate text-sm font-semibold text-[var(--color-text-primary)]">
          {active?.name ?? 'ワークスペースを選択'}
        </span>
        <ChevronUpDownIcon className="h-4 w-4 shrink-0 text-[var(--color-text-muted)]" aria-hidden="true" />
      </button>

      {open && (
        // listbox ではなく menu にしてある。listbox の option は**対話要素を含められない**ので、
        // option の中にボタンを置くと、支援技術からは押せるものが見えない
        // （押下の判定も親の option には来ない）。ここでは押せるもの自体を menuitem にする。
        <ul
          role="menu"
          aria-label="ワークスペース"
          className="absolute left-0 right-0 z-20 mt-1 max-h-64 overflow-y-auto rounded-lg border border-surface-3 bg-surface-1 py-1 shadow-lg"
        >
          {workspaces.map((workspace) => (
            <li key={workspace.slug} role="none">
              <button
                type="button"
                role="menuitem"
                aria-current={workspace.slug === activeSlug}
                onClick={() => {
                  onSelect(workspace.slug);
                  setOpen(false);
                }}
                className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm hover:bg-surface-2"
              >
                <span className="min-w-0 flex-1 truncate">{workspace.name}</span>
                {workspace.slug === activeSlug && (
                  <CheckIcon className="h-4 w-4 shrink-0 text-brand-500" aria-hidden="true" />
                )}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
