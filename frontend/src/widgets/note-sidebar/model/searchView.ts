import type { NotePage, NoteSpace } from '@/entities/note';

/** 表示順（スペースの並び → 見えるスペースに無いもの）に平坦化した検索結果 1 面分。 */
export interface SearchView {
  groups: { space: NoteSpace; pages: NotePage[] }[];
  orphan: NotePage[];
  flat: NotePage[];
}

/**
 * buildSearchView は結果をスペースごとに束ね、キーボード選択用の平坦な並びも作る。
 *
 * 結果は木の形では出さない — 一致した行の祖先は応答に含まれておらず、フロントで
 * 木を組み立てるとサーバーの伏せた祖先を推測で埋めることになる。場所の手掛かりは
 * スペースの見出しまで。見えるスペース一覧に無いスペースのページ（個別に許可された
 * ページ等）は名前が引けないので、末尾に見出しなしで並べる（落とすと「検索では
 * 返ったのに画面に出ない」という消え方をする）。
 */
export function buildSearchView(pages: NotePage[], spaces: NoteSpace[]): SearchView {
  const bySpace = new Map<string, NotePage[]>();
  for (const page of pages) {
    const list = bySpace.get(page.spaceId) ?? [];
    list.push(page);
    bySpace.set(page.spaceId, list);
  }
  const groups = spaces
    .filter((space) => bySpace.has(space.id))
    .map((space) => ({ space, pages: bySpace.get(space.id) ?? [] }));
  const known = new Set(spaces.map((space) => space.id));
  const orphan = pages.filter((page) => !known.has(page.spaceId));
  return { groups, orphan, flat: [...groups.flatMap((g) => g.pages), ...orphan] };
}
