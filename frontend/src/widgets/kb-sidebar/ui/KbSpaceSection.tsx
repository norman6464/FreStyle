import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ChevronRightIcon, PlusIcon } from '@heroicons/react/24/outline';
import { flattenKbTree, type KbDropTarget, type KbPage, type KbSpace } from '@/entities/knowledge-base';
import { toDropTarget, type KbDropZone } from '../model/dropZone';
import { useToast } from '@/shared/lib/hooks/useToast';
import type { KbSpaceState } from '../model/useKnowledgeBaseTree';
import KbPageRow from './KbPageRow';
import KbHiddenChildrenRow from './KbHiddenChildrenRow';

export interface KbSpaceSectionProps {
  space: KbSpace;
  state: KbSpaceState | undefined;
  workspaceSlug: string;
  activePageId?: string;
  expandedPageIds: ReadonlySet<string>;
  onToggleSpace: (spaceId: string) => void;
  onTogglePage: (pageId: string) => void;
  onRetry: (spaceId: string) => void;
  /** ページを作る。**失敗は投げてくる。** */
  onCreatePage: (spaceId: string, parentId?: string) => Promise<KbPage>;
  /** 題名を変える。**失敗は投げてくる。** */
  onRenamePage: (spaceId: string, pageId: string, title: string) => Promise<KbPage>;
  /** 子孫ごとアーカイブする。**失敗は投げてくる。** */
  onArchivePage: (spaceId: string, pageId: string) => Promise<void>;
  /** 現役へ戻す。**失敗は投げてくる。** */
  onUnarchivePage: (spaceId: string, pageId: string) => Promise<void>;
  /** アーカイブ済みを見ているか。 */
  archivedMode: boolean;
  /** ドラッグで動かす。**先に画面が動き、断られたら元へ戻る。失敗は投げてくる。** */
  onMovePage: (spaceId: string, pageId: string, target: KbDropTarget) => Promise<void>;
}

/**
 * KbSpaceSection はスペース 1 つ分の見出しと、その配下のページの木。
 *
 * スペースは**同時に複数見たい**（自分のページと部署のページを行き来する）ので、
 * 切り替えではなく見出しとして並べる。ワークスペースは会社の境界で同時に見る場面が無いので、
 * あちらは切り替えにしてある。同時に見たいものは並べ、排他のものは切り替える。
 *
 * # 失敗を必ず利用者に伝える
 *
 * 書き換えの呼び出しはすべてここで try/catch し、**失敗したときだけ知らせを出す**。
 * このリポジトリには「操作は失敗したのに成功の表示が出る」轍が既にあり
 * （コース削除・教材の保存）、原因はどれも操作関数が失敗を投げなかったことだった。
 * model 側は投げる作りにしてあるので、握り潰すには catch を書くしかない。
 * **中身が空の catch は書かないこと。**
 *
 * 成功のトーストは出さない。作られたページが木に現れ、そのページへ移動するので、
 * 起きたことは画面が示している。何も起きていないときにだけ言葉が要る。
 */
export default function KbSpaceSection({
  space,
  state,
  workspaceSlug,
  activePageId,
  expandedPageIds,
  onToggleSpace,
  onTogglePage,
  onRetry,
  onCreatePage,
  onRenamePage,
  onArchivePage,
  onUnarchivePage,
  archivedMode,
  onMovePage,
}: KbSpaceSectionProps) {
  const navigate = useNavigate();
  const { showToast } = useToast();
  // 作った直後のページは、そのまま題名を書き換えられる状態で出す
  // （「無題」のまま置き去りにされるのを減らす）。
  const [renamingPageId, setRenamingPageId] = useState<string | null>(null);
  // 掴んでいる行と、いま指している落下先。ドラッグの間だけの見た目の状態で、
  // 木そのものには触らない（動かすのは落としたとき）。
  const [draggingPageId, setDraggingPageId] = useState<string | null>(null);
  const [dropAt, setDropAt] = useState<{ pageId: string; zone: KbDropZone } | null>(null);

  const open = state?.open ?? false;
  const entries = state?.tree ? flattenKbTree(state.tree.pages, expandedPageIds) : [];
  const hiddenAtRoot = state?.tree?.hasHiddenChildren ?? false;

  const createPage = async (parentId?: string) => {
    try {
      const page = await onCreatePage(space.id, parentId);
      setRenamingPageId(page.id);
      navigate(`/kb/${workspaceSlug}/pages/${page.id}`);
    } catch {
      showToast('error', 'ページを作成できませんでした');
    }
  };

  const commitRename = async (pageId: string, title: string) => {
    try {
      await onRenamePage(space.id, pageId, title);
      setRenamingPageId(null);
    } catch {
      showToast('error', '名前を変更できませんでした');
      // 入力欄は開いたままにする（投げると KbInlineRename がフォーカスを戻す）。
      // ここで閉じると、書いた文字は消えるのに元の題名が残り、保存されたのか分からなくなる。
      throw new Error('rename failed');
    }
  };

  const archivePage = async (pageId: string) => {
    try {
      await onArchivePage(space.id, pageId);
    } catch {
      showToast('error', 'アーカイブできませんでした');
    }
  };

  const unarchivePage = async (pageId: string) => {
    try {
      await onUnarchivePage(space.id, pageId);
    } catch {
      showToast('error', '復帰できませんでした');
    }
  };

  const endDrag = () => {
    setDraggingPageId(null);
    setDropAt(null);
  };

  const dropOnRow = async (pageId: string, zone: KbDropZone) => {
    const moving = draggingPageId;
    endDrag();
    if (!moving || moving === pageId) return;
    try {
      await onMovePage(space.id, moving, toDropTarget(zone, pageId));
    } catch {
      // 並びは model 側で動かす前へ戻っている。ここは知らせるだけ。
      showToast('error', '移動できませんでした');
    }
  };

  return (
    <section className="mb-1">
      <h2 className="group flex items-center gap-1 rounded-md pr-1 hover:bg-surface-2">
        <button
          type="button"
          onClick={() => onToggleSpace(space.id)}
          aria-expanded={open}
          className="flex min-w-0 flex-1 items-center gap-1 rounded-md px-1 py-1 text-left text-xs font-semibold uppercase tracking-wide text-[var(--color-text-muted)]"
        >
          <ChevronRightIcon
            className={`h-3 w-3 shrink-0 transition-transform ${open ? 'rotate-90' : ''}`}
            aria-hidden="true"
          />
          <span className="truncate">{space.name}</span>
        </button>
        {/* スペース直下に作る。見出しは名前を変えられないので ＋ だけ出す。
            アーカイブ済みを見ているときは出さない（そこには作れない）。 */}
        {!archivedMode && (
        <button
          type="button"
          onClick={() => void createPage()}
          aria-label={`${space.name} にページを追加`}
          className="shrink-0 rounded p-1 text-[var(--color-text-muted)] opacity-0 transition-opacity hover:bg-surface-3 focus:opacity-100 group-hover:opacity-100"
        >
          <PlusIcon className="h-4 w-4" aria-hidden="true" />
        </button>
        )}
      </h2>

      {open && (
        <div className="mt-0.5">
          {state?.loading && (
            <p className="px-2 py-1 text-xs text-[var(--color-text-muted)]">読み込み中…</p>
          )}

          {/*
            失敗は必ず出し、やり直せるようにする。黙って空にすると「ページが 1 枚も無い」と
            見分けが付かず、利用者は消えたと思う。
          */}
          {state?.error && (
            <div className="px-2 py-1 text-xs text-red-600">
              <p>{state.error}</p>
              <button
                type="button"
                onClick={() => onRetry(space.id)}
                className="mt-0.5 underline hover:no-underline"
              >
                再試行
              </button>
            </div>
          )}

          {!state?.loading && !state?.error && entries.length === 0 && !hiddenAtRoot && (
            <p className="px-2 py-1 text-xs text-[var(--color-text-muted)]">
              {archivedMode ? 'アーカイブしたページはありません' : 'ページがありません'}
            </p>
          )}

          {entries.length > 0 && (
            // 平らな配列だが aria-level で段を伝えるので、木として読み上げられる。
            <ul role="tree" aria-label={`${space.name} のページ`} className="space-y-px">
              {entries.map((entry) =>
                entry.kind === 'page' ? (
                  <KbPageRow
                    key={entry.page.id}
                    row={entry}
                    workspaceSlug={workspaceSlug}
                    active={entry.page.id === activePageId}
                    onToggle={onTogglePage}
                    renaming={entry.page.id === renamingPageId}
                    onStartRename={setRenamingPageId}
                    onCancelRename={() => setRenamingPageId(null)}
                    onCommitRename={commitRename}
                    onCreateChild={(parentId) => void createPage(parentId)}
                    onArchive={(pageId) => void archivePage(pageId)}
                    archivedMode={archivedMode}
                    onUnarchive={(pageId) => void unarchivePage(pageId)}
                    // アーカイブ済みでは並べ替えを受け付けない（現役に戻してから）。
                    draggable={!archivedMode}
                    dragging={entry.page.id === draggingPageId}
                    onDragStart={setDraggingPageId}
                    onDragEnd={endDrag}
                    dropZone={
                      dropAt?.pageId === entry.page.id && draggingPageId !== entry.page.id
                        ? dropAt.zone
                        : null
                    }
                    onDragOverRow={(pageId, zone) => setDropAt({ pageId, zone })}
                    onDropOnRow={(pageId, zone) => void dropOnRow(pageId, zone)}
                  />
                ) : (
                  <KbHiddenChildrenRow key={`hidden-${entry.parentId}`} depth={entry.depth} />
                ),
              )}
            </ul>
          )}

          {/* スペース直下の印。1 件も見えないスペースでは backend が必ず false を返す。 */}
          {hiddenAtRoot && (
            <p className="px-2 py-0.5 text-xs text-[var(--color-text-muted)]">
              表示できないページがあります
            </p>
          )}
        </div>
      )}
    </section>
  );
}
