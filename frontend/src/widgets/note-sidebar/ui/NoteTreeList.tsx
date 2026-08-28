import type { NoteDropTarget, NotePageTreeNode } from '@/entities/note';
import type { NoteDropZone } from '../model/dropZone';
import NotePageRow, { type NoteRowCallbacks } from './NotePageRow';
import NoteHiddenChildrenRow from './NoteHiddenChildrenRow';

export interface NoteTreeListProps extends NoteRowCallbacks {
  nodes: NotePageTreeNode[];
  /** 0 がスペース直下。字下げにだけ使う（段の深さは入れ子そのものが表す）。 */
  depth: number;
  /** この段の親。最上段は null。「ひとつ外側へ」の行き先を決めるのに要る。 */
  parentId: string | null;
  /** この段の直下に、自分には見えないページが在るか。 */
  hasHiddenChildren: boolean;
  expandedPageIds: ReadonlySet<string>;
  activePageId?: string;
  workspaceSlug: string;
  renamingPageId: string | null;
  draggingPageId: string | null;
  dropAt: { pageId: string; zone: NoteDropZone } | null;
  archivedMode: boolean;
  label: string;
}

/**
 * NoteTreeList は木の 1 段を描く。子は**入れ子の ul** として描く。
 *
 * 平らに並べて aria-level で段を伝える形はやめた。`role="tree"` を名乗る以上、
 * 矢印キーでの移動まで用意するのが筋だが、行の中にリンクと操作ボタンが同居している以上、
 * それは Tab とは別のもう 1 つの操作体系を作ることになる。名乗りだけ残すのは嘘なので、
 * **名乗りを外して、段の深さは入れ子そのものに表させる**（読み上げは入れ子を辿れる）。
 *
 * 字下げの depth は見た目のためだけに残す。深さの情報としては使わない。
 */
export default function NoteTreeList({
  nodes,
  depth,
  parentId,
  hasHiddenChildren,
  expandedPageIds,
  activePageId,
  workspaceSlug,
  renamingPageId,
  draggingPageId,
  dropAt,
  archivedMode,
  label,
  ...callbacks
}: NoteTreeListProps) {
  return (
    <ul aria-label={label} className="space-y-px">
      {nodes.map((node, index) => {
        const expanded = node.children.length > 0 && expandedPageIds.has(node.page.id);
        return (
          <li key={node.page.id}>
            <NotePageRow
              node={node}
              depth={depth}
              siblings={nodes}
              index={index}
              parentId={parentId}
              expanded={expanded}
              workspaceSlug={workspaceSlug}
              active={node.page.id === activePageId}
              renaming={node.page.id === renamingPageId}
              archivedMode={archivedMode}
              dragging={node.page.id === draggingPageId}
              dropZone={
                dropAt?.pageId === node.page.id && draggingPageId !== node.page.id
                  ? dropAt.zone
                  : null
              }
              {...callbacks}
            />
            {expanded && (
              <NoteTreeList
                nodes={node.children}
                depth={depth + 1}
                parentId={node.page.id}
                hasHiddenChildren={node.hasHiddenChildren}
                expandedPageIds={expandedPageIds}
                activePageId={activePageId}
                workspaceSlug={workspaceSlug}
                renamingPageId={renamingPageId}
                draggingPageId={draggingPageId}
                dropAt={dropAt}
                archivedMode={archivedMode}
                label={`${node.page.title} の中`}
                {...callbacks}
              />
            )}
            {/*
              開けない段（見える子が 1 枚も無い）は開閉の三角が出ないので、
              伏せた印はここに出す。開ける段の分は上の入れ子側が出す。
            */}
            {node.children.length === 0 && node.hasHiddenChildren && (
              <NoteHiddenChildrenRow depth={depth + 1} />
            )}
          </li>
        );
      })}
      {hasHiddenChildren && <NoteHiddenChildrenRow depth={depth} />}
    </ul>
  );
}

export type { NoteDropTarget };
