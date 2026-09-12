import { useParams, useNavigate } from 'react-router-dom';
import { Bars3Icon } from '@heroicons/react/24/outline';
import { KbSidebar } from '@/widgets/kb-sidebar';
import { SecondaryPanel } from '@/widgets/secondary-panel';
import { useMobilePanelState } from '@/shared/lib/hooks/useMobilePanelState';
import { useKbSpaceEntry, KbSpaceTabs } from '@/entities/kb';

const ROLE_LABEL: Record<string, string> = {
  admin: '管理者',
  editor: '編集者',
  commenter: 'コメント可',
  viewer: '閲覧者',
};

/**
 * KbSpaceOverviewPage はスペースの「概要」画面（段14）。
 *
 * 凝った内容は作らない — 見本にこの画面の詳細までは無いため、自分の役割程度を示す
 * 最小限にとどめる（Storybook で目視して明らかにおかしい箇所だけ後で直す）。
 */
export default function KbSpaceOverviewPage() {
  const { spaceId } = useParams<{ spaceId?: string }>();
  const navigate = useNavigate();
  const { isOpen: mobilePanelOpen, open: openMobilePanel, close: closeMobilePanel } = useMobilePanelState();

  const { workspaceSlug, space, noSpaces, loading, error } = useKbSpaceEntry(spaceId, (id) =>
    navigate(`/kb/spaces/${id}`, { replace: true }),
  );

  if (error) {
    return (
      <div className="flex h-full items-center justify-center px-6 text-center text-sm text-[var(--color-text-muted)]">
        {error}
      </div>
    );
  }

  if (noSpaces) {
    return (
      <div className="flex h-full items-center justify-center px-6 text-center">
        <div>
          <p className="mb-1 text-base font-semibold text-[var(--color-text-secondary)]">
            アクセスできるスペースがありません
          </p>
          <p className="text-sm text-[var(--color-text-muted)]">ナレッジでスペースを作ると使えるようになります。</p>
        </div>
      </div>
    );
  }

  if (loading || !space || !workspaceSlug) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-[var(--color-text-muted)]">
        読み込み中…
      </div>
    );
  }

  return (
    <div className="flex h-full">
      <SecondaryPanel title="ナレッジ" peekable storageKey="frestyle.panel.note" mobileOpen={mobilePanelOpen} onMobileClose={closeMobilePanel}>
        <KbSidebar workspaceSlug={workspaceSlug} spaceId={space.id} />
      </SecondaryPanel>

      <main className="flex min-w-0 flex-1 flex-col overflow-hidden">
        <div className="flex items-center border-b border-surface-3 bg-surface-1 px-4 py-2 md:hidden">
          <button type="button" onClick={openMobilePanel} aria-label="ナレッジを開く" className="p-1">
            <Bars3Icon className="h-5 w-5" aria-hidden="true" />
          </button>
        </div>

        <KbSpaceTabs space={space} active="overview" />

        <div className="min-h-0 flex-1 overflow-y-auto px-6 py-8">
          <p className="text-sm text-[var(--color-text-tertiary)]">
            このスペースでの自分の役割: {ROLE_LABEL[space.role] ?? space.role}
          </p>
        </div>
      </main>
    </div>
  );
}
