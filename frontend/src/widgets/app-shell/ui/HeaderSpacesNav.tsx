import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ChevronDownIcon, FolderIcon } from '@heroicons/react/24/outline';
import { useToast } from '@/shared/lib/hooks/useToast';
import { useClickOutside } from '@/shared/lib/hooks/useClickOutside';
import { NameCreateForm } from '@/shared/ui';
import { KbRepository, useCurrentKbSpace, type KbMySpace } from '@/entities/kb';

export interface HeaderSpacesNavProps {
  /** ナビリンクと揃える見た目のクラス（Header.tsx の navLinkClass）。 */
  className: string;
  /** モバイルメニュー内では縦積みにする（display:block）。 */
  block?: boolean;
}

/**
 * HeaderSpacesNav はヘッダーの「スペース ▾」。他のスペースへの切替とスペースの作成を持つ
 * （entities/kb の共有値 useCurrentKbSpace で「今いるスペース」を受け取る）。
 *
 * 今いるスペースがまだ分からない（KB のページを一度も開いていない）ときはドロップ
 * ダウンを持たない素のリンクにする — その場合ここで一覧を出しても、どのワーク
 * スペースのスペース一覧を出すべきか決められないため、入口解決を持つ /kb/spaces へ
 * 素通しする（resolveEntryKbSpaceId が最初に見つかったスペースへ移す）。
 */
export default function HeaderSpacesNav({ className, block = false }: HeaderSpacesNavProps) {
  const current = useCurrentKbSpace();
  const navigate = useNavigate();
  const { showToast } = useToast();
  const [open, setOpen] = useState(false);
  const [mySpaces, setMySpaces] = useState<KbMySpace[] | null>(null);
  const [addingSpace, setAddingSpace] = useState(false);
  const [addingPrivateSpace, setAddingPrivateSpace] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  useClickOutside(containerRef, open, () => setOpen(false));

  const workspaceSlug = current?.workspaceSlug ?? null;

  useEffect(() => {
    if (!open || !workspaceSlug) return;
    let cancelled = false;
    KbRepository.fetchMySpaces(workspaceSlug)
      .then((list) => {
        if (!cancelled) setMySpaces(list);
      })
      .catch(() => {
        if (!cancelled) setMySpaces([]);
      });
    return () => {
      cancelled = true;
    };
  }, [open, workspaceSlug]);

  if (!current) {
    return (
      <Link to="/kb/spaces" className={`${className} ${block ? 'block' : ''}`}>
        スペース
      </Link>
    );
  }

  const createSpace = async (input: { name: string; visibility?: 'workspace' | 'private' }) => {
    try {
      const space = await KbRepository.createSpace(current.workspaceSlug, input);
      setAddingSpace(false);
      setAddingPrivateSpace(false);
      setOpen(false);
      navigate(`/kb/spaces/${space.id}`);
    } catch {
      showToast('error', 'スペースを作成できませんでした');
      throw new Error('create space failed');
    }
  };

  return (
    <div className="relative" ref={containerRef}>
      <button
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        aria-expanded={open}
        className={`${className} ${block ? 'flex w-full items-center justify-between' : 'flex items-center gap-1'}`}
      >
        スペース
        <ChevronDownIcon className="h-3.5 w-3.5" aria-hidden="true" />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-20 mt-1 w-56 rounded-lg border border-surface-3 bg-surface-1 py-1 shadow-lg">
          {mySpaces === null && (
            <p className="px-3 py-1.5 text-xs text-[var(--color-text-muted)]">読み込み中…</p>
          )}
          {mySpaces?.map((s) => (
            <Link
              key={s.id}
              to={`/kb/spaces/${s.id}`}
              onClick={() => setOpen(false)}
              className={`flex items-center gap-1.5 px-3 py-1.5 text-sm hover:bg-surface-2 ${
                s.id === current.spaceId
                  ? 'font-semibold text-[var(--color-text-primary)]'
                  : 'text-[var(--color-text-secondary)]'
              }`}
            >
              <FolderIcon className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
              <span className="truncate">{s.name}</span>
            </Link>
          ))}
          <div className="mt-1 border-t border-surface-3 pt-1">
            {addingSpace ? (
              <div className="px-2 pb-1">
                <NameCreateForm what="スペース" onCreate={createSpace} />
                <button
                  type="button"
                  onClick={() => setAddingSpace(false)}
                  className="w-full px-2 pb-1 text-left text-xs text-[var(--color-text-muted)] hover:underline"
                >
                  やめる
                </button>
              </div>
            ) : (
              <button
                type="button"
                onClick={() => setAddingSpace(true)}
                className="w-full px-3 py-1.5 text-left text-sm text-[var(--color-text-secondary)] hover:bg-surface-2"
              >
                スペースを作成
              </button>
            )}
            {addingPrivateSpace ? (
              <div className="px-2 pb-1">
                <NameCreateForm
                  what="プライベートスペース"
                  onCreate={(input) => createSpace({ ...input, visibility: 'private' })}
                />
                <button
                  type="button"
                  onClick={() => setAddingPrivateSpace(false)}
                  className="w-full px-2 pb-1 text-left text-xs text-[var(--color-text-muted)] hover:underline"
                >
                  やめる
                </button>
              </div>
            ) : (
              <button
                type="button"
                onClick={() => setAddingPrivateSpace(true)}
                className="w-full px-3 py-1.5 text-left text-sm text-[var(--color-text-secondary)] hover:bg-surface-2"
              >
                プライベートスペースを作成
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
