import { useNavigate } from 'react-router-dom';
import { NoteWorkspaceSwitcher, useWorkspaceList } from '@/entities/note';

/**
 * HeaderWorkspaceSwitcher はヘッダーから所属ワークスペースを切り替える入口。
 *
 * 「いま開いている」状態はページ（/notes・/p/:id）側だけが持っており、ヘッダーとは
 * 共有していない。選んだ先はナビゲーションの state で /notes へ渡し、初期表示だけに使う
 * （以後の切替はサイドバー側の状態が正）。
 */
export default function HeaderWorkspaceSwitcher() {
  const navigate = useNavigate();
  const { workspaces, loading, createWorkspace, deleteWorkspace } = useWorkspaceList();

  if (loading && workspaces.length === 0) return null;
  if (workspaces.length === 0) return null;

  return (
    <div className="hidden w-44 flex-shrink-0 md:block">
      <NoteWorkspaceSwitcher
        workspaces={workspaces}
        activeSlug={null}
        onSelect={(slug) => navigate('/notes', { state: { workspaceSlug: slug } })}
        onCreate={async (input) => {
          const workspace = await createWorkspace(input);
          navigate('/notes', { state: { workspaceSlug: workspace.slug } });
        }}
        onDelete={deleteWorkspace}
      />
    </div>
  );
}
