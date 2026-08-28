import { useEffect, useRef, useState } from 'react';
import { EllipsisHorizontalIcon, PlusIcon } from '@heroicons/react/24/outline';
import type { NoteDropTarget, NoteMoveActions } from '@/entities/note';

export interface NoteRowActionsProps {
  /** 読み上げに使う対象の名前（「〜の下にページを追加」のように読ませる）。 */
  label: string;
  onCreateChild: () => void;
  /** 未指定ならメニュー自体を出さない（スペースの見出しなど、名前を変えられない行）。 */
  onRename?: () => void;
  onArchive?: () => void;
  /** 動かせる 4 つの向き。null の向きは項目自体を出さない（押せない項目を並べない）。 */
  moves?: NoteMoveActions;
  onMove?: (target: NoteDropTarget) => void;
}

/**
 * NoteRowActions は行の右端に出る操作（＋ と ⋯）。
 *
 * 常に見えていると木の形が読みにくくなるので、**触れているときと、キーボードで
 * たどり着いたときだけ**濃くする。`opacity` で消しているだけで DOM からは外さない —
 * 外すと Tab の順序が触れるたびに変わり、キーボードでは追えなくなる。
 */
export default function NoteRowActions({
  label,
  onCreateChild,
  onRename,
  onArchive,
  moves,
  onMove,
}: NoteRowActionsProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menuOpen) return;
    const onDocumentMouseDown = (event: MouseEvent) => {
      if (
        containerRef.current &&
        event.target instanceof Node &&
        containerRef.current.contains(event.target)
      ) {
        return;
      }
      setMenuOpen(false);
    };
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setMenuOpen(false);
    };
    document.addEventListener('mousedown', onDocumentMouseDown);
    document.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('mousedown', onDocumentMouseDown);
      document.removeEventListener('keydown', onKeyDown);
    };
  }, [menuOpen]);

  return (
    <div
      ref={containerRef}
      className={`relative flex shrink-0 items-center gap-0.5 transition-opacity ${
        menuOpen ? 'opacity-100' : 'opacity-0 focus-within:opacity-100 group-hover:opacity-100'
      }`}
    >
      {onRename && (
        <>
          <button
            type="button"
            onClick={() => setMenuOpen((prev) => !prev)}
            aria-expanded={menuOpen}
            aria-label={`${label} の操作`}
            className="rounded p-1 text-[var(--color-text-muted)] hover:bg-surface-3"
          >
            <EllipsisHorizontalIcon className="h-4 w-4" aria-hidden="true" />
          </button>

          {menuOpen && (
            // 素のボタンの一覧として出す（menu を名乗ると矢印キーでの移動を約束することになる）。
            <ul className="absolute right-0 top-full z-20 mt-1 w-40 rounded-lg border border-surface-3 bg-surface-1 py-1 shadow-lg">
              <li>
                <button
                  type="button"
                  onClick={() => {
                    setMenuOpen(false);
                    onRename();
                  }}
                  className="w-full px-3 py-1.5 text-left text-sm hover:bg-surface-2"
                >
                  名前を変更
                </button>
              </li>
              {/*
                動かす項目。**ドラッグと同じ行き先を、同じ経路で送る。**
                キーボードだけの人にはこれが唯一の並べ替えの手段になるので、
                押せない向きは項目ごと出さない（押しても何も起きない項目を並べない）。
              */}
              {moves &&
                onMove &&
                (
                  [
                    ['up', '上へ移動'],
                    ['down', '下へ移動'],
                    ['indent', 'ひとつ内側へ'],
                    ['outdent', 'ひとつ外側へ'],
                  ] as const
                ).map(([key, text]) => {
                  const target = moves[key];
                  if (!target) return null;
                  return (
                    <li key={key}>
                      <button
                        type="button"
                        onClick={() => {
                          setMenuOpen(false);
                          onMove(target);
                        }}
                        className="w-full px-3 py-1.5 text-left text-sm hover:bg-surface-2"
                      >
                        {text}
                      </button>
                    </li>
                  );
                })}
              {onArchive && (
                <li>
                  <button
                    type="button"
                    onClick={() => {
                      setMenuOpen(false);
                      onArchive();
                    }}
                    className="w-full px-3 py-1.5 text-left text-sm hover:bg-surface-2"
                  >
                    アーカイブ
                  </button>
                </li>
              )}
            </ul>
          )}
        </>
      )}

      <button
        type="button"
        onClick={onCreateChild}
        aria-label={`${label} の下にページを追加`}
        title="中にページを作成"
        className="rounded p-1 text-[var(--color-text-muted)] hover:bg-surface-3"
      >
        <PlusIcon className="h-4 w-4" aria-hidden="true" />
      </button>
    </div>
  );
}
