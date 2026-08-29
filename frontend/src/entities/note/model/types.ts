/**
 * ノートの型。backend の `kb*Response`（`backend/internal/handler/kb_*_handler.go`）と 1:1。
 *
 * 3 つの入れ子で出来ている。
 *
 *   ワークスペース  会社の境界。同時に 2 つ見る場面が無いので UI では「切り替え」で表す
 *     └ スペース    部署や個人の区画。同時に見たいので UI では「見出し」で並べる
 *         └ ページ  木。親を持ち、兄弟の並び順は配列の順序で表される
 *
 * リッチ文書（entities/document）とは別系統であることに注意。あちらは所有者スコープの
 * 平らな一覧で、こちらは付与（grant）と例外（restriction）で解決する木。当面は並存する。
 */

/** ワークスペース 1 件。内部 UUID は外に出さず、URL も API も slug で指す。 */
export interface NoteWorkspace {
  slug: string;
  name: string;
  createdAt: string;
  /** 自分がこのワークスペースの admin か。削除操作を出してよいかの判定に使う。 */
  canManage: boolean;
}

/** スペース 1 件。key はワークスペース内で一意の短い識別子。 */
export interface NoteSpace {
  id: string;
  key: string;
  name: string;
  /** サイドバーの節分け。workspace = チーム（全員） / private = プライベート（付与された人だけ）。 */
  visibility: 'workspace' | 'private';
  createdAt: string;
}

/**
 * ページ 1 件（本文は含まない）。
 *
 * **並び順のキー（position）は入っていない。** backend が意図的に返していない。
 * 分数インデックスの整数部は末尾追加のたびに 1 ずつ増えるので、a0 と a3 が見えて
 * a1 a2 が見えなければ、その間に 2 枚あることがそのまま読めてしまうため。
 *
 * 並び順は**配列の順序そのもの**が持っている。ここで並べ替えないこと。
 */
export interface NotePage {
  id: string;
  spaceId: string;
  parentId?: string;
  title: string;
  createdByUserId: number;
  archivedAt?: string;
  createdAt: string;
  updatedAt: string;
}

/**
 * ツリーの 1 ノード。
 *
 * hasHiddenChildren は「この段の直下に、自分には見えないページが在るか」。
 * **枚数も題名も返ってこない。**
 *
 * 見えないページをただ消すと、木に穴が空いた理由が分からず「壊れている」と読まれるので、
 * 居ることだけを示す。枚数を出さないのは、利用者にとって「2 枚」と「7 枚」の差が行動を
 * 何も変えないのに、伏せた量に比例して漏れる情報が増えるため。
 *
 * なお**見えない親の配下は印にも出ない**（backend 側で見ていない。見ると
 * 「見えない枝の中にも何かある」ことまで漏れるため）。
 */
export interface NotePageTreeNode {
  page: NotePage;
  children: NotePageTreeNode[];
  hasHiddenChildren: boolean;
  /**
   * 親がアーカイブ済みか。**アーカイブ済みの一覧でだけ意味を持つ**（現役では常に false）。
   *
   * これは事実であって判断ではない。復帰できるかの規則は「親がアーカイブ中なら断る」で、
   * backend の usecase が持っている。ここで canRestore という名前にすると、
   * 同じ規則がフロントにも写り、必ずずれる。
   */
  parentArchived: boolean;
}

/**
 * ツリー取得の応答全体。
 *
 * hasHiddenChildren はスペース直下に見えないページが在るか。
 * **1 件も見えないスペースでは必ず false** になる（存在しないスペースと撃ち分けると、
 * 応答の差からスペース ID の実在を数え上げられてしまうため）。
 */
export interface NotePageTree {
  pages: NotePageTreeNode[];
  hasHiddenChildren: boolean;
}

/**
 * ページのメタ情報と本文（ProseMirror の doc JSON）の組。
 *
 * doc の中身は tiptap のスキーマそのもの。型は shared/ui/RichTextEditor の RichDocContent と
 * 同じものを指すが、entities から shared/ui の部品型に依存させたくないので unknown で受け、
 * 描画する画面側で検証する（isRichDoc）。
 */
export interface NotePageDoc {
  page: NotePage;
  doc: unknown;
}

/**
 * /p/{pageId} の解決結果。URL はページ ID しか持たないので、
 * 所属ワークスペースの slug と編集可否をサーバーが一緒に返す。
 */
export interface NoteResolvedPage {
  workspaceSlug: string;
  /** ワークスペースの表示名（パンくず用）。 */
  workspaceName: string;
  page: NotePage;
  doc: unknown;
  canEdit: boolean;
  /**
   * 閲覧できる祖先だけが根から順に入る（パンくず用）。
   * 見えない祖先は行ごと無い — 木と同じ規則で、穴があき得る。
   */
  ancestors: NoteAncestorRef[];
}

/** パンくず 1 段分（ページ ID と現在の題名）。 */
export interface NoteAncestorRef {
  id: string;
  title: string;
}
