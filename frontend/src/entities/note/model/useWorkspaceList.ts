import { useCallback, useEffect, useState } from 'react';
import NoteRepository from '../api/noteRepository';
import { emitNoteTreeEvent, subscribeNoteTreeEvents } from './noteTreeEvents';
import type { NoteWorkspace } from './types';

/**
 * useWorkspaceList は所属ワークスペースの一覧・作成・削除だけを扱う軽量な hook。
 *
 * widgets/note-sidebar の useNoteTree はスペース・ページの木まで抱える重い hook なので、
 * ヘッダーのようにワークスペースの出入りだけが要る場所ではこちらを使う。
 *
 * 作成・削除は noteTreeEvents で他インスタンスへ知らせ、他インスタンス（NoteSidebar の
 * useNoteTree・他画面の useWorkspaceList）からの通知も購読する。SecondaryPanel が
 * モバイル用/デスクトップ用の DOM を常に両方マウントするため、同じ画面内でも
 * ワークスペース一覧を持つインスタンスは複数存在し、片方の変更を他方が自動では知れない。
 */
export function useWorkspaceList() {
  const [workspaces, setWorkspaces] = useState<NoteWorkspace[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(() => {
    setLoading(true);
    setError(null);
    NoteRepository.fetchWorkspaces()
      .then((list) => setWorkspaces(list))
      .catch(() => setError('ワークスペースを読み込めませんでした'))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    return subscribeNoteTreeEvents((event) => {
      if (event.type === 'workspace-created') {
        setWorkspaces((prev) => (prev.some((w) => w.slug === event.workspace.slug) ? prev : [...prev, event.workspace]));
        return;
      }
      if (event.type === 'workspace-deleted') {
        setWorkspaces((prev) => prev.filter((w) => w.slug !== event.workspaceSlug));
      }
    });
  }, []);

  const createWorkspace = useCallback(async (input: { name: string }): Promise<NoteWorkspace> => {
    const workspace = await NoteRepository.createWorkspace(input);
    setWorkspaces((prev) => [...prev, workspace]);
    emitNoteTreeEvent({ type: 'workspace-created', workspace });
    return workspace;
  }, []);

  const deleteWorkspace = useCallback(async (slug: string): Promise<void> => {
    await NoteRepository.deleteWorkspace(slug);
    setWorkspaces((prev) => prev.filter((w) => w.slug !== slug));
    emitNoteTreeEvent({ type: 'workspace-deleted', workspaceSlug: slug });
  }, []);

  return { workspaces, loading, error, retry: load, createWorkspace, deleteWorkspace };
}
