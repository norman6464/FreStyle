import { Link } from 'react-router-dom';
import { ChevronRightIcon } from '@heroicons/react/24/outline';
import { KbPageGroupIcon, KbPageIcon } from '@/shared/ui/icons/kb';
import type { KbTreeRow } from '@/entities/knowledge-base';

/** 1 段下がるごとの字下げ幅（px）。三角の幅とほぼ同じにして、段が目で追えるようにする。 */
export const KB_INDENT_PX = 14;

/** 三角（または三角ぶんの空き）の幅。伏せた件数の行の字下げを揃えるのに使う。 */
export const KB_TOGGLE_WIDTH_PX = 22;

export interface KbPageRowProps {
  row: KbTreeRow;
  workspaceSlug: string;
  /** いま開いているページか。 */
  active: boolean;
  onToggle: (pageId: string) => void;
}

/**
 * KbPageRow はツリーの 1 行。開閉の三角と題名へのリンク。
 *
 * 開閉は**リンクとは別のボタン**にしてある。行そのものを押すと開閉する作りにすると、
 * 「開くつもりで押したらページが切り替わった」が必ず起きる。押した場所で意味が変わる操作は、
 * キーボードやスクリーンリーダーからはさらに区別が付かない。
 */
export default function KbPageRow({ row, workspaceSlug, active, onToggle }: KbPageRowProps) {
  const { page, depth, hasChildren, expanded } = row;

  // 子を持つページはフォルダ、持たないページは紙。
  //
  // 見える子が居るかで選ぶので、**伏せた子しか居ないページは紙のまま**になる。
  // ここでフォルダにすると、開閉の三角が無いのにフォルダ、という食い違った行になり、
  // さらに「この下に何かある」ことを形からも二重に漏らす。
  const Icon = hasChildren ? KbPageGroupIcon : KbPageIcon;

  return (
    <li role="none">
      <div
        role="treeitem"
        aria-level={depth + 1}
        aria-expanded={hasChildren ? expanded : undefined}
        aria-selected={active}
        className={`group flex items-center gap-1 rounded-md pr-1 transition-colors ${
          active ? 'bg-brand-500/10 text-brand-600' : 'hover:bg-surface-2'
        }`}
        style={{ paddingLeft: depth * KB_INDENT_PX }}
      >
        {hasChildren ? (
          <button
            type="button"
            onClick={() => onToggle(page.id)}
            aria-label={expanded ? `${page.title} を閉じる` : `${page.title} を開く`}
            className="shrink-0 rounded p-1 text-[var(--color-text-muted)] hover:bg-surface-3"
          >
            <ChevronRightIcon
              className={`h-3.5 w-3.5 transition-transform ${expanded ? 'rotate-90' : ''}`}
              aria-hidden="true"
            />
          </button>
        ) : (
          // 子が無い行にも同じ幅を空ける。空けないと題名の左端が段ごとに揃わない。
          <span style={{ width: KB_TOGGLE_WIDTH_PX }} className="shrink-0" aria-hidden="true" />
        )}

        <Link
          to={`/kb/${workspaceSlug}/pages/${page.id}`}
          className={`flex min-w-0 flex-1 items-center gap-1.5 py-1 text-sm ${
            active ? 'font-medium' : 'text-[var(--color-text-primary)]'
          }`}
        >
          <Icon className="h-4 w-4 shrink-0 text-[var(--color-text-muted)]" />
          <span className="truncate">{page.title}</span>
        </Link>
      </div>
    </li>
  );
}
