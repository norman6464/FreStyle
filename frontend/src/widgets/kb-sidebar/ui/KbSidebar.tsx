import { useNavigate } from 'react-router-dom';
import { useKnowledgeBaseTree } from '../model/useKnowledgeBaseTree';
import KbWorkspaceSwitcher from './KbWorkspaceSwitcher';
import KbSpaceSection from './KbSpaceSection';

export interface KbSidebarProps {
  /** URL が指しているワークスペース。未指定なら所属の先頭を開く。 */
  workspaceSlug?: string;
  /** URL が指しているページ。現在位置の強調と、祖先の自動展開に使う。 */
  activePageId?: string;
}

/**
 * KbSidebar はナレッジ基盤の「場所を示す面」。
 *
 * 上から ワークスペースの切替 → スペースの見出し → ページの木。
 * いまは読むだけで、作る・動かす・印は次の段で足す。
 *
 * ここに置かないもの: 更新日時・作成者。サイドバーは場所を示す面であって、
 * 属性を並べる面ではない。行に情報を足すほど、木の形そのものが読みにくくなる。
 */
export default function KbSidebar({ workspaceSlug, activePageId }: KbSidebarProps) {
  const navigate = useNavigate();
  const {
    workspaces,
    workspacesLoading,
    workspacesError,
    retryWorkspaces,
    activeSlug,
    spaces,
    spacesLoading,
    spacesError,
    retrySpaces,
    spaceStates,
    toggleSpace,
    retrySpace,
    expandedPageIds,
    togglePage,
  } = useKnowledgeBaseTree({ workspaceSlug, activePageId });

  return (
    <nav aria-label="ナレッジ基盤" className="flex h-full flex-col overflow-y-auto p-2">
      <KbWorkspaceSwitcher
        workspaces={workspaces}
        activeSlug={activeSlug}
        // 切り替えたら URL も合わせる。状態だけ変えると、再読み込みや共有で別の場所が開く。
        onSelect={(slug) => navigate(`/kb/${slug}`)}
      />

      {workspacesLoading && (
        <p className="px-2 py-2 text-xs text-[var(--color-text-muted)]">読み込み中…</p>
      )}
      {workspacesError && (
        <div className="px-2 py-2 text-xs text-red-600">
          <p>{workspacesError}</p>
          <button type="button" onClick={retryWorkspaces} className="mt-0.5 underline hover:no-underline">
            再試行
          </button>
        </div>
      )}

      {!workspacesLoading && !workspacesError && workspaces.length === 0 && (
        // 所属が無いと API は全部 404 になる。「壊れている」ではなく「まだ居ない」と伝える。
        <p className="px-2 py-4 text-xs leading-relaxed text-[var(--color-text-muted)]">
          所属しているワークスペースがありません。管理者に招待してもらってください。
        </p>
      )}

      <div className="mt-2 min-h-0 flex-1">
        {spacesLoading && (
          <p className="px-2 py-1 text-xs text-[var(--color-text-muted)]">読み込み中…</p>
        )}
        {spacesError && (
          <div className="px-2 py-1 text-xs text-red-600">
            <p>{spacesError}</p>
            <button type="button" onClick={retrySpaces} className="mt-0.5 underline hover:no-underline">
              再試行
            </button>
          </div>
        )}

        {!spacesLoading && !spacesError && activeSlug && spaces.length === 0 && (
          <p className="px-2 py-1 text-xs text-[var(--color-text-muted)]">
            見られるスペースがありません。
          </p>
        )}

        {activeSlug &&
          spaces.map((space) => (
            <KbSpaceSection
              key={space.id}
              space={space}
              state={spaceStates[space.id]}
              workspaceSlug={activeSlug}
              activePageId={activePageId}
              expandedPageIds={expandedPageIds}
              onToggleSpace={toggleSpace}
              onTogglePage={togglePage}
              onRetry={retrySpace}
            />
          ))}
      </div>
    </nav>
  );
}
