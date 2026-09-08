import { useCallback, useEffect, useRef, useState } from 'react';
import { KbRepository, type KbPageSuggestion } from '@/entities/kb';

export interface KbPageSuggestionsState {
  suggestions: KbPageSuggestion[];
  loading: boolean;
  /** 失敗の理由。null なら失敗していない。 */
  error: string | null;
}

const EMPTY: KbPageSuggestionsState = { suggestions: [], loading: false, error: null };

const LOAD_FAILED =
  '提案を読み込めませんでした。通信が切れたか、このページを見る立場でなくなっています。開き直すと最新の状態が出ます。';

/** 一覧取得の宛先。応答が着地してよいかの判定にこれを使う（useKbPageVersions と同じ形）。 */
interface SuggestionsTarget {
  key: string;
  workspaceSlug: string;
  pageId: string;
}

function targetOf(
  workspaceSlug: string | undefined,
  pageId: string | undefined,
  open: boolean,
): SuggestionsTarget | null {
  // パネルが開いている間だけ取りに行く（版一覧と同じ理由 — 常設のバッジを持たない）。
  if (!open || !workspaceSlug || !pageId) return null;
  return { key: `${workspaceSlug} ${pageId}`, workspaceSlug, pageId };
}

/**
 * useKbPageSuggestions はページ 1 枚の open な提案一覧と、採用・却下を持つ。
 *
 * 一覧の取得は open（パネルが開いているか）ゲート付き（useKbPageVersions と同じ理由）。
 * 採用・却下は useKbComments の resolve/reopen と同じ mutate の形 — 宛先と要求の連番を
 * 確かめてから成功した応答を反映し、**失敗は投げる**（呼び出し側 KbPage がトーストで知らせる）。
 * 成功したら一覧からその提案を取り除くだけで、一覧を丸ごと引き直さない。
 */
export function useKbPageSuggestions(
  workspaceSlug: string | undefined,
  pageId: string | undefined,
  open: boolean,
) {
  const [state, setState] = useState<KbPageSuggestionsState>(EMPTY);

  const active = useRef<SuggestionsTarget | null>(null);
  const seq = useRef(0);
  // 取得中に採用・却下が割り込んで成功したかの判定用（useKbComments の writeCount と同じ理由）。
  const writeCount = useRef(0);
  const target = targetOf(workspaceSlug, pageId, open);
  const targetKey = target?.key ?? null;

  const load = useCallback(async (to: SuggestionsTarget) => {
    const request = ++seq.current;
    const writesAtStart = writeCount.current;
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const suggestions = await KbRepository.listOpenSuggestions(to.workspaceSlug, to.pageId);
      if (active.current?.key !== to.key || seq.current !== request) return;
      if (writeCount.current !== writesAtStart) {
        // 取得中に採用・却下が成功していた。取得結果はそれより前のスナップショットで
        // 古いので上書きしない（解決済みの提案が一覧に生き返って見えるのを防ぐ）。
        setState((prev) => ({ ...prev, loading: false }));
        return;
      }
      setState({ suggestions, loading: false, error: null });
    } catch {
      if (active.current?.key !== to.key || seq.current !== request) return;
      if (writeCount.current !== writesAtStart) {
        setState((prev) => ({ ...prev, loading: false }));
        return;
      }
      setState({ ...EMPTY, error: LOAD_FAILED });
    }
  }, []);

  useEffect(() => {
    active.current = target;
    if (!target) {
      seq.current += 1;
      setState(EMPTY);
      return;
    }
    void load(target);
    // target は毎描画で作り直すオブジェクトなので、鍵で比べる。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [targetKey, load]);

  /**
   * mutate は採用・却下を 1 回行い、成功したら一覧からその提案を取り除く。
   * **失敗は投げる**（呼び出し側 KbPage がトーストで知らせる）。
   */
  const mutate = useCallback(
    async (
      suggestionId: string,
      run: (to: SuggestionsTarget) => Promise<KbPageSuggestion>,
    ): Promise<KbPageSuggestion> => {
      const to = active.current;
      if (!to) throw new Error('提案の一覧が確定していないため操作できません。');
      const request = seq.current;
      const result = await run(to);
      if (active.current?.key === to.key && seq.current === request) {
        writeCount.current += 1;
        setState((prev) => ({
          ...prev,
          suggestions: prev.suggestions.filter((s) => s.id !== suggestionId),
        }));
      }
      return result;
    },
    [],
  );

  const accept = useCallback(
    (suggestionId: string) =>
      mutate(suggestionId, (to) => KbRepository.acceptSuggestion(to.workspaceSlug, to.pageId, suggestionId)),
    [mutate],
  );

  const reject = useCallback(
    (suggestionId: string) =>
      mutate(suggestionId, (to) => KbRepository.rejectSuggestion(to.workspaceSlug, to.pageId, suggestionId)),
    [mutate],
  );

  return { ...state, accept, reject };
}
