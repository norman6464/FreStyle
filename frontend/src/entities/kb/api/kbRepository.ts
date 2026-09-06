import apiClient from '@/shared/api/axios';
import { toArray } from '@/shared/lib/toArray';
import { KB_API } from '@/shared/config/apiRoutes';
import type {
  KbGrantablePrincipal,
  KbGrantRole,
  KbPage,
  KbPageDoc,
  KbPageGrant,
  KbPageTree,
  KbResolvedPage,
  KbSpace,
  KbWorkspace,
} from '../model/types';

/**
 * ナレッジの API（/api/v2/kb/…）の薄いラッパ。
 *
 * 認可はすべて backend が持つ。ここでフィルタを掛けないこと。返ってくる一覧は
 * **既に「その人に見えるものだけ」**になっていて、見えないものは応答に存在しない。
 * フロントで絞り込みを重ねると、同じ判断が 2 箇所に分かれて必ずずれる。
 *
 * 「無い」と「見えない」はどちらも 404 で返る（撃ち分けると ID の総当たりで実在が分かるため）。
 * したがって 404 を「あなたには見えません」と表示してはいけない。存在しないかもしれない。
 */
const KbRepository = {
  /** 自分が所属しているワークスペースの一覧。所属が無ければ空配列。 */
  async fetchWorkspaces(): Promise<KbWorkspace[]> {
    const res = await apiClient.get<KbWorkspace[]>(KB_API.workspaces);
    return toArray<KbWorkspace>(res.data);
  },

  /**
   * ワークスペース配下の、自分に見えるスペースの一覧。
   *
   * ワークスペースのメンバーなら誰でも叩けて、返る中身が権限で変わる（サイドバーの入口なので
   * admin では絞っていない）。未所属・存在しない slug はどちらも 404。
   */
  async fetchSpaces(workspaceSlug: string): Promise<KbSpace[]> {
    const res = await apiClient.get<KbSpace[]>(KB_API.spaces(workspaceSlug));
    return toArray<KbSpace>(res.data);
  },

  /**
   * スペース配下のページツリー。**一度に全件**返る（子を個別に取る経路はまだ配線していない）。
   *
   * 実データで重くなったら遅延読み込みへ切り替える余地はあるが、先に作り込む理由が無い。
   */
  async fetchPageTree(
    workspaceSlug: string,
    spaceId: string,
    options: { archived?: boolean } = {},
  ): Promise<KbPageTree> {
    const res = await apiClient.get<KbPageTree>(KB_API.pages(workspaceSlug, spaceId), {
      // 既定は現役。別の口ではなく同じ口のスコープなので、応答の形も同じ。
      params: options.archived ? { archived: 'true' } : undefined,
    });
    // pages が欠けた応答（想定外）でも描画側が落ちないよう、配列だけは必ず用意する。
    return {
      pages: toArray(res.data?.pages),
      hasHiddenChildren: res.data?.hasHiddenChildren ?? false,
    };
  },

  /**
   * ワークスペースを作る。作った本人がそのワークスペースの admin になる。
   *
   * slug は**テナントをまたいで一意**なので、使われていれば 409 で返る
   * （この応答自体は今後見直す予定）。**失敗は例外として投げる。**
   */
  /**
   * ワークスペースを配下ごと消す。**戻せない。**
   * 会社に紐づくワークスペースはサーバーが 403 で断る（誰であっても消せない）。
   */
  async deleteWorkspace(workspaceSlug: string): Promise<void> {
    await apiClient.delete(KB_API.workspace(workspaceSlug));
  },

  async createWorkspace(input: { name: string }): Promise<KbWorkspace> {
    const res = await apiClient.post<KbWorkspace>(KB_API.workspaces, input);
    return res.data;
  },

  /**
   * スペースを作る。ワークスペースの admin だけが叩ける。
   *
   * key はワークスペース内で一意。**失敗は例外として投げる。**
   */
  async createSpace(
    workspaceSlug: string,
    input: { name: string; visibility?: 'workspace' | 'private' },
  ): Promise<KbSpace> {
    const res = await apiClient.post<KbSpace>(KB_API.spaces(workspaceSlug), input);
    return res.data;
  },

  /**
   * スペースの表示名を変える。key は URL・権限の参照に使うので変えられない。
   * 管理権限が無ければ 403、見えないスペースは 404。**失敗は例外として投げる。**
   */
  async renameSpace(workspaceSlug: string, spaceId: string, name: string): Promise<KbSpace> {
    const res = await apiClient.patch<KbSpace>(KB_API.space(workspaceSlug, spaceId), {
      name,
    });
    return res.data;
  },

  /**
   * ワークスペース全体を題名で検索する。返るのは閲覧できる現役ページだけ。
   * 見える範囲の判定はツリーと同じ規則をサーバーが持つ。**失敗は例外として投げる。**
   */
  async searchPages(workspaceSlug: string, query: string, limit?: number): Promise<KbPage[]> {
    const res = await apiClient.get<KbPage[]>(KB_API.search(workspaceSlug), {
      params: { q: query, ...(limit ? { limit } : {}) },
    });
    return toArray(res.data);
  },

  /**
   * ページを作る。parentId を省くとスペース直下、渡すとその子として作る。
   *
   * **失敗は例外として投げる**（axios がそうする）。ここで握り潰して null や false を返すと、
   * 呼び出し側は失敗を知りようがない。このリポジトリには「操作は失敗したのに成功の表示が出る」
   * 轍が既にあり、原因はどれも操作関数が失敗を投げなかったことだった。
   */
  async createPage(
    workspaceSlug: string,
    spaceId: string,
    input: { title: string; parentId?: string },
  ): Promise<KbPage> {
    const res = await apiClient.post<KbPage>(KB_API.pages(workspaceSlug, spaceId), {
      title: input.title,
      // backend は空文字を「親なし」として扱う（binding が omitempty ではないため必ず送る）。
      parentId: input.parentId ?? '',
    });
    return res.data;
  },

  /**
   * ページを子孫ごと物理削除する。アーカイブと違い戻せない。
   * **失敗は例外として投げる**（createPage と同じ理由）。
   */
  async deletePage(workspaceSlug: string, pageId: string): Promise<void> {
    await apiClient.delete(KB_API.page(workspaceSlug, pageId));
  },

  /** ページの題名を変える。**失敗は例外として投げる**（createPage と同じ理由）。 */
  async renamePage(workspaceSlug: string, pageId: string, title: string): Promise<KbPage> {
    const res = await apiClient.patch<KbPage>(KB_API.page(workspaceSlug, pageId), { title });
    return res.data;
  },

  /**
   * ページを（子孫ごと）動かす。**失敗は例外として投げる。**
   *
   * parentId を空にするとスペース直下へ戻す。位置は隣のページの ID で表す
   * （並び順のキーは持っていない。応答に入っていないため）。
   */
  async movePage(
    workspaceSlug: string,
    pageId: string,
    input: { parentId: string; beforePageId?: string; afterPageId?: string },
  ): Promise<KbPage> {
    const res = await apiClient.post<KbPage>(
      `${KB_API.page(workspaceSlug, pageId)}/move`,
      input,
    );
    return res.data;
  },

  /**
   * ページを（子孫ごと）アーカイブする。冪等。**失敗は例外として投げる。**
   */
  async archivePage(workspaceSlug: string, pageId: string): Promise<void> {
    await apiClient.post(`${KB_API.page(workspaceSlug, pageId)}/archive`);
  },

  /**
   * アーカイブしたページを（同時にアーカイブされた子孫ごと）現役へ戻す。
   *
   * 親がまだアーカイブ中なら backend が断る（子だけを戻すと、ツリーに現れない
   * 迷子ページができるため）。**失敗は例外として投げる。**
   */
  async unarchivePage(workspaceSlug: string, pageId: string): Promise<KbPage> {
    const res = await apiClient.post<KbPage>(
      `${KB_API.page(workspaceSlug, pageId)}/unarchive`,
    );
    return res.data;
  },

  /**
   * ページ 1 枚をメタ情報と本文込みで取得する。
   *
   * 閲覧できないページと存在しないページはどちらも 404。祖先に例外が張られていれば
   * 直リンクでも開けない（継承する）。
   */
  async fetchPage(workspaceSlug: string, pageId: string): Promise<KbPageDoc> {
    const res = await apiClient.get<KbPageDoc>(KB_API.page(workspaceSlug, pageId));
    return res.data;
  },

  /**
   * ページ ID だけでページと所属ワークスペースを解決する（/kb/{pageId} の入口）。
   *
   * 閲覧できないページと存在しないページはどちらも 404（実在を読ませない）。
   * 応答の workspaceSlug を以降の呼び出し（木・保存）に使う。
   */
  async resolvePage(pageId: string): Promise<KbResolvedPage> {
    const res = await apiClient.get<KbResolvedPage>(KB_API.resolvePage(pageId));
    return res.data;
  },

  /**
   * ページ本文（ProseMirror doc）を丸ごと置き換える。編集権限が要る。
   * 保存されるのは行スキーマから組み立て直した正規形で、応答はその正規形を返す。
   */
  /**
   * そのページ自身に張られた既定の役割を返す。
   *
   * **「このページを見られる人の一覧」ではない。** 上の段（ワークスペース / スペース /
   * 祖先のページ）から届いている相手は含まれず、空でも「誰も見られない」の意味にならない。
   * 画面はそれが分かる見せ方をすること。
   */
  async listPageGrants(workspaceSlug: string, pageId: string): Promise<KbPageGrant[]> {
    const res = await apiClient.get<KbPageGrant[]>(KB_API.pageGrants(workspaceSlug, pageId));
    return toArray<KbPageGrant>(res.data);
  },

  /**
   * 権限を張れる相手を表示名つきで返す（相手選び用）。
   *
   * name は空文字で返り得る（名前を引けなかった相手）。行は落とさないこと。
   */
  async listGrantablePrincipals(
    workspaceSlug: string,
    pageId: string,
  ): Promise<KbGrantablePrincipal[]> {
    const res = await apiClient.get<KbGrantablePrincipal[]>(
      KB_API.pagePrincipals(workspaceSlug, pageId),
    );
    return toArray<KbGrantablePrincipal>(res.data);
  },

  /**
   * ページでの既定の役割を主体に与える（同じ主体には 1 行だけなので上書きになる）。
   *
   * **これで誰かを弱めることはできない。** 既定は 3 段から届いて最も強いものが実効に
   * なるので、上位で editor を得ている相手に viewer を張っても editor のまま。
   */
  async grantPageRole(
    workspaceSlug: string,
    pageId: string,
    principalId: string,
    role: KbGrantRole,
  ): Promise<KbPageGrant> {
    const res = await apiClient.put<KbPageGrant>(
      KB_API.pageGrant(workspaceSlug, pageId, principalId),
      { role },
    );
    return res.data;
  },

  /** ページでの既定の役割を剥がす（冪等）。上の段から届いている分は残る。 */
  async revokePageRole(workspaceSlug: string, pageId: string, principalId: string): Promise<void> {
    await apiClient.delete(KB_API.pageGrant(workspaceSlug, pageId, principalId));
  },

  async replaceContent(
    workspaceSlug: string,
    pageId: string,
    doc: unknown,
  ): Promise<{ doc: unknown; builtAt: string }> {
    const res = await apiClient.put<{ doc: unknown; builtAt: string }>(
      KB_API.pageContent(workspaceSlug, pageId),
      { doc },
    );
    return res.data;
  },
};

export default KbRepository;
