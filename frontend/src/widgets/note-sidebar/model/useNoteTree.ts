import { useCallback, useEffect, useRef, useState } from 'react';
import {
  NOTE_NEW_PAGE_TITLE,
  NoteRepository,
  emitNoteTreeEvent,
  collectNoteAncestorIds,
  subscribeNoteTreeEvents,
  replaceNotePageInTree,
  moveNotePageInTree,
  type NoteDropTarget,
  type NotePage,
  type NotePageTree,
  type NoteSpace,
  type NoteWorkspace,
} from '@/entities/note';

/** 1 スペース分の読み込み状態。開くまでは何も取りに行かない。 */
export interface NoteSpaceState {
  open: boolean;
  loading: boolean;
  /** 取得に失敗した理由。表示して再試行させる（黙って空にしない）。 */
  error: string | null;
  tree: NotePageTree | null;
}

export interface UseKnowledgeBaseTreeOptions {
  /** URL が指しているワークスペース。未指定なら所属の先頭を選ぶ。 */
  workspaceSlug?: string;
  /** URL が指しているページ。祖先を自動で開くために使う。 */
  activePageId?: string;
}

/**
 * useNoteTree はサイドバーが要るものを全部揃える。
 *
 * 取りに行く順は ワークスペース → スペース → （開いたスペースだけ）ページの木。
 * **スペースの木は開くまで取りに行かない。** 全スペースの木を最初に取ると、
 * スペースの数だけ要求が飛ぶ（いわゆる N+1）。木は 1 回で全件返る作りなので、
 * 開いていないスペースの分は丸ごと無駄になる。
 *
 * 失敗は必ず状態として持ち、画面に出す。既にこのリポジトリには
 * 「操作は失敗したのに成功の表示が出る」轍があるので、同じ形を作らない。
 */
export function useNoteTree(options: UseKnowledgeBaseTreeOptions = {}) {
  const { workspaceSlug, activePageId } = options;

  const [workspaces, setWorkspaces] = useState<NoteWorkspace[]>([]);
  const [workspacesLoading, setWorkspacesLoading] = useState(true);
  const [workspacesError, setWorkspacesError] = useState<string | null>(null);

  const [activeSlug, setActiveSlug] = useState<string | null>(workspaceSlug ?? null);

  const [spaces, setSpaces] = useState<NoteSpace[]>([]);
  const [spacesLoading, setSpacesLoading] = useState(false);
  const [spacesError, setSpacesError] = useState<string | null>(null);

  const [spaceStates, setSpaceStatesRaw] = useState<Record<string, NoteSpaceState>>({});
  // いまの spaceStates を同期して持つ控え。
  //
  // useState の更新関数は**呼んだその場では走らない**（次の描画で走る）。その中で
  // 結果を外の変数へ書き出して直後に読むと、まだ書かれていないことがある。
  // 実際、移動の可否をその形で判定していて、たまたま動いていただけだった。
  // 読みたいときは ref を読む。更新関数は state を作るだけに保つ。
  const spaceStatesRef = useRef<Record<string, NoteSpaceState>>({});
  const setSpaceStates = useCallback(
    (update: (prev: Record<string, NoteSpaceState>) => Record<string, NoteSpaceState>) => {
      const next = update(spaceStatesRef.current);
      spaceStatesRef.current = next;
      setSpaceStatesRaw(next);
    },
    [],
  );
  /** setTree はそのスペースの木だけを差し替える。 */
  const setTree = useCallback(
    (spaceId: string, tree: NotePageTree | null) => {
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
    NoteRepository.fetchWorkspaces()
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

    NoteRepository.fetchSpaces(activeSlug)
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

  // スペース単位の要求連番。ワークスペースの世代（generation）だけだと、同じスペースへの
  // 要求どうしの追い越しを防げない — 削除後の取り直しより先に投げた古い応答が後から届くと、
  // 消したはずのページが木に蘇る（選ぶと 404）。最後に投げた要求だけを採用する。
  const spaceTreeSeq = useRef<Record<string, number>>({});

  /** 1 スペース分の木を取りに行く。開いたとき・再試行のときだけ呼ぶ。 */
  const loadSpaceTree = useCallback(
    (spaceId: string) => {
      if (!activeSlug) return;
      const token = generation.current;
      const seq = (spaceTreeSeq.current[spaceId] ?? 0) + 1;
      spaceTreeSeq.current[spaceId] = seq;
      setSpaceStates((prev) => ({
        ...prev,
        [spaceId]: { ...(prev[spaceId] ?? emptySpaceState(true)), loading: true, error: null },
      }));

      NoteRepository.fetchPageTree(activeSlug, spaceId, { archived: archivedModeRef.current })
        .then((tree) => {
          if (token !== generation.current || seq !== spaceTreeSeq.current[spaceId]) return;
          setSpaceStates((prev) => ({
            ...prev,
            [spaceId]: { open: true, loading: false, error: null, tree },
          }));
        })
        .catch(() => {
          if (token !== generation.current || seq !== spaceTreeSeq.current[spaceId]) return;
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
      const ancestors = collectNoteAncestorIds(state.tree.pages, activePageId);
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
  /**
   * ワークスペースを作る。**失敗は握り潰さず投げる。**
   *
   * 作った本人が admin になるので、続けてスペースを作れる。作ったら一覧を取り直し、
   * そのワークスペースへ切り替える（作ってから自分で選び直させない）。
   */
  const createWorkspace = useCallback(
    async (input: { name: string }): Promise<NoteWorkspace> => {
      // URL に出る slug はサーバーが自動採番する（人に決めさせない）。
      const workspace = await NoteRepository.createWorkspace({ name: input.name });
      setWorkspaces((prev) => [...prev, workspace]);
      setActiveSlug(workspace.slug);
      return workspace;
    },
    [],
  );

  /**
   * スペースを作る。**失敗は握り潰さず投げる。**
   *
   * 作ったら一覧に足して開いておく。取り直さないのは、いま作ったものが必ず含まれると
   * 分かっているため（サーバーが返した行をそのまま足す）。
   */
  const createSpace = useCallback(
    async (input: { name: string; visibility?: 'workspace' | 'private' }): Promise<NoteSpace> => {
      if (!activeSlug) throw new Error('workspace is not selected');
      const space = await NoteRepository.createSpace(activeSlug, input);
      setSpaces((prev) => [...prev, space]);
      setSpaceStates((prev) => ({ ...prev, [space.id]: emptySpaceState(true) }));
      return space;
    },
    [activeSlug, setSpaceStates],
  );

  const renameSpace = useCallback(
    async (spaceId: string, name: string): Promise<NoteSpace> => {
      if (!activeSlug) throw new Error('workspace is not selected');
      const space = await NoteRepository.renameSpace(activeSlug, spaceId, name);
      // 見出しは spaces の配列から描くので、そこだけ差し替える（木は名前を持たない）。
      setSpaces((prev) => prev.map((s) => (s.id === space.id ? space : s)));
      return space;
    },
    [activeSlug],
  );

  const setArchivedMode = useCallback((next: boolean) => {
    // 切り替え前に投げた要求を採用しない（古いスコープの木が後から届く）。
    generation.current += 1;
    // ref を先に更新する。この後の取り直しは必ず新しいスコープで走る。
    archivedModeRef.current = next;
    setArchivedModeState(next);
    setExpandedPageIds(new Set());
    setSpaceStates((prev) => {
      const cleared: Record<string, NoteSpaceState> = {};
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
      await NoteRepository.archivePage(activeSlug, pageId);
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
      await NoteRepository.unarchivePage(activeSlug, pageId);
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
    async (spaceId: string, pageId: string, target: NoteDropTarget): Promise<void> => {
      if (!activeSlug) throw new Error('workspace is not selected');
      // 同じスペースで移動が走っている間は次を受け付けない。
      //
      // 重ねると、1 本目が失敗したときに戻す先が 2 本目の結果の上になり、どちらが正か
      // 決められなくなる（2 本目は 1 本目の結果の上に積まれているため）。
      // 直列にすれば「巻き戻し先は必ず自分が動かす前」という約束が保てる。
      if (movingSpaces.current.has(spaceId)) throw new Error('move already in flight');

      const current = spaceStatesRef.current[spaceId];
      if (!current?.tree) throw new Error('invalid drop target');
      const pages = moveNotePageInTree(current.tree.pages, pageId, target);
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
        await NoteRepository.movePage(activeSlug, pageId, request);
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

  // ページ画面（/p）での作成・改名を木に映す。作成は木ごと取り直し（親子関係の
  // 差し込み位置をこちらで計算しない — サーバーの並び順が正）、改名は 1 枚差し替え。
  useEffect(() => {
    return subscribeNoteTreeEvents((event) => {
      if (event.type === 'page-created') {
        const parentId = event.page.parentId;
        if (parentId) {
          setExpandedPageIds((prev) => (prev.has(parentId) ? prev : new Set([...prev, parentId])));
        }
        loadSpaceTree(event.page.spaceId);
        return;
      }
      if (event.type !== 'page-renamed') return;
      setSpaceStates((prev) => {
        const current = prev[event.page.spaceId];
        if (!current?.tree) return prev;
        const pages = replaceNotePageInTree(current.tree.pages, event.page);
        if (pages === current.tree.pages) return prev;
        return { ...prev, [event.page.spaceId]: { ...current, tree: { ...current.tree, pages } } };
      });
    });
  }, [loadSpaceTree, setSpaceStates]);

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
    async (spaceId: string, parentId?: string): Promise<NotePage> => {
      if (!activeSlug) throw new Error('workspace is not selected');
      const page = await NoteRepository.createPage(activeSlug, spaceId, {
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
    async (spaceId: string, pageId: string, title: string): Promise<NotePage> => {
      if (!activeSlug) throw new Error('workspace is not selected');
      const page = await NoteRepository.renamePage(activeSlug, pageId, title);
      setSpaceStates((prev) => {
        const current = prev[spaceId];
        if (!current?.tree) return prev;
        const pages = replaceNotePageInTree(current.tree.pages, page);
        if (pages === current.tree.pages) return prev;
        return { ...prev, [spaceId]: { ...current, tree: { ...current.tree, pages } } };
      });
      return page;
    },
    [activeSlug, setSpaceStates],
  );

  /**
   * ページを子孫ごと物理削除する。**失敗は握り潰さず投げる。**
   * 成功したらその段の木を取り直す（部分木がまとめて消えるので 1 枚差し替えでは表せない）。
   */
  const deletePage = useCallback(
    async (spaceId: string, pageId: string): Promise<void> => {
      if (!activeSlug) throw new Error('workspace is not selected');
      await NoteRepository.deletePage(activeSlug, pageId);
      // 開いている画面が「消えた場所」かの判定はページ側が行う（ページは自分の祖先を
      // サーバー応答で知っている。サイドバーの現役の木では、アーカイブ済みの子孫を
      // 開いている場合を見落とす）。
      emitNoteTreeEvent({ type: 'page-deleted', pageId });
      loadSpaceTree(spaceId);
    },
    [activeSlug, loadSpaceTree],
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
    deletePage,
    archivePage,
    unarchivePage,
    movePage,
    createWorkspace,
    createSpace,
    renameSpace,
    selectWorkspace: setActiveSlug,
    archivedMode,
    setArchivedMode,
  };
}

/** parentIdOf は木の中でそのページの親の ID を返す（スペース直下なら null）。 */
function parentIdOf(tree: NotePageTree | null, pageId: string): string | null {
  if (!tree) return null;
  const walk = (nodes: NotePageTree['pages'], parentId: string | null): string | null | undefined => {
    for (const node of nodes) {
      if (node.page.id === pageId) return parentId;
      const found = walk(node.children, node.page.id);
      if (found !== undefined) return found;
    }
    return undefined;
  };
  return walk(tree.pages, null) ?? null;
}

/** 新しく作ったページの題名。正本は entities/note（エディタの /page と共用）。 */
export const KB_NEW_PAGE_TITLE = NOTE_NEW_PAGE_TITLE;

function emptySpaceState(open: boolean): NoteSpaceState {
  return { open, loading: false, error: null, tree: null };
}
