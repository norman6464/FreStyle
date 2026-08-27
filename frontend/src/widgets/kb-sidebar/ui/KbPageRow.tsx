import { Link } from 'react-router-dom';
import { ChevronRightIcon } from '@heroicons/react/24/outline';
import { KbPageGroupIcon, KbPageIcon } from '@/shared/ui/icons/kb';
import { kbMoveActions, type KbDropTarget, type KbPageTreeNode } from '@/entities/knowledge-base';
import KbRowActions from './KbRowActions';
import KbInlineRename from './KbInlineRename';
import { dropZoneFromEvent, type KbDropZone } from '../model/dropZone';

/** 1 段下がるごとの字下げ幅（px）。三角の幅とほぼ同じにして、段が目で追えるようにする。 */
export const KB_INDENT_PX = 14;

/** 三角（または三角ぶんの空き）の幅。伏せた印の行の字下げを揃えるのに使う。 */
export const KB_TOGGLE_WIDTH_PX = 22;

/** 段をまたいでそのまま渡していく操作。行ごとに変わらないものだけを集めてある。 */
export interface KbRowCallbacks {
  onToggle: (pageId: string) => void;
  onStartRename: (pageId: string) => void;
  onCancelRename: () => void;
  /** 確定。**失敗は投げてくる**ので、握り潰さない限り必ず表に出る。 */
  onCommitRename: (pageId: string, title: string) => Promise<void>;
  onCreateChild: (parentId: string) => void;
  onArchive: (pageId: string) => void;
  onUnarchive: (pageId: string) => void;
  /** メニューやドラッグから動かす。 */
  onMove: (pageId: string, target: KbDropTarget) => void;
  onDragStart: (pageId: string) => void;
  onDragEnd: () => void;
  onDragOverRow: (pageId: string, zone: KbDropZone) => void;
  onDropOnRow: (pageId: string, zone: KbDropZone) => void;
}

export interface KbPageRowProps extends KbRowCallbacks {
  node: KbPageTreeNode;
  /** 字下げに使う段の数。深さの情報としては使わない（それは入れ子が表す）。 */
  depth: number;
  /** この行が並んでいる兄弟の一覧と、その中での位置。動かせる向きを決めるのに要る。 */
  siblings: KbPageTreeNode[];
  index: number;
  parentId: string | null;
  expanded: boolean;
  workspaceSlug: string;
  active: boolean;
  renaming: boolean;
  archivedMode: boolean;
  dragging: boolean;
  dropZone: KbDropZone | null;
}

/**
 * KbPageRow はツリーの 1 行。開閉の三角と題名へのリンク、右端に操作。
 *
 * 開閉は**リンクとは別のボタン**にしてある。行そのものを押すと開閉する作りにすると、
 * 「開くつもりで押したらページが切り替わった」が必ず起きる。
 *
 * # role="treeitem" を名乗らない
 *
 * 矢印キーでの移動を用意しないまま名乗るのは嘘になる。用意しようとすると、行の中に
 * リンクと操作ボタンが同居しているぶん、Tab とは別のもう 1 つの操作体系を作ることになる。
 * 段の深さは**入れ子の ul** が表し、いま開いているページは **aria-current** が表す。
 * どちらも標準の意味で、支援技術に別の約束をしない。
 */
export default function KbPageRow({
  node,
  depth,
  siblings,
  index,
  parentId,
  expanded,
  workspaceSlug,
  active,
  renaming,
  archivedMode,
  dragging,
  dropZone,
  onToggle,
  onStartRename,
  onCancelRename,
  onCommitRename,
  onCreateChild,
  onArchive,
  onUnarchive,
  onMove,
  onDragStart,
  onDragEnd,
  onDragOverRow,
  onDropOnRow,
}: KbPageRowProps) {
  const { page } = node;
  const hasChildren = node.children.length > 0;

  // 子を持つページはフォルダ、持たないページは紙。
  //
  // 見える子が居るかで選ぶので、**伏せた子しか居ないページは紙のまま**になる。
  // ここでフォルダにすると、開閉の三角が無いのにフォルダ、という食い違った行になり、
  // さらに「この下に何かある」ことを形からも二重に漏らす。
  const Icon = hasChildren ? KbPageGroupIcon : KbPageIcon;

  // 落下先を線と枠で描き分ける。並べ替え（上下の線）と入れ子（枠）は別の操作なので、
  // 見た目でも別にする。同じ強調にすると、どちらになるのか落とすまで分からない。
  const dropClass =
    dropZone === 'into'
      ? 'ring-1 ring-inset ring-brand-400'
      : dropZone === 'before'
        ? 'border-t-2 border-brand-400'
        : dropZone === 'after'
          ? 'border-b-2 border-brand-400'
          : '';

  const draggable = !archivedMode;

  return (
    <div
      draggable={draggable && !renaming}
      onDragStart={(event) => {
        // text/plain を入れておかないと Firefox がドラッグを開始しない。
        event.dataTransfer.setData('text/plain', page.id);
        event.dataTransfer.effectAllowed = 'move';
        onDragStart(page.id);
      }}
      onDragEnd={onDragEnd}
      onDragOver={(event) => {
        if (!draggable) return;
        // 既定の動作を止めないと drop が起きない（ブラウザの決まり）。
        event.preventDefault();
        event.dataTransfer.dropEffect = 'move';
        onDragOverRow(
          page.id,
          dropZoneFromEvent(event.currentTarget.getBoundingClientRect(), event.clientY),
        );
      }}
      onDrop={(event) => {
        if (!draggable) return;
        event.preventDefault();
        onDropOnRow(
          page.id,
          dropZoneFromEvent(event.currentTarget.getBoundingClientRect(), event.clientY),
        );
      }}
      className={`group flex items-center gap-1 rounded-md pr-1 transition-colors ${
        active ? 'bg-brand-500/10 text-brand-600' : 'hover:bg-surface-2'
      } ${dragging ? 'opacity-40' : ''} ${dropClass}`}
      style={{ paddingLeft: depth * KB_INDENT_PX }}
    >
      {hasChildren ? (
        <button
          type="button"
          onClick={() => onToggle(page.id)}
          aria-expanded={expanded}
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

      {renaming ? (
        <div className="flex min-w-0 flex-1 items-center gap-1.5 py-0.5">
          <Icon className="h-4 w-4 shrink-0 text-[var(--color-text-muted)]" />
          <KbInlineRename
            initialTitle={page.title}
            onCommit={(title) => onCommitRename(page.id, title)}
            onCancel={onCancelRename}
          />
        </div>
      ) : (
        <>
          <Link
            to={`/kb/${workspaceSlug}/pages/${page.id}`}
            // いま開いているページであることは、role ではなくここが表す。
            aria-current={active ? 'page' : undefined}
            className={`flex min-w-0 flex-1 items-center gap-1.5 py-1 text-sm ${
              active ? 'font-medium' : 'text-[var(--color-text-primary)]'
            }`}
          >
            <Icon className="h-4 w-4 shrink-0 text-[var(--color-text-muted)]" />
            <span className="truncate">{page.title}</span>
          </Link>
          {archivedMode ? (
            // アーカイブ済みの行では、作る・名前を変える・動かすは出さない（現役に戻してから）。
            // 復帰できるのはアーカイブの根だけ。親がまだアーカイブ中の行に出すと、
            // 押せるのに必ず断られるボタンになる。
            !node.parentArchived && (
              <button
                type="button"
                onClick={() => onUnarchive(page.id)}
                className="shrink-0 rounded px-1.5 py-0.5 text-xs text-[var(--color-text-muted)] opacity-0 transition-opacity hover:bg-surface-3 focus:opacity-100 group-hover:opacity-100"
              >
                復帰
              </button>
            )
          ) : (
            <KbRowActions
              label={page.title}
              onCreateChild={() => onCreateChild(page.id)}
              onRename={() => onStartRename(page.id)}
              onArchive={() => onArchive(page.id)}
              // ドラッグと同じ行き先を、同じ経路で送る。キーボードのためだけに
              // 別の口を作らない（作ると失敗の扱いも二重になる）。
              moves={kbMoveActions(siblings, index, parentId)}
              onMove={(target) => onMove(page.id, target)}
            />
          )}
        </>
      )}
    </div>
  );
}
