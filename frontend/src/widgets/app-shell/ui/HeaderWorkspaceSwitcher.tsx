import { useNavigate } from 'react-router-dom';
import { KbWorkspaceSwitcher, useWorkspaceList } from '@/entities/kb';
import { useToast } from '@/shared/lib/hooks/useToast';

/**
 * HeaderWorkspaceSwitcher はヘッダーから所属ワークスペースを切り替える入口。
 *
 * 「いま開いている」状態はページ（/kb・/kb/:id）側だけが持っており、ヘッダーとは
 * 共有していない。選んだ先はナビゲーションの state で /kb へ渡す — resolveEntryPageId
 * （pages/kb/model/resolveEntryPage.ts）がそれを見て、切り替え先の最初のページへ移る
 * （以後の切替はサイドバー側の状態が正）。
 */
export default function HeaderWorkspaceSwitcher() {
  const navigate = useNavigate();
  const { showToast } = useToast();
  const { workspaces, loading, createWorkspace, deleteWorkspace } = useWorkspaceList();

  if (loading && workspaces.length === 0) return null;
  if (workspaces.length === 0) return null;

  return (
    <div className="hidden w-44 flex-shrink-0 md:block">
      <KbWorkspaceSwitcher
        workspaces={workspaces}
        activeSlug={null}
        onSelect={(slug) => navigate('/kb', { state: { workspaceSlug: slug } })}
        onCreate={async (input) => {
          try {
            const workspace = await createWorkspace(input);
            navigate('/kb', { state: { workspaceSlug: workspace.slug } });
          } catch {
            showToast('error', 'ワークスペースを作成できませんでした');
            throw new Error('create workspace failed');
          }
        }}
        onDelete={async (slug) => {
          try {
            await deleteWorkspace(slug);
          } catch {
            showToast('error', 'ワークスペースを削除できませんでした');
          }
        }}
      />
    </div>
  );
}
