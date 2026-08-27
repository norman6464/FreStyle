/**
 * ナレッジ基盤の型。backend の `kb*Response`（`backend/internal/handler/kb_*_handler.go`）と 1:1。
 *
 * 3 つの入れ子で出来ている。
 *
 *   ワークスペース  会社の境界。同時に 2 つ見る場面が無いので UI では「切り替え」で表す
 *     └ スペース    部署や個人の区画。同時に見たいので UI では「見出し」で並べる
 *         └ ページ  木。親を持ち、兄弟の中での位置を position で持つ
 *
 * リッチ文書（entities/document）とは別系統であることに注意。あちらは所有者スコープの
 * 平らな一覧で、こちらは付与（grant）と例外（restriction）で解決する木。当面は並存する。
 */

/** ワークスペース 1 件。内部 UUID は外に出さず、URL も API も slug で指す。 */
export interface KbWorkspace {
  slug: string;
  name: string;
  createdAt: string;
}

/** スペース 1 件。key はワークスペース内で一意の短い識別子。 */
export interface KbSpace {
  id: string;
  key: string;
  name: string;
  createdAt: string;
}

/**
 * ページ 1 件（本文は含まない）。
 *
 * position は**分数インデックス**。兄弟の並び順を「a0」「a1」のような文字列で持ち、
 * 辞書順で並べる。間に挿すときは隣り合う 2 つの中間値を計算するだけで済み、
 * 他の行を書き換えない（整数の連番だと以降を全部ずらす必要があり、木では現実的でない）。
 */
export interface KbPage {
  id: string;
  spaceId: string;
  parentId?: string;
  position: string;
  title: string;
  createdByUserId: number;
  archivedAt?: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * ツリーの 1 ノード。
 *
 * hiddenChildCount は「この段の直下にある、自分には見えないページの数」。
 * 題名は返ってこない（件数だけ）。0 のときは何も示さない。
 *
 * 見えないページをただ消すと、木に穴が空いた理由が分からず「壊れている」と読まれるので、
 * 居ることだけを示す。なお**見えない親の配下は件数にも入らない**（backend 側で数えていない。
 * 数えると「見えない枝の中に何枚あるか」まで漏れるため）。
 */
export interface KbPageTreeNode {
  page: KbPage;
  children: KbPageTreeNode[];
  hiddenChildCount: number;
}

/**
 * ツリー取得の応答全体。
 *
 * hiddenChildCount はスペース直下で伏せた件数。**1 件も見えないスペースでは必ず 0** になる
 * （存在しないスペースと撃ち分けると、スペース ID の総当たりで実在が分かってしまうため）。
 */
export interface KbPageTree {
  pages: KbPageTreeNode[];
  hiddenChildCount: number;
}

/**
 * ページのメタ情報と本文（ProseMirror の doc JSON）の組。
 *
 * doc の中身は tiptap のスキーマそのもの。型は shared/ui/RichTextEditor の RichDocContent と
 * 同じものを指すが、entities から shared/ui の部品型に依存させたくないので unknown で受け、
 * 描画する画面側で検証する（isRichDoc）。
 */
export interface KbPageDoc {
  page: KbPage;
  doc: unknown;
}
