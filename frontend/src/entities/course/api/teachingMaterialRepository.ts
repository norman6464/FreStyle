import api from '@/shared/api/axios';
import { TEACHING_MATERIALS } from '@/shared/config/apiRoutes';
import type { RichDocContent } from '@/shared/ui/RichTextEditor';
import type { TeachingMaterial } from '../model/types';

/**
 * 教材 API ラッパ（個別 CRUD）。
 *
 * 一覧取得はコース配下なので CourseRepository.listMaterials を使う。
 * actor の role / company は backend 側で context から取り出して自動フィルタするため、
 * フロントは workspaceId を渡さない（IDOR 対策）。
 */
export interface TeachingMaterialCreatePayload {
  /** 所属コース ID（必須）。 */
  courseId: number;
  title: string;
  orderInCourse: number;
  isPublished: boolean;
}

export interface TeachingMaterialUpdatePayload {
  title: string;
  orderInCourse: number;
  isPublished: boolean;
}

/** リッチ本文の保存入力。expectedRevision の不一致はサーバが 409 を返す（楽観ロック）。 */
export interface TeachingMaterialDocPayload {
  doc: RichDocContent;
  expectedRevision: number;
}

const TeachingMaterialRepository = {
  async get(id: number): Promise<TeachingMaterial> {
    const res = await api.get<TeachingMaterial>(TEACHING_MATERIALS.byId(id));
    return res.data;
  },

  async create(payload: TeachingMaterialCreatePayload): Promise<TeachingMaterial> {
    const res = await api.post<TeachingMaterial>(TEACHING_MATERIALS.create, payload);
    return res.data;
  },

  async update(id: number, payload: TeachingMaterialUpdatePayload): Promise<TeachingMaterial> {
    const res = await api.put<TeachingMaterial>(TEACHING_MATERIALS.byId(id), payload);
    return res.data;
  },

  async updateDoc(id: number, payload: TeachingMaterialDocPayload): Promise<TeachingMaterial> {
    const res = await api.put<TeachingMaterial>(TEACHING_MATERIALS.doc(id), payload);
    return res.data;
  },

  async remove(id: number): Promise<void> {
    await api.delete(TEACHING_MATERIALS.byId(id));
  },
};

export default TeachingMaterialRepository;
