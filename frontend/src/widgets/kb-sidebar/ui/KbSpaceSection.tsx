import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ChevronRightIcon, EllipsisHorizontalIcon, PlusIcon } from '@heroicons/react/24/outline';
import type { KbDropTarget, KbPage, KbSpace } from '@/entities/knowledge-base';
import { toDropTarget, type KbDropZone } from '../model/dropZone';
import { useToast } from '@/shared/lib/hooks/useToast';
import type { KbSpaceState } from '../model/useKnowledgeBaseTree';
import KbTreeList from './KbTreeList';
import KbInlineRename from './KbInlineRename';

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
  /** スペースの表示名を変える。**失敗は投げてくる。** */
  onRenameSpace: (spaceId: string, name: string) => Promise<KbSpace>;
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
  onRenameSpace,
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
  const tree = state?.tree ?? null;
  const hiddenAtRoot = tree?.hasHiddenChildren ?? false;
  // 見出しの名前を書き換え中か（行の renaming と同じ流儀の、見出し版）。
  const [renamingSpace, setRenamingSpace] = useState(false);
  const [spaceMenuOpen, setSpaceMenuOpen] = useState(false);

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

  const movePage = async (pageId: string, target: KbDropTarget) => {
    try {
      await onMovePage(space.id, pageId, target);
    } catch {
      // 並びは model 側で動かす前へ戻っている。ここは知らせるだけ。
      showToast('error', '移動できませんでした');
    }
  };

  const dropOnRow = async (pageId: string, zone: KbDropZone) => {
    const moving = draggingPageId;
    endDrag();
    if (!moving || moving === pageId) return;
    await movePage(moving, toDropTarget(zone, pageId));
  };

  const commitSpaceRename = async (name: string) => {
    try {
      await onRenameSpace(space.id, name);
      setRenamingSpace(false);
    } catch {
      showToast('error', 'スペースの名前を変更できませんでした');
      // 入力欄は開いたままにする（ページの改名と同じ理由 — 閉じると書いた文字が消える）。
      throw new Error('rename space failed');
    }
  };

  return (
    <section className="mb-1">
      <h2 className="group relative flex items-center gap-1 rounded-md pr-1 hover:bg-surface-2">
        {renamingSpace ? (
          <div className="flex min-w-0 flex-1 items-center gap-1 px-1 py-1">
            <ChevronRightIcon
              className={`h-3 w-3 shrink-0 transition-transform ${open ? 'rotate-90' : ''}`}
              aria-hidden="true"
            />
            <KbInlineRename
              initialTitle={space.name}
              ariaLabel="スペースの名前"
              onCommit={commitSpaceRename}
              onCancel={() => setRenamingSpace(false)}
            />
          </div>
        ) : (
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
        )}
        {/* 見出しの操作は**常時**出す（行はホバーで現れるが、見出しは面の入口なので
            見えていないと「どこから作るのか」が分からない）。見本の画面もこの形。
            アーカイブ済みを見ているときは出さない（そこには作れず、名前も戻ってから変える）。 */}
        {!archivedMode && !renamingSpace && (
          <>
            <button
              type="button"
              onClick={() => void createPage()}
              aria-label={`${space.name} にページを追加`}
              title="ページを追加"
              className="shrink-0 rounded p-1 text-[var(--color-text-muted)] transition-colors hover:bg-surface-3"
            >
              <PlusIcon className="h-4 w-4" aria-hidden="true" />
            </button>
            <button
              type="button"
              onClick={() => setSpaceMenuOpen((prev) => !prev)}
              // menu を名乗らない（矢印キーでの移動を用意していないため）。開閉は aria-expanded が表す。
              aria-expanded={spaceMenuOpen}
              aria-label={`${space.name} の操作`}
              className="shrink-0 rounded p-1 text-[var(--color-text-muted)] transition-colors hover:bg-surface-3"
            >
              <EllipsisHorizontalIcon className="h-4 w-4" aria-hidden="true" />
            </button>
            {spaceMenuOpen && (
              // 素のボタンの一覧として出す（menu を名乗ると矢印キーでの移動を約束することになる）。
              <ul className="absolute right-0 top-full z-20 mt-1 w-44 rounded-lg border border-surface-3 bg-surface-1 py-1 shadow-lg">
                <li>
                  <button
                    type="button"
                    onClick={() => {
                      setSpaceMenuOpen(false);
                      setRenamingSpace(true);
                    }}
                    className="w-full px-3 py-1.5 text-left text-sm normal-case tracking-normal hover:bg-surface-2"
                  >
                    スペースの名前を変更
                  </button>
                </li>
              </ul>
            )}
          </>
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

          {!state?.loading && !state?.error && !tree?.pages.length && !hiddenAtRoot && (
            <p className="px-2 py-1 text-xs text-[var(--color-text-muted)]">
              {archivedMode ? 'アーカイブしたページはありません' : 'ページがありません'}
            </p>
          )}

          {tree && tree.pages.length > 0 && (
            <KbTreeList
              nodes={tree.pages}
              depth={0}
              parentId={null}
              hasHiddenChildren={tree.hasHiddenChildren}
              expandedPageIds={expandedPageIds}
              activePageId={activePageId}
              workspaceSlug={workspaceSlug}
              renamingPageId={renamingPageId}
              draggingPageId={draggingPageId}
              dropAt={dropAt}
              archivedMode={archivedMode}
              label={`${space.name} のページ`}
              onToggle={onTogglePage}
              onStartRename={setRenamingPageId}
              onCancelRename={() => setRenamingPageId(null)}
              onCommitRename={commitRename}
              onCreateChild={(parentId) => void createPage(parentId)}
              onArchive={(pageId) => void archivePage(pageId)}
              onUnarchive={(pageId) => void unarchivePage(pageId)}
              onMove={(pageId, target) => void movePage(pageId, target)}
              onDragStart={setDraggingPageId}
              onDragEnd={endDrag}
              onDragOverRow={(pageId, zone) => setDropAt({ pageId, zone })}
              onDropOnRow={(pageId, zone) => void dropOnRow(pageId, zone)}
            />
          )}

          {/*
            スペース直下の印。ページが 1 枚も無いときはここが出す（一覧そのものを描かないため）。
            1 枚でもあれば一覧の末尾に出る。1 件も見えないスペースでは backend が必ず false を返す。
          */}
          {!tree?.pages.length && hiddenAtRoot && (
            <p className="px-2 py-0.5 text-xs text-[var(--color-text-muted)]">
              表示できないページがあります
            </p>
          )}
        </div>
      )}
    </section>
  );
}
