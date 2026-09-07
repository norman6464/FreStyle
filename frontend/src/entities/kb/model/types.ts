/**
 * ナレッジの型。backend の `kb*Response`（`backend/internal/handler/kb_*_handler.go`）と 1:1。
 *
 * 3 つの入れ子で出来ている。
 *
 *   ワークスペース  会社の境界。同時に 2 つ見る場面が無いので UI では「切り替え」で表す
 *     └ スペース    部署や個人の区画。同時に見たいので UI では「見出し」で並べる
 *         └ ページ  木。親を持ち、兄弟の並び順は配列の順序で表される
 */

/** ワークスペース 1 件。内部 UUID は外に出さず、URL も API も slug で指す。 */
export interface KbWorkspace {
  slug: string;
  name: string;
  createdAt: string;
  /** 自分がこのワークスペースの admin か。削除操作を出してよいかの判定に使う。 */
  canManage: boolean;
}

/** スペース 1 件。key はワークスペース内で一意の短い識別子。 */
export interface KbSpace {
  id: string;
  key: string;
  name: string;
  /** サイドバーの節分け。workspace = チーム（全員） / private = プライベート（付与された人だけ）。 */
  visibility: 'workspace' | 'private';
  createdAt: string;
}

/**
 * ページのアイコン。いまは絵文字だけ（`type` を持たせておくのは、いつか他の種類
 * （アップロード画像など）が増えたときに判別できるようにするため）。
 */
export interface KbIcon {
  type: 'emoji';
  value: string;
}

/**
 * 「最終編集者」の参照 1 件。
 *
 * name は表示名で、**引けなければ空文字**（backend が行を落とさずそう返す。
 * KbGrantablePrincipal と同じ約束）。
 */
export interface KbEditorRef {
  userId: number;
  name: string;
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
export interface KbPage {
  id: string;
  spaceId: string;
  parentId?: string;
  title: string;
  createdByUserId: number;
  archivedAt?: string;
  createdAt: string;
  updatedAt: string;
  /** 未設定は null（明示的に外した）と undefined（旧応答）の両方であり得る。 */
  icon?: KbIcon | null;
  lastEditedByUserId?: number;
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
export interface KbPageTreeNode {
  page: KbPage;
  children: KbPageTreeNode[];
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
export interface KbPageTree {
  pages: KbPageTreeNode[];
  hasHiddenChildren: boolean;
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

/**
 * /kb/{pageId} の解決結果。URL はページ ID しか持たないので、
 * 所属ワークスペースの slug と編集可否をサーバーが一緒に返す。
 */
export interface KbResolvedPage {
  workspaceSlug: string;
  /** ワークスペースの表示名（パンくず用）。 */
  workspaceName: string;
  page: KbPage;
  doc: unknown;
  canEdit: boolean;
  /**
   * このページの権限を変えられるか（共有ボタンを出すかの判定に使う）。
   *
   * ナレッジは付与（grant）だけで解決する木で、打ち消す層を持たない。したがって
   * canEdit を弱める例外も無く、canManage は上位から届く権限をそのまま見る値になる。
   */
  canManage: boolean;
  /**
   * 閲覧できる祖先だけが根から順に入る（パンくず用）。
   * 見えない祖先は行ごと無い — 木と同じ規則で、穴があき得る。
   */
  ancestors: KbAncestorRef[];
  /** 本文を最後に保存した人。旧応答（デプロイ順）や不明なユーザーでは無い/空文字。 */
  lastEditedBy?: KbEditorRef | null;
  /** 最終編集の日時（= page_snapshots.built_at）。lastEditedBy と対になる。 */
  lastEditedAt?: string | null;
  /**
   * カバー画像。未設定は null（明示的に外した）と undefined（旧応答）の両方があり得る
   * （KbIcon と同じ約束）。**一覧・木の KbPage には出てこない**（N+1 回避のため、
   * backend は一覧応答では解決しない — このページ単体の解決応答でだけ入る）。
   */
  cover?: KbResolvedCover | null;
}

/** パンくず 1 段分（ページ ID と現在の題名）。 */
export interface KbAncestorRef {
  id: string;
  title: string;
}

/**
 * 解決済みのカバー画像。durable な保存形式は S3 の key（"kb/…"）だが、ここに来るのは
 * サーバーが既に署名して解決した後の一時 URL — そのまま `<img src>` に使ってよい。
 */
export interface KbResolvedCover {
  type: 'file';
  url: string;
}

/** 既定の役割。強い順に admin > editor > commenter > viewer。 */
export type KbGrantRole = 'admin' | 'editor' | 'commenter' | 'viewer';

/**
 * ページ自身に張られた既定の役割 1 件。
 *
 * **「このページを見られる人」ではない。** 返るのはこの段で足した行だけで、
 * ワークスペース / スペース / 祖先のページから届いている相手は含まれない。
 * 空でも「誰も見られない」ではなく「この段では何も足していない」の意味になる。
 */
export interface KbPageGrant {
  pageId: string;
  principalId: string;
  role: KbGrantRole;
  createdAt: string;
  updatedAt: string;
}

/**
 * 権限を張れる相手 1 件。
 *
 * name は表示名で、**引けなかった場合は空文字**（backend が行を落とさずそう返す）。
 * 画面もそれに合わせて行を消さない — 消すと、その相手に張った権限が一覧に出たまま
 * 選べなくなる。
 */
export interface KbGrantablePrincipal {
  id: string;
  kind: 'user' | 'group' | 'space_all';
  name: string;
}
