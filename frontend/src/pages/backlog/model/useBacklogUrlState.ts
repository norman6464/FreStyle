import { useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';

export type BacklogTab = 'tickets' | 'statuses' | 'types';

const TABS: readonly BacklogTab[] = ['tickets', 'statuses', 'types'];

/** 既定の面。URL に出さない（既定を省くと URL が短く保てる）。 */
const DEFAULT_TAB: BacklogTab = 'tickets';

function readTab(value: string | null): BacklogTab {
  return TABS.includes(value as BacklogTab) ? (value as BacklogTab) : DEFAULT_TAB;
}

/**
 * useBacklogUrlState は一覧の文脈（どの面か・アーカイブを見ているか・どのチケットを選んだか）を
 * URL に載せる。
 *
 * 画面の中に閉じた状態にすると、チケットを開いて戻ってきたときに絞り込みも選択も消える。
 * 戻る先が「現役の先頭」に固定されると、朝に何十件も捌く動きが成立しない。
 *
 * 履歴は汚さない（`replace`）。面の切り替えや行の選択で戻る操作の回数が増えると、
 * 「戻る」でバックログから出るのに何度も押すことになるため。
 */
export function useBacklogUrlState() {
  const [params, setParams] = useSearchParams();

  const tab = readTab(params.get('tab'));
  const archived = params.get('archived') === '1';
  const selectedId = params.get('ticket');

  const update = useCallback(
    (patch: { tab?: BacklogTab; archived?: boolean; selectedId?: string | null }) => {
      setParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          if (patch.tab !== undefined) {
            if (patch.tab === DEFAULT_TAB) next.delete('tab');
            else next.set('tab', patch.tab);
          }
          if (patch.archived !== undefined) {
            if (patch.archived) next.set('archived', '1');
            else next.delete('archived');
          }
          if (patch.selectedId !== undefined) {
            if (patch.selectedId) next.set('ticket', patch.selectedId);
            else next.delete('ticket');
          }
          return next;
        },
        { replace: true },
      );
    },
    [setParams],
  );

  const setTab = useCallback((value: BacklogTab) => update({ tab: value }), [update]);
  const setArchived = useCallback((value: boolean) => update({ archived: value }), [update]);
  const selectTicket = useCallback((value: string | null) => update({ selectedId: value }), [update]);

  /** スペースを移ったときに前のスペースの文脈を持ち越さない。 */
  const reset = useCallback(
    () => update({ tab: DEFAULT_TAB, archived: false, selectedId: null }),
    [update],
  );

  return { tab, archived, selectedId, setTab, setArchived, selectTicket, reset };
}
