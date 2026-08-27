import apiClient from '@/shared/api/axios';
import { toArray } from '@/shared/lib/toArray';
import { KNOWLEDGE_BASE } from '@/shared/config/apiRoutes';
import type { KbPage, KbPageDoc, KbPageTree, KbSpace, KbWorkspace } from '../model/types';

/**
 * ナレッジ基盤 API（/api/v2/kb/…）の薄いラッパ。
 *
 * 認可はすべて backend が持つ。ここでフィルタを掛けないこと。返ってくる一覧は
 * **既に「その人に見えるものだけ」**になっていて、見えないものは応答に存在しない。
 * フロントで絞り込みを重ねると、同じ判断が 2 箇所に分かれて必ずずれる。
 *
 * 「無い」と「見えない」はどちらも 404 で返る（撃ち分けると ID の総当たりで実在が分かるため）。
 * したがって 404 を「あなたには見えません」と表示してはいけない。存在しないかもしれない。
 */
const KnowledgeBaseRepository = {
  /** 自分が所属しているワークスペースの一覧。所属が無ければ空配列。 */
  async fetchWorkspaces(): Promise<KbWorkspace[]> {
    const res = await apiClient.get<KbWorkspace[]>(KNOWLEDGE_BASE.workspaces);
    return toArray<KbWorkspace>(res.data);
  },

  /**
   * ワークスペース配下の、自分に見えるスペースの一覧。
   *
   * ワークスペースのメンバーなら誰でも叩けて、返る中身が権限で変わる（サイドバーの入口なので
   * admin では絞っていない）。未所属・存在しない slug はどちらも 404。
   */
  async fetchSpaces(workspaceSlug: string): Promise<KbSpace[]> {
    const res = await apiClient.get<KbSpace[]>(KNOWLEDGE_BASE.spaces(workspaceSlug));
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
    const res = await apiClient.get<KbPageTree>(KNOWLEDGE_BASE.pages(workspaceSlug, spaceId), {
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
    const res = await apiClient.post<KbPage>(KNOWLEDGE_BASE.pages(workspaceSlug, spaceId), {
      title: input.title,
      // backend は空文字を「親なし」として扱う（binding が omitempty ではないため必ず送る）。
      parentId: input.parentId ?? '',
    });
    return res.data;
  },

  /** ページの題名を変える。**失敗は例外として投げる**（createPage と同じ理由）。 */
  async renamePage(workspaceSlug: string, pageId: string, title: string): Promise<KbPage> {
    const res = await apiClient.patch<KbPage>(KNOWLEDGE_BASE.page(workspaceSlug, pageId), { title });
    return res.data;
  },

  /**
   * ページを（子孫ごと）アーカイブする。冪等。**失敗は例外として投げる。**
   */
  async archivePage(workspaceSlug: string, pageId: string): Promise<void> {
    await apiClient.post(`${KNOWLEDGE_BASE.page(workspaceSlug, pageId)}/archive`);
  },

  /**
   * アーカイブしたページを（同時にアーカイブされた子孫ごと）現役へ戻す。
   *
   * 親がまだアーカイブ中なら backend が断る（子だけを戻すと、ツリーに現れない
   * 迷子ページができるため）。**失敗は例外として投げる。**
   */
  async unarchivePage(workspaceSlug: string, pageId: string): Promise<KbPage> {
    const res = await apiClient.post<KbPage>(
      `${KNOWLEDGE_BASE.page(workspaceSlug, pageId)}/unarchive`,
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
    const res = await apiClient.get<KbPageDoc>(KNOWLEDGE_BASE.page(workspaceSlug, pageId));
    return res.data;
  },
};

export default KnowledgeBaseRepository;
