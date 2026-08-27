import { ChevronRightIcon } from '@heroicons/react/24/outline';
import { flattenKbTree, type KbSpace } from '@/entities/knowledge-base';
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
}

/**
 * KbSpaceSection はスペース 1 つ分の見出しと、その配下のページの木。
 *
 * スペースは**同時に複数見たい**（自分のページと部署のページを行き来する）ので、
 * 切り替えではなく見出しとして並べる。ワークスペースは会社の境界で同時に見る場面が無いので、
 * あちらは切り替えにしてある。同時に見たいものは並べ、排他のものは切り替える。
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
}: KbSpaceSectionProps) {
  const open = state?.open ?? false;
  const entries = state?.tree ? flattenKbTree(state.tree.pages, expandedPageIds) : [];
  const hiddenAtRoot = state?.tree?.hiddenChildCount ?? 0;

  return (
    <section className="mb-1">
      <h2>
        <button
          type="button"
          onClick={() => onToggleSpace(space.id)}
          aria-expanded={open}
          className="flex w-full items-center gap-1 rounded-md px-1 py-1 text-left text-xs font-semibold uppercase tracking-wide text-[var(--color-text-muted)] hover:bg-surface-2"
        >
          <ChevronRightIcon
            className={`h-3 w-3 shrink-0 transition-transform ${open ? 'rotate-90' : ''}`}
            aria-hidden="true"
          />
          <span className="truncate">{space.name}</span>
        </button>
      </h2>

      {open && (
        <div className="mt-0.5">
          {state?.loading && (
            <p className="px-2 py-1 text-xs text-[var(--color-text-muted)]">読み込み中…</p>
          )}

          {/*
            失敗は必ず出し、やり直せるようにする。黙って空にすると「ページが 1 枚も無い」と
            見分けが付かず、利用者は消えたと思う。このリポジトリには
            「失敗しても成功の表示が出る」轍が既にあるので、同じ形を作らない。
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

          {!state?.loading && !state?.error && entries.length === 0 && hiddenAtRoot === 0 && (
            <p className="px-2 py-1 text-xs text-[var(--color-text-muted)]">ページがありません</p>
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
                  />
                ) : (
                  <KbHiddenChildrenRow
                    key={`hidden-${entry.parentId}`}
                    depth={entry.depth}
                    count={entry.count}
                  />
                ),
              )}
            </ul>
          )}

          {/* スペース直下で伏せた件数。1 件も見えないスペースでは backend が必ず 0 を返す。 */}
          {hiddenAtRoot > 0 && (
            <p className="px-2 py-0.5 text-xs text-[var(--color-text-muted)]">
              {hiddenAtRoot} ページは表示できません
            </p>
          )}
        </div>
      )}
    </section>
  );
}
