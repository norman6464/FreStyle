import { useCallback, useEffect, useState } from 'react';
import NoteRepository from '../api/noteRepository';
import type { NoteWorkspace } from './types';

/**
 * useWorkspaceList は所属ワークスペースの一覧・作成・削除だけを扱う軽量な hook。
 *
 * widgets/note-sidebar の useNoteTree はスペース・ページの木まで抱える重い hook なので、
 * ヘッダーのようにワークスペースの出入りだけが要る場所ではこちらを使う。
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

  const createWorkspace = useCallback(async (input: { name: string }): Promise<NoteWorkspace> => {
    const workspace = await NoteRepository.createWorkspace(input);
    setWorkspaces((prev) => [...prev, workspace]);
    return workspace;
  }, []);

  const deleteWorkspace = useCallback(async (slug: string): Promise<void> => {
    await NoteRepository.deleteWorkspace(slug);
    setWorkspaces((prev) => prev.filter((w) => w.slug !== slug));
  }, []);

  return { workspaces, loading, error, retry: load, createWorkspace, deleteWorkspace };
}
