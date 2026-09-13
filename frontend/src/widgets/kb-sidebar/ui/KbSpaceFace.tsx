import { useState } from 'react';
import { Link } from 'react-router-dom';
import { EllipsisHorizontalIcon, PlusIcon } from '@heroicons/react/24/outline';
import { useToast } from '@/shared/lib/hooks/useToast';
import type { KbPage, KbSpace } from '@/entities/kb';
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
}

/**
 * KbSpaceFace は「今いるスペース」の顔（段14）。KbSpaceSection の見出し部分を引き継ぐが、
 * 開閉トグルは無い（サイドバーが表示するのは常にこの 1 スペースのため）。代わりに
 * 固定ナビ 4 項目（概要・すべてのページ・お気に入り・メンバー）を持つ。
 *
 * 他のスペースへの切替はヘッダーの「スペース ▾」（HeaderSpacesNav）が持つため、
 * ここには無い。
 */
export default function KbSpaceFace({
  space,
  workspaceSlug,
  workspaceCanManage,
  archivedMode,
  onCreatePage,
  onCreatedFromTemplate,
  onRenameSpace,
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
