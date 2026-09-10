import { useCallback, useEffect, useRef, useState } from 'react';
import { TicketRepository, type Label, type LabelInput } from '@/entities/ticket';

const LOAD_FAILED = 'ラベルを読み込めませんでした。時間をおいて開き直すと最新の状態が出ます。';

/**
 * useTicketLabels はスペースのラベル定義（一覧・作成・改名・削除）を読み書きする。
 *
 * 状態・種別のマスタ（useTicketMasters）と違い、使用中件数のような取得し直さないと
 * ずれる派生値を持たないので、作成・更新・削除の応答をそのまま手元へ反映する
 * （マスタのように毎回一覧を取り直さない）。
 *
 * チケットへの付け外し（ticket.labels の更新）はここでは持たない — 対象がチケット
 * 1 件の状態（useTicketList / useTicketPage）に属するため、それぞれの hook に持たせる。
 */
export function useTicketLabels(workspaceSlug: string | undefined, spaceId: string | undefined) {
  const [labels, setLabels] = useState<Label[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const active = useRef<string | null>(null);

  const load = useCallback(async (slug: string, space: string) => {
    const key = `${slug} ${space}`;
    setLoading(true);
    setError(null);
    try {
      const list = await TicketRepository.fetchLabels(slug, space);
      if (active.current !== key) return;
      setLabels(list);
    } catch {
      if (active.current !== key) return;
      setError(LOAD_FAILED);
    } finally {
      if (active.current === key) setLoading(false);
    }
  }, []);

  useEffect(() => {
    const key = workspaceSlug && spaceId ? `${workspaceSlug} ${spaceId}` : null;
    active.current = key;
    if (!workspaceSlug || !spaceId) {
      setLabels([]);
      return;
    }
    void load(workspaceSlug, spaceId);
  }, [workspaceSlug, spaceId, load]);

  const refresh = useCallback(() => {
    if (workspaceSlug && spaceId) void load(workspaceSlug, spaceId);
  }, [workspaceSlug, spaceId, load]);

  const requireScope = useCallback((): [string, string] => {
    if (!workspaceSlug || !spaceId) throw new Error('backlog: no active scope');
    return [workspaceSlug, spaceId];
  }, [workspaceSlug, spaceId]);

  const createLabel = useCallback(
    async (input: LabelInput) => {
      const [slug, space] = requireScope();
      const created = await TicketRepository.createLabel(slug, space, input);
      if (active.current === `${slug} ${space}`) setLabels((prev) => [...prev, created]);
      return created;
    },
    [requireScope],
  );

  const updateLabel = useCallback(
    async (labelId: string, input: LabelInput) => {
      const [slug, space] = requireScope();
      const updated = await TicketRepository.updateLabel(slug, space, labelId, input);
      if (active.current === `${slug} ${space}`) {
        setLabels((prev) => prev.map((l) => (l.id === labelId ? updated : l)));
      }
      return updated;
    },
    [requireScope],
  );

  const deleteLabel = useCallback(
    async (labelId: string) => {
      const [slug, space] = requireScope();
      await TicketRepository.deleteLabel(slug, space, labelId);
      if (active.current === `${slug} ${space}`) setLabels((prev) => prev.filter((l) => l.id !== labelId));
    },
    [requireScope],
  );

  return { labels, loading, error, refresh, createLabel, updateLabel, deleteLabel };
}
