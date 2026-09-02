import api from '@/shared/api/axios';
import { toArray } from '@/shared/lib/toArray';
import { COURSES } from '@/shared/config/apiRoutes';
import type {
  Course,
  CourseWithProgress,
  MaterialGrant,
  MaterialGrantRole,
  MaterialPrincipal,
  TeachingMaterial,
  UserChapterView,
} from '../model/types';

/**
 * コース API ラッパ。
 *
 * actor の role / company は backend 側で context から取り出して自動フィルタするため、
 * フロントは workspaceId を渡さない（IDOR 対策）。
 */
export interface CoursePayload {
  title: string;
  description: string;
  /** 学習領域カテゴリ（空 = 未分類） */
  category: string;
  /** 主に扱う言語・技術（空 = 言語が主題でない） */
  language: string;
  sortOrder: number;
  isPublished: boolean;
}

const CourseRepository = {
  async list(): Promise<CourseWithProgress[]> {
    const res = await api.get<CourseWithProgress[]>(COURSES.list);
    return toArray<CourseWithProgress>(res.data);
  },

  async get(id: number): Promise<Course> {
    const res = await api.get<Course>(COURSES.byId(id));
    return res.data;
  },

  async listMaterials(courseId: number): Promise<TeachingMaterial[]> {
    const res = await api.get<TeachingMaterial[]>(COURSES.materials(courseId));
    return toArray<TeachingMaterial>(res.data);
  },

  /** コース内で最後に閲覧した章の閲覧記録を返す。履歴なし（204）のときは null。 */
  async lastViewed(courseId: number): Promise<UserChapterView | null> {
    const res = await api.get<UserChapterView | undefined>(COURSES.lastViewed(courseId));
    return res.status === 204 || !res.data ? null : res.data;
  },

  async create(payload: CoursePayload): Promise<Course> {
    const res = await api.post<Course>(COURSES.list, payload);
    return res.data;
  },

  async update(id: number, payload: CoursePayload): Promise<Course> {
    const res = await api.put<Course>(COURSES.byId(id), payload);
    return res.data;
  },

  async remove(id: number): Promise<void> {
    await api.delete(COURSES.byId(id));
  },

  /**
   * そのコース自身に張られた付与を返す。
   *
   * **「このコースを編集できる人の一覧」ではない。** ワークスペースの admin は含まれず、
   * 空でも「誰も編集できない」の意味にならない。画面はそれが分かる見せ方をすること。
   */
  async listGrants(courseId: number | string): Promise<MaterialGrant[]> {
    const res = await api.get<MaterialGrant[]>(COURSES.grants(courseId));
    return toArray<MaterialGrant>(res.data);
  },

  /** 権限を張れる相手を表示名つきで返す（相手選び用）。name は空文字で返り得る。 */
  async listGrantablePrincipals(courseId: number | string): Promise<MaterialPrincipal[]> {
    const res = await api.get<MaterialPrincipal[]>(COURSES.principals(courseId));
    return toArray<MaterialPrincipal>(res.data);
  },

  /**
   * コースでの既定の役割を与える（同じ主体には 1 行だけなので上書き）。配下の章にも効く。
   *
   * **これで誰かを弱めることはできない。** 合成は最も強いものを採るので、上位で得ている
   * 役割は下がらない。
   */
  async grantRole(courseId: number | string, principalId: string, role: MaterialGrantRole): Promise<MaterialGrant> {
    const res = await api.put<MaterialGrant>(COURSES.grant(courseId, principalId), { role });
    return res.data;
  },

  /** コースでの既定の役割を剥がす（冪等）。 */
  async revokeRole(courseId: number | string, principalId: string): Promise<void> {
    await api.delete(COURSES.grant(courseId, principalId));
  },
};

export default CourseRepository;
