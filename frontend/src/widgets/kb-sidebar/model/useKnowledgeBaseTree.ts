import { useCallback, useEffect, useRef, useState } from 'react';
import {
  KnowledgeBaseRepository,
  collectKbAncestorIds,
  replaceKbPageInTree,
  moveKbPageInTree,
  type KbDropTarget,
  type KbPage,
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

  const [spaceStates, setSpaceStatesRaw] = useState<Record<string, KbSpaceState>>({});
  // いまの spaceStates を同期して持つ控え。
  //
  // useState の更新関数は**呼んだその場では走らない**（次の描画で走る）。その中で
  // 結果を外の変数へ書き出して直後に読むと、まだ書かれていないことがある。
  // 実際、移動の可否をその形で判定していて、たまたま動いていただけだった。
  // 読みたいときは ref を読む。更新関数は state を作るだけに保つ。
  const spaceStatesRef = useRef<Record<string, KbSpaceState>>({});
  const setSpaceStates = useCallback(
    (update: (prev: Record<string, KbSpaceState>) => Record<string, KbSpaceState>) => {
      const next = update(spaceStatesRef.current);
      spaceStatesRef.current = next;
      setSpaceStatesRaw(next);
    },
    [],
  );
  /** setTree はそのスペースの木だけを差し替える。 */
  const setTree = useCallback(
    (spaceId: string, tree: KbPageTree | null) => {
      setSpaceStates((prev) => {
        const current = prev[spaceId];
        if (!current) return prev;
        return { ...prev, [spaceId]: { ...current, tree } };
      });
    },
    [setSpaceStates],
  );
  // 移動が走っているスペース。同じスペースの移動を重ねないための札。
  const movingSpaces = useRef<Set<string>>(new Set());
  // アーカイブ済みを見ているか。**ワークスペース全体で 1 つ**の切り替えにしてある。
  // スペースごとに持たせると「いまどちらを見ているのか」が場所によって変わり、
  // 木を置き換えるという体験が成立しない（設計の「必要なときだけ木を置き換える」）。
  const [archivedMode, setArchivedModeState] = useState(false);
  // 木を取りに行く関数が**常に「いまのスコープ」で取る**ようにするための控え。
  //
  // state を直接読むと、その関数を作った時点のスコープが閉じ込められる。書き換えの
  // 完了後に木を取り直す経路（アーカイブ・復帰）は await をまたぐので、その間に
  // 切り替えられると**古いスコープで取りに行き、その結果が新しい表示に入る**。
  // ref から読めば、誰がいつ呼んでも取りに行く先はいまのスコープになる。
  const archivedModeRef = useRef(false);
  const [expandedPageIds, setExpandedPageIds] = useState<ReadonlySet<string>>(new Set());

  // 切り替えを速く繰り返したときに、古い応答が新しい表示を上書きするのを防ぐ。
  // 「最後に投げた要求」だけを採用する（AbortController でも良いが、採用可否だけなら世代番号で足りる）。
  const generation = useRef(0);

  // 所属ワークスペースの一覧。
  //
  // 取りに行く処理を effect の中に直接書かず関数に切り出してあるのは、**失敗したときに
  // 同じ経路でやり直せるようにする**ため。effect の中に閉じ込めると、依存が変わらない限り
  // 二度と走らず、利用者は画面を再読み込みするしか手が無くなる。
  const loadWorkspaces = useCallback(() => {
    setWorkspacesLoading(true);
    setWorkspacesError(null);
    KnowledgeBaseRepository.fetchWorkspaces()
      .then((list) => {
        setWorkspaces(list);
        setWorkspacesError(null);
        // URL が何も指していなければ先頭を開く。所属が 0 件なら選ばない。
        setActiveSlug((current) => current ?? list[0]?.slug ?? null);
      })
      .catch(() => {
        setWorkspacesError('ワークスペースを読み込めませんでした');
      })
      .finally(() => {
        setWorkspacesLoading(false);
      });
  }, []);

  useEffect(() => {
    loadWorkspaces();
  }, [loadWorkspaces]);

  // URL 側が変わったら追従する（戻る / 進むでも表示が合う）。
  useEffect(() => {
    if (workspaceSlug) setActiveSlug(workspaceSlug);
  }, [workspaceSlug]);

  // 選んでいるワークスペースのスペース一覧。
  const loadSpaces = useCallback(() => {
    if (!activeSlug) return;
    const token = ++generation.current;
    setSpacesLoading(true);
    // **古い一覧を先に捨てる。** 残したまま新しい応答を待つと、その間だけ
    // 「前のワークスペースのスペース」と「新しい slug」が組み合わさって表示され、
    // そこを開くと別ワークスペースの spaceId で木を取りに行く。
    setSpaces([]);
    setSpacesError(null);
    setSpaceStates(() => ({}));
    setExpandedPageIds(new Set());

    KnowledgeBaseRepository.fetchSpaces(activeSlug)
      .then((list) => {
        if (token !== generation.current) return;
        setSpaces(list);
        setSpacesError(null);
        // 先頭のスペースだけ開いておく（何も開いていない画面は「壊れている」と読まれる）。
        if (list[0]) setSpaceStates(() => ({ [list[0].id]: emptySpaceState(true) }));
      })
      .catch(() => {
        if (token !== generation.current) return;
        setSpaces([]);
        setSpacesError('スペースを読み込めませんでした');
      })
      .finally(() => {
        if (token === generation.current) setSpacesLoading(false);
      });
  }, [activeSlug, setSpaceStates]);

  useEffect(() => {
    loadSpaces();
  }, [loadSpaces]);

  /** 1 スペース分の木を取りに行く。開いたとき・再試行のときだけ呼ぶ。 */
  const loadSpaceTree = useCallback(
    (spaceId: string) => {
      if (!activeSlug) return;
      const token = generation.current;
      setSpaceStates((prev) => ({
        ...prev,
        [spaceId]: { ...(prev[spaceId] ?? emptySpaceState(true)), loading: true, error: null },
      }));

      KnowledgeBaseRepository.fetchPageTree(activeSlug, spaceId, { archived: archivedModeRef.current })
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
    [activeSlug, setSpaceStates],
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
  }, [setSpaceStates]);

  /**
   * 現役とアーカイブ済みを切り替える。
   *
   * 取得済みの木は**捨てる**。同じスペースでも中身がまったく別なので、残しておくと
   * 切り替えた直後だけ前のスコープの木が見える。開いていたスペース（open）は保つ。
   */
  const setArchivedMode = useCallback((next: boolean) => {
    // 切り替え前に投げた要求を採用しない（古いスコープの木が後から届く）。
    generation.current += 1;
    // ref を先に更新する。この後の取り直しは必ず新しいスコープで走る。
    archivedModeRef.current = next;
    setArchivedModeState(next);
    setExpandedPageIds(new Set());
    setSpaceStates((prev) => {
      const cleared: Record<string, KbSpaceState> = {};
      for (const [spaceId, state] of Object.entries(prev)) {
        cleared[spaceId] = { open: state.open, loading: false, error: null, tree: null };
      }
      return cleared;
    });
  }, [setSpaceStates]);

  /**
   * ページを（子孫ごと）アーカイブする。**失敗は握り潰さず投げる。**
   *
   * 成功したらその段の木を取り直す。消えるのは 1 枚とは限らない（子孫ごと消える）ので、
   * 手元で 1 枚だけ抜くと表示と中身がずれる。
   */
  const archivePage = useCallback(
    async (spaceId: string, pageId: string): Promise<void> => {
      if (!activeSlug) throw new Error('workspace is not selected');
      await KnowledgeBaseRepository.archivePage(activeSlug, pageId);
      loadSpaceTree(spaceId);
    },
    [activeSlug, loadSpaceTree],
  );

  /**
   * アーカイブしたページを現役へ戻す。**失敗は握り潰さず投げる。**
   *
   * 戻るのも 1 枚とは限らないので、こちらも木を取り直す。
   */
  const unarchivePage = useCallback(
    async (spaceId: string, pageId: string): Promise<void> => {
      if (!activeSlug) throw new Error('workspace is not selected');
      await KnowledgeBaseRepository.unarchivePage(activeSlug, pageId);
      loadSpaceTree(spaceId);
    },
    [activeSlug, loadSpaceTree],
  );

  /**
   * ドラッグで動かす。**先に画面を動かし、断られたら元の並びへ戻す。**
   *
   * ドラッグは即座に動かないと使えないが、サーバーは拒否しうる（権限・競合・循環）。
   * だから先に動かす。ただし**戻せる形でしか動かさない**こと — 戻せないと、画面と
   * DB が食い違ったまま利用者が次の操作をする。
   *
   * 巻き戻しは木の取り直しではなく、**動かす前の木をそのまま書き戻す**。取り直すと
   * 失敗が見えないまま画面だけ整い、しかも取り直しの間に別の操作が挟まると
   * どちらが正か分からなくなる。
   *
   * 成功しても取り直さない。サーバーが受け入れた並びは、こちらが先に描いたものと同じ
   * （どの兄弟の隣かで指定しているので、解釈が割れる余地が無い）。
   *
   * **失敗は握り潰さず投げる。** 巻き戻しはここで済ませるが、知らせるのは呼び出し側。
   */
  const movePage = useCallback(
    async (spaceId: string, pageId: string, target: KbDropTarget): Promise<void> => {
      if (!activeSlug) throw new Error('workspace is not selected');
      // 同じスペースで移動が走っている間は次を受け付けない。
      //
      // 重ねると、1 本目が失敗したときに戻す先が 2 本目の結果の上になり、どちらが正か
      // 決められなくなる（2 本目は 1 本目の結果の上に積まれているため）。
      // 直列にすれば「巻き戻し先は必ず自分が動かす前」という約束が保てる。
      if (movingSpaces.current.has(spaceId)) throw new Error('move already in flight');

      const current = spaceStatesRef.current[spaceId];
      if (!current?.tree) throw new Error('invalid drop target');
      const pages = moveKbPageInTree(current.tree.pages, pageId, target);
      // 動かせない指定（自分自身・自分の子孫の中・落下先が無い）は、投げる前に断る。
      if (!pages) throw new Error('invalid drop target');

      // 動かす前の木を控える。これが唯一の巻き戻し先。
      const previous = current.tree;
      const optimistic = { ...current.tree, pages };
      movingSpaces.current.add(spaceId);
      setTree(spaceId, optimistic);
      // 子として入れたときは、その段を開いておく。開かないと、動かしたページが
      // 畳まれた段の中に入り、成功したのに画面から消えたように見える。
      if (target.kind === 'into') {
        setExpandedPageIds((prev) =>
          prev.has(target.pageId) ? prev : new Set([...prev, target.pageId]),
        );
      }

      // 落下先を API の言葉へ移す。**並び順のキーは送らない**（そもそも持っていない）。
      const request =
        target.kind === 'into'
          ? { parentId: target.pageId }
          : {
              parentId: parentIdOf(previous, target.pageId) ?? '',
              ...(target.kind === 'before'
                ? { beforePageId: target.pageId }
                : { afterPageId: target.pageId }),
            };

      try {
        await KnowledgeBaseRepository.movePage(activeSlug, pageId, request);
      } catch (error) {
        // 自分が描いた木がまだ表示されているときだけ戻す。別のものに変わっていたら
        // （スコープの切り替え・取り直し）、そちらのほうが新しいので触らない。
        if (spaceStatesRef.current[spaceId]?.tree === optimistic) {
          setTree(spaceId, previous);
        }
        throw error;
      } finally {
        movingSpaces.current.delete(spaceId);
      }
    },
    [activeSlug, setTree],
  );

  const togglePage = useCallback((pageId: string) => {
    setExpandedPageIds((prev) => {
      const next = new Set(prev);
      if (next.has(pageId)) next.delete(pageId);
      else next.add(pageId);
      return next;
    });
  }, []);

  /**
   * ページを作る。**失敗は握り潰さず投げる。**
   *
   * このリポジトリには「操作は失敗したのに成功の表示が出る」轍が既にあり
   * （コース削除・教材の保存）、原因はどれも**操作関数が失敗を投げなかった**こと。
   * 返り値の真偽で伝えると、呼び出し側は見なくても書けてしまう。投げれば、
   * 握り潰すには try/catch を書くしかなく、握り潰したことがコードに残る。
   *
   * 成功したらその段の木を取り直す。**兄弟のどこに入るかを決めるのはサーバー**なので、
   * 手元で組み立てると必ずずれる（並び順のキーは応答にも入っていない）。
   */
  const createPage = useCallback(
    async (spaceId: string, parentId?: string): Promise<KbPage> => {
      if (!activeSlug) throw new Error('workspace is not selected');
      const page = await KnowledgeBaseRepository.createPage(activeSlug, spaceId, {
        title: KB_NEW_PAGE_TITLE,
        parentId,
      });
      // 親の下に作ったなら、その親を開いておく（開かないと作ったページが見えない）。
      if (parentId) {
        setExpandedPageIds((prev) => (prev.has(parentId) ? prev : new Set([...prev, parentId])));
      }
      loadSpaceTree(spaceId);
      return page;
    },
    [activeSlug, loadSpaceTree],
  );

  /**
   * 題名を変える。**失敗は握り潰さず投げる**（createPage と同じ理由）。
   *
   * 成功したら木ごと取り直さず、サーバーが返したページで 1 枚だけ差し替える。
   * 取り直すと一瞬空になり、開いていた段も畳まれて見えるため。
   */
  const renamePage = useCallback(
    async (spaceId: string, pageId: string, title: string): Promise<KbPage> => {
      if (!activeSlug) throw new Error('workspace is not selected');
      const page = await KnowledgeBaseRepository.renamePage(activeSlug, pageId, title);
      setSpaceStates((prev) => {
        const current = prev[spaceId];
        if (!current?.tree) return prev;
        const pages = replaceKbPageInTree(current.tree.pages, page);
        if (pages === current.tree.pages) return prev;
        return { ...prev, [spaceId]: { ...current, tree: { ...current.tree, pages } } };
      });
      return page;
    },
    [activeSlug, setSpaceStates],
  );

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
    retryWorkspaces: loadWorkspaces,
    activeSlug,
    selectWorkspace: setActiveSlug,
    spaces,
    spacesLoading,
    spacesError,
    retrySpaces: loadSpaces,
    spaceStates,
    toggleSpace,
    retrySpace,
    expandedPageIds,
    togglePage,
    createPage,
    renamePage,
    archivePage,
    unarchivePage,
    movePage,
    archivedMode,
    setArchivedMode,
  };
}

/** parentIdOf は木の中でそのページの親の ID を返す（スペース直下なら null）。 */
function parentIdOf(tree: KbPageTree | null, pageId: string): string | null {
  if (!tree) return null;
  const walk = (nodes: KbPageTree['pages'], parentId: string | null): string | null | undefined => {
    for (const node of nodes) {
      if (node.page.id === pageId) return parentId;
      const found = walk(node.children, node.page.id);
      if (found !== undefined) return found;
    }
    return undefined;
  };
  return walk(tree.pages, null) ?? null;
}

/** 新しく作ったページの題名。作った直後にその場で書き換えられる。 */
export const KB_NEW_PAGE_TITLE = '無題';

function emptySpaceState(open: boolean): KbSpaceState {
  return { open, loading: false, error: null, tree: null };
}
