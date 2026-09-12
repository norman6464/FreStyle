import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  ChevronUpDownIcon,
  EllipsisHorizontalIcon,
  FolderIcon,
  PlusIcon,
} from '@heroicons/react/24/outline';
import { useToast } from '@/shared/lib/hooks/useToast';
import { NameCreateForm } from '@/shared/ui';
import { KbRepository, type KbMySpace, type KbPage, type KbSpace } from '@/entities/kb';
import KbInlineRename from './KbInlineRename';
import { useKbPageTemplates } from '../model/useKbPageTemplates';
import KbTemplatePickerModal from './KbTemplatePickerModal';

export interface KbSpaceFaceProps {
  space: KbSpace;
  workspaceSlug: string;
  workspaceCanManage: boolean;
  archivedMode: boolean;
  /** ページをスペース直下に作る（作成後の題名入力・遷移は呼び出し側が持つ）。 */
  onCreatePage: () => void;
  /** テンプレートから作った直後のページへの遷移・木への反映は呼び出し側が持つ。 */
  onCreatedFromTemplate: (page: KbPage) => void;
  onRenameSpace: (name: string) => Promise<KbSpace>;
  /**
   * スペースを作る（切替ドロップダウンの下部から）。useKbTree の createSpace をそのまま
   * 渡す — visibility の食い違いを検査する規則を、ここで二重に持たないため。
   */
  onCreateSpace: (input: { name: string; visibility?: 'workspace' | 'private' }) => Promise<KbSpace>;
}

/**
 * KbSpaceFace は「今いるスペース」の顔（段14）。KbSpaceSection の見出し部分を引き継ぐが、
 * 開閉トグルは無い（サイドバーが表示するのは常にこの 1 スペースのため）。代わりに
 * スペース切替（一時的な繋ぎ — W3 でヘッダーの「スペース▾」に正式に移すまで）と、
 * 固定ナビ 4 項目（概要・すべてのページ・お気に入り・メンバー）を持つ。
 */
export default function KbSpaceFace({
  space,
  workspaceSlug,
  workspaceCanManage,
  archivedMode,
  onCreatePage,
  onCreatedFromTemplate,
  onRenameSpace,
  onCreateSpace,
}: KbSpaceFaceProps) {
  const { showToast } = useToast();
  const [renaming, setRenaming] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [templatePickerOpen, setTemplatePickerOpen] = useState(false);
  const templates = useKbPageTemplates(workspaceSlug, space.id, templatePickerOpen);

  const commitRename = async (name: string) => {
    try {
      await onRenameSpace(name);
      setRenaming(false);
    } catch {
      showToast('error', 'スペースの名前を変更できませんでした');
      // 入力欄は開いたままにする（閉じると書いた文字が消えるが、元の題名は保存されていない）。
      throw new Error('rename space failed');
    }
  };

  const createFromTemplate = async (templateId: string, title: string) => {
    const created = await templates.createPageFromTemplate({ templateId, title });
    setTemplatePickerOpen(false);
    onCreatedFromTemplate(created);
  };

  return (
    <div className="mb-2">
      <div className="group relative flex items-center gap-1 rounded-md pr-1 hover:bg-surface-2">
        {renaming ? (
          <div className="flex min-w-0 flex-1 items-center gap-1 px-1 py-1.5">
            <KbInlineRename
              initialTitle={space.name}
              ariaLabel="スペースの名前"
              onCommit={commitRename}
              onCancel={() => setRenaming(false)}
            />
          </div>
        ) : (
          <span className="min-w-0 flex-1 truncate px-1 py-1.5 text-sm font-semibold text-[var(--color-text-primary)]">
            {space.name}
          </span>
        )}
        {!archivedMode && !renaming && (
          <>
            <button
              type="button"
              onClick={onCreatePage}
              aria-label={`${space.name} にページを追加`}
              title="ページを追加"
              className="shrink-0 rounded p-1 text-[var(--color-text-muted)] transition-colors hover:bg-surface-3"
            >
              <PlusIcon className="h-4 w-4" aria-hidden="true" />
            </button>
            <button
              type="button"
              onClick={() => setMenuOpen((prev) => !prev)}
              aria-expanded={menuOpen}
              aria-label={`${space.name} の操作`}
              className="shrink-0 rounded p-1 text-[var(--color-text-muted)] transition-colors hover:bg-surface-3"
            >
              <EllipsisHorizontalIcon className="h-4 w-4" aria-hidden="true" />
            </button>
            {menuOpen && (
              <ul className="absolute right-0 top-full z-20 mt-1 w-44 rounded-lg border border-surface-3 bg-surface-1 py-1 shadow-lg">
                <li>
                  <button
                    type="button"
                    onClick={() => {
                      setMenuOpen(false);
                      setRenaming(true);
                    }}
                    className="w-full px-3 py-1.5 text-left text-sm text-[var(--color-text-primary)] hover:bg-surface-2"
                  >
                    スペースの名前を変更
                  </button>
                </li>
                <li>
                  <button
                    type="button"
                    onClick={() => {
                      setMenuOpen(false);
                      setTemplatePickerOpen(true);
                    }}
                    className="w-full px-3 py-1.5 text-left text-sm text-[var(--color-text-primary)] hover:bg-surface-2"
                  >
                    雛形から作る
                  </button>
                </li>
              </ul>
            )}
          </>
        )}
      </div>

      {!archivedMode && (
        <KbSpaceSwitcher workspaceSlug={workspaceSlug} activeSpaceId={space.id} onCreateSpace={onCreateSpace} />
      )}

      {!archivedMode && (
        <nav aria-label={`${space.name} の画面`} className="mt-1 flex flex-col">
          <Link
            to={`/kb/spaces/${space.id}`}
            className="rounded-md px-2 py-1 text-xs text-[var(--color-text-muted)] transition-colors hover:bg-surface-2"
          >
            概要
          </Link>
          <Link
            to={`/kb/spaces/${space.id}/pages`}
            className="rounded-md px-2 py-1 text-xs text-[var(--color-text-muted)] transition-colors hover:bg-surface-2"
          >
            すべてのページ
          </Link>
          <Link
            to={`/kb/spaces/${space.id}/favorites`}
            className="rounded-md px-2 py-1 text-xs text-[var(--color-text-muted)] transition-colors hover:bg-surface-2"
          >
            お気に入り
          </Link>
          <Link
            to={`/kb/spaces/${space.id}/members`}
            className="rounded-md px-2 py-1 text-xs text-[var(--color-text-muted)] transition-colors hover:bg-surface-2"
          >
            メンバー
          </Link>
        </nav>
      )}

      <KbTemplatePickerModal
        isOpen={templatePickerOpen}
        templates={templates.templates}
        loading={templates.loading}
        error={templates.error}
        canManageTemplates={workspaceCanManage}
        onConfirm={createFromTemplate}
        onDelete={templates.deleteTemplate}
        onClose={() => setTemplatePickerOpen(false)}
      />
    </div>
  );
}

/**
 * KbSpaceSwitcher は他のスペースへ移る一時的な繋ぎ（W3 でヘッダーの「スペース▾」に
 * 正式に移すまで）。開いたときだけ自分がアクセスできるスペース一覧（/me/spaces）を取る。
 */
function KbSpaceSwitcher({
  workspaceSlug,
  activeSpaceId,
  onCreateSpace,
}: {
  workspaceSlug: string;
  activeSpaceId: string;
  onCreateSpace: (input: { name: string; visibility?: 'workspace' | 'private' }) => Promise<KbSpace>;
}) {
  const navigate = useNavigate();
  const { showToast } = useToast();
  const [open, setOpen] = useState(false);
  const [mySpaces, setMySpaces] = useState<KbMySpace[] | null>(null);
  const [addingSpace, setAddingSpace] = useState(false);
  const [addingPrivateSpace, setAddingPrivateSpace] = useState(false);

  useEffect(() => {
    if (!open) return;
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

  const createSpace = async (input: { name: string; visibility?: 'workspace' | 'private' }) => {
    try {
      const space = await onCreateSpace(input);
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
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        aria-expanded={open}
        aria-label="スペースを切り替える"
        className="mb-1 flex w-full items-center gap-1 rounded-md px-1 py-1 text-xs text-[var(--color-text-muted)] transition-colors hover:bg-surface-2"
      >
        <ChevronUpDownIcon className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
        <span>スペースを切り替える</span>
      </button>
      {open && (
        <div className="absolute left-0 top-full z-20 w-56 rounded-lg border border-surface-3 bg-surface-1 py-1 shadow-lg">
          {mySpaces === null && (
            <p className="px-3 py-1.5 text-xs text-[var(--color-text-muted)]">読み込み中…</p>
          )}
          {mySpaces?.map((s) => (
            <Link
              key={s.id}
              to={`/kb/spaces/${s.id}`}
              onClick={() => setOpen(false)}
              className={`flex items-center gap-1.5 px-3 py-1.5 text-sm hover:bg-surface-2 ${
                s.id === activeSpaceId ? 'font-semibold text-[var(--color-text-primary)]' : 'text-[var(--color-text-secondary)]'
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
                <NameCreateForm what="プライベートスペース" onCreate={(input) => createSpace({ ...input, visibility: 'private' })} />
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
