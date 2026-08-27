import { useCallback, useEffect, useRef, useState } from 'react';
import {
  KnowledgeBaseRepository,
  collectKbAncestorIds,
  type KbPageTree,
  type KbSpace,
  type KbWorkspace,
} from '@/entities/knowledge-base';

/** 1 スペース分の読み込み状態。開くまでは何も取りに行かない。 */
export interface KbSpaceState {
  open: boolean;
  loading: boolean;
  /** 取得に失敗した理由。表示して再試行させる（黙って空にしない）。 */
  error: string | null;
  tree: KbPageTree | null;
}

export interface UseKnowledgeBaseTreeOptions {
  /** URL が指しているワークスペース。未指定なら所属の先頭を選ぶ。 */
  workspaceSlug?: string;
  /** URL が指しているページ。祖先を自動で開くために使う。 */
  activePageId?: string;
}

/**
 * useKnowledgeBaseTree はサイドバーが要るものを全部揃える。
 *
 * 取りに行く順は ワークスペース → スペース → （開いたスペースだけ）ページの木。
 * **スペースの木は開くまで取りに行かない。** 全スペースの木を最初に取ると、
 * スペースの数だけ要求が飛ぶ（いわゆる N+1）。木は 1 回で全件返る作りなので、
 * 開いていないスペースの分は丸ごと無駄になる。
 *
 * 失敗は必ず状態として持ち、画面に出す。既にこのリポジトリには
 * 「操作は失敗したのに成功の表示が出る」轍があるので、同じ形を作らない。
 */
export function useKnowledgeBaseTree(options: UseKnowledgeBaseTreeOptions = {}) {
  const { workspaceSlug, activePageId } = options;

  const [workspaces, setWorkspaces] = useState<KbWorkspace[]>([]);
  const [workspacesLoading, setWorkspacesLoading] = useState(true);
  const [workspacesError, setWorkspacesError] = useState<string | null>(null);

  const [activeSlug, setActiveSlug] = useState<string | null>(workspaceSlug ?? null);

  const [spaces, setSpaces] = useState<KbSpace[]>([]);
  const [spacesLoading, setSpacesLoading] = useState(false);
  const [spacesError, setSpacesError] = useState<string | null>(null);

  const [spaceStates, setSpaceStates] = useState<Record<string, KbSpaceState>>({});
  const [expandedPageIds, setExpandedPageIds] = useState<ReadonlySet<string>>(new Set());

  // 切り替えを速く繰り返したときに、古い応答が新しい表示を上書きするのを防ぐ。
  // 「最後に投げた要求」だけを採用する（AbortController でも良いが、採用可否だけなら世代番号で足りる）。
  const generation = useRef(0);

  // 所属ワークスペースの一覧（1 回だけ）。
  useEffect(() => {
    let alive = true;
    setWorkspacesLoading(true);
    KnowledgeBaseRepository.fetchWorkspaces()
      .then((list) => {
        if (!alive) return;
        setWorkspaces(list);
        setWorkspacesError(null);
        // URL が何も指していなければ先頭を開く。所属が 0 件なら選ばない。
        setActiveSlug((current) => current ?? list[0]?.slug ?? null);
      })
      .catch(() => {
        if (!alive) return;
        setWorkspacesError('ワークスペースを読み込めませんでした');
      })
      .finally(() => {
        if (alive) setWorkspacesLoading(false);
      });
    return () => {
      alive = false;
    };
  }, []);

  // URL 側が変わったら追従する（戻る / 進むでも表示が合う）。
  useEffect(() => {
    if (workspaceSlug) setActiveSlug(workspaceSlug);
  }, [workspaceSlug]);

  // 選んでいるワークスペースのスペース一覧。切り替えたら木の状態も捨てる。
  useEffect(() => {
    if (!activeSlug) return;
    const token = ++generation.current;
    setSpacesLoading(true);
    setSpaceStates({});
    setExpandedPageIds(new Set());

    KnowledgeBaseRepository.fetchSpaces(activeSlug)
      .then((list) => {
        if (token !== generation.current) return;
        setSpaces(list);
        setSpacesError(null);
        // 先頭のスペースだけ開いておく（何も開いていない画面は「壊れている」と読まれる）。
        if (list[0]) setSpaceStates({ [list[0].id]: emptySpaceState(true) });
      })
      .catch(() => {
        if (token !== generation.current) return;
        setSpaces([]);
        setSpacesError('スペースを読み込めませんでした');
      })
      .finally(() => {
        if (token === generation.current) setSpacesLoading(false);
      });
  }, [activeSlug]);

  /** 1 スペース分の木を取りに行く。開いたとき・再試行のときだけ呼ぶ。 */
  const loadSpaceTree = useCallback(
    (spaceId: string) => {
      if (!activeSlug) return;
      const token = generation.current;
      setSpaceStates((prev) => ({
        ...prev,
        [spaceId]: { ...(prev[spaceId] ?? emptySpaceState(true)), loading: true, error: null },
      }));

      KnowledgeBaseRepository.fetchPageTree(activeSlug, spaceId)
        .then((tree) => {
          if (token !== generation.current) return;
          setSpaceStates((prev) => ({
            ...prev,
            [spaceId]: { open: true, loading: false, error: null, tree },
          }));
        })
        .catch(() => {
          if (token !== generation.current) return;
          setSpaceStates((prev) => ({
            ...prev,
            [spaceId]: {
              open: true,
              loading: false,
              error: 'ページを読み込めませんでした',
              tree: prev[spaceId]?.tree ?? null,
            },
          }));
        });
    },
    [activeSlug],
  );

  // 開いていて、まだ取っていないスペースを取りに行く。
  // 「開く」操作と「取りに行く」処理を分けてあるので、初期表示でも再試行でも同じ経路を通る。
  useEffect(() => {
    for (const [spaceId, state] of Object.entries(spaceStates)) {
      if (state.open && !state.loading && state.tree === null && state.error === null) {
        loadSpaceTree(spaceId);
      }
    }
  }, [spaceStates, loadSpaceTree]);

  // 現在位置のページの祖先を開く。
  //
  // 木の読み込み完了とは別の effect にしてある。同じ場所でやると、木が届いた瞬間しか
  // 反応せず、**既に読み込んだ木の中で別のページへ移動したとき**に祖先が開かない
  // （リンクを辿ると、開いたページが閉じた枝の中に隠れたままになる）。
  useEffect(() => {
    if (!activePageId) return;
    for (const state of Object.values(spaceStates)) {
      if (!state.tree) continue;
      const ancestors = collectKbAncestorIds(state.tree.pages, activePageId);
      if (ancestors.length === 0) continue;
      setExpandedPageIds((prev) => {
        // 既に全部開いていれば新しい集合を作らない（作ると再描画が無限に続く）。
        if (ancestors.every((id) => prev.has(id))) return prev;
        return new Set([...prev, ...ancestors]);
      });
    }
  }, [activePageId, spaceStates]);

  const toggleSpace = useCallback((spaceId: string) => {
    setSpaceStates((prev) => {
      const current = prev[spaceId];
      // 閉じても取得済みの木は捨てない（開き直すたびに取りに行くと待たせるだけ）。
      if (current) return { ...prev, [spaceId]: { ...current, open: !current.open } };
      return { ...prev, [spaceId]: emptySpaceState(true) };
    });
  }, []);

  const togglePage = useCallback((pageId: string) => {
    setExpandedPageIds((prev) => {
      const next = new Set(prev);
      if (next.has(pageId)) next.delete(pageId);
      else next.add(pageId);
      return next;
    });
  }, []);

  const retrySpace = useCallback(
    (spaceId: string) => {
      loadSpaceTree(spaceId);
    },
    [loadSpaceTree],
  );

  return {
    workspaces,
    workspacesLoading,
    workspacesError,
    activeSlug,
    selectWorkspace: setActiveSlug,
    spaces,
    spacesLoading,
    spacesError,
    spaceStates,
    toggleSpace,
    retrySpace,
    expandedPageIds,
    togglePage,
  };
}

function emptySpaceState(open: boolean): KbSpaceState {
  return { open, loading: false, error: null, tree: null };
}
