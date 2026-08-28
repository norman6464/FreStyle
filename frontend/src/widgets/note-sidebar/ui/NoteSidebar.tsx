import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArchiveBoxIcon, MagnifyingGlassIcon, PlusIcon } from '@heroicons/react/24/outline';
import { useToast } from '@/shared/lib/hooks/useToast';
import NoteCreateForm from './NoteCreateForm';
import { useNoteTree } from '../model/useNoteTree';
import NoteWorkspaceSwitcher from './NoteWorkspaceSwitcher';
import NoteSpaceSection from './NoteSpaceSection';
import NoteSearchDialog from './NoteSearchDialog';
import { NoteRepository, type NotePage } from '@/entities/note';

export interface NoteSidebarProps {
  /** URL が指しているワークスペース。未指定なら所属の先頭を開く。 */
  workspaceSlug?: string;
  /** URL が指しているページ。現在位置の強調と、祖先の自動展開に使う。 */
  activePageId?: string;
}

/**
 * NoteSidebar はナレッジ基盤の「場所を示す面」。
 *
 * 上から ワークスペースの切替 → スペースの見出し → ページの木。
 * いまは読むだけで、作る・動かす・印は次の段で足す。
 *
 * ここに置かないもの: 更新日時・作成者。サイドバーは場所を示す面であって、
 * 属性を並べる面ではない。行に情報を足すほど、木の形そのものが読みにくくなる。
 */
export default function NoteSidebar({ workspaceSlug, activePageId }: NoteSidebarProps) {
  const navigate = useNavigate();
  const { showToast } = useToast();
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
    createWorkspace,
    createSpace,
    renameSpace,
    selectWorkspace,
    createPage,
    renamePage,
    archivePage,
    unarchivePage,
    movePage,
    archivedMode,
    setArchivedMode,
  } = useNoteTree({ workspaceSlug, activePageId });

  // 題名の検索。実体はサーバー（ツリーと同じ規則で、閲覧できるページだけが返る）。
  // 検索モーダルの開閉。検索そのもの（デバウンス・世代番号・再試行）はモーダルが持つ。
  // サイドバー本体は場所（木）を示すことに徹する — 常設の入力欄が場所の面を圧迫し、
  // 木と結果が同じ狭い面で入れ替わる形は見本合わせでやめた（設計 artifact 参照）。
  const [searchOpen, setSearchOpen] = useState(false);
  // スペース追加フォームの開閉。既にスペースがあるときの追加入口（0 件のときは常設フォーム）。
  const [addingSpace, setAddingSpace] = useState(false);

  return (
    <nav aria-label="ナレッジ基盤" className="flex h-full flex-col overflow-y-auto p-2">
      <NoteWorkspaceSwitcher
        workspaces={workspaces}
        activeSlug={activeSlug}
        // URL はワークスペースを持たない（/p/{pageId} だけ）。切り替えは画面の状態で、
        // ページを開けば URL から場所が確定する。
        onSelect={(slug) => {
          selectWorkspace(slug);
          navigate('/notes');
        }}
        onCreate={async (input) => {
          try {
            await createWorkspace(input);
          } catch {
            showToast('error', 'ワークスペースを作成できませんでした');
            throw new Error('create workspace failed');
          }
          // 切替（onSelect）と同じ理由で一覧へ戻す。戻さないと、開いていた旧ワーク
          // スペースのページと、新ワークスペースを指すサイドバーが食い違ったまま残る。
          navigate('/notes');
        }}
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
        // 所属が無いと API は全部 404 になる。「壊れている」ではなく「ここから始める」と伝える。
        //
        // **入口をここに置くのが要点。** 作る手段が無いと、ワークスペースを作る API が
        // あってもサイドバーには永久にたどり着けない（実際そうなっていた）。
        <div>
          <p className="px-2 pt-3 text-xs leading-relaxed text-[var(--color-text-muted)]">
            まだワークスペースがありません。作るとページを置けるようになります。
          </p>
          <NoteCreateForm
            what="ワークスペース"
            onCreate={async (input) => {
              try {
                await createWorkspace(input);
              } catch {
                showToast('error', 'ワークスペースを作成できませんでした');
                throw new Error('create workspace failed');
              }
            }}
          />
        </div>
      )}

      {/* スペースが 0 件でも出す。スペースは 1 つも見えないが、個別に許可された
          ページだけ持つ人がいる — その人にとって検索が唯一の入口になる。 */}
      {activeSlug && !archivedMode && (
        <button
          type="button"
          onClick={() => setSearchOpen(true)}
          className="mt-2 flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-xs text-[var(--color-text-muted)] transition-colors hover:bg-surface-2"
        >
          <MagnifyingGlassIcon className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
          <span>検索</span>
        </button>
      )}

      <div className="mt-2 min-h-0 flex-1">
        {/* 節の見出し（見本合わせ）。スペースは共有の木で、個人の領域（プライベート）は
            権限モデルの設計とセットで別の段。 */}
        {activeSlug && !archivedMode && spaces.length > 0 && (
          <p className="px-2 pb-1 text-[11px] font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
            チームスペース
          </p>
        )}
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
          // ワークスペースを作っただけではスペースは付いてこない。ここでも入口を出す
          // （出さないと「見られるスペースがありません」で行き止まりになる）。
          <div>
            <p className="px-2 pt-1 text-xs leading-relaxed text-[var(--color-text-muted)]">
              まだスペースがありません。部署や個人ごとの区画を作ります。
            </p>
            <NoteCreateForm
              what="スペース"
              onCreate={async (input) => {
                try {
                  await createSpace(input);
                } catch {
                  showToast('error', 'スペースを作成できませんでした');
                  throw new Error('create space failed');
                }
              }}
            />
          </div>
        )}

        {activeSlug &&
          spaces.map((space) => (
            <NoteSpaceSection
              key={space.id}
              space={space}
              state={spaceStates[space.id]}
              workspaceSlug={activeSlug}
              activePageId={activePageId}
              expandedPageIds={expandedPageIds}
              onToggleSpace={toggleSpace}
              onTogglePage={togglePage}
              onRetry={retrySpace}
              onCreatePage={createPage}
              onRenamePage={renamePage}
              onArchivePage={archivePage}
              onUnarchivePage={unarchivePage}
              archivedMode={archivedMode}
              onRenameSpace={renameSpace}
              onMovePage={movePage}
            />
          ))}

        {/*
          スペースの追加入口。0 件のときの常設フォームとは別に、既にあるときも
          ここから増やせる（無いと 1 つ作った時点で追加の手段が UI から消える）。
          作成できるかの判定はサーバーが持つ — 押せても、権限が無ければ失敗の知らせが出る。
        */}
        {activeSlug && !archivedMode && spaces.length > 0 && !spacesLoading && (
          addingSpace ? (
            <div className="mt-1 rounded-md border border-surface-3">
              <NoteCreateForm
                what="スペース"
                onCreate={async (input) => {
                  try {
                    await createSpace(input);
                  } catch {
                    showToast('error', 'スペースを作成できませんでした');
                    throw new Error('create space failed');
                  }
                  setAddingSpace(false);
                }}
              />
              <button
                type="button"
                onClick={() => setAddingSpace(false)}
                className="w-full px-2 pb-2 text-left text-xs text-[var(--color-text-muted)] hover:underline"
              >
                やめる
              </button>
            </div>
          ) : (
            <button
              type="button"
              onClick={() => setAddingSpace(true)}
              className="mt-1 flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-xs text-[var(--color-text-muted)] transition-colors hover:bg-surface-2"
            >
              <PlusIcon className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
              <span>スペースを追加</span>
            </button>
          )
        )}
      </div>

      {/*
        アーカイブは下段の入口に畳む。普段は視界に入らず、押したときだけ木を置き換える。
        切り替えはワークスペース全体で 1 つ — スペースごとに持たせると
        「いまどちらを見ているのか」が場所によって変わる。
      */}
      {searchOpen && activeSlug && (
        <NoteSearchDialog
          workspaceSlug={activeSlug}
          spaces={spaces}
          onClose={() => setSearchOpen(false)}
        />
      )}

      {activeSlug && (
        <button
          type="button"
          onClick={() => setArchivedMode(!archivedMode)}
          aria-pressed={archivedMode}
          // 見た目は「アーカイブ」の一語だが、読み上げ名は押すと何が起きるかにする。
          // 行のメニューにも「アーカイブ」があるので、名前が同じだとどちらか分からない。
          aria-label={archivedMode ? '現役のページに戻る' : 'アーカイブしたページを表示'}
          className={`mt-2 flex shrink-0 items-center gap-1.5 rounded-md px-2 py-1.5 text-xs transition-colors ${
            archivedMode
              ? 'bg-brand-500/10 text-brand-600'
              : 'text-[var(--color-text-muted)] hover:bg-surface-2'
          }`}
        >
          <ArchiveBoxIcon className="h-4 w-4 shrink-0" aria-hidden="true" />
          <span>アーカイブ</span>
        </button>
      )}
    </nav>
  );
}
